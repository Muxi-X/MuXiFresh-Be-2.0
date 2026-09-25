package logic

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
)

// maxInterviewCommentLen 是面评正文的字符数（rune）上限，防止单文档过大。
const maxInterviewCommentLen = 20000

type SetInterviewCommentLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetInterviewCommentLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetInterviewCommentLogic {
	return &SetInterviewCommentLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetInterviewCommentLogic) SetInterviewComment(req *types.SetInterviewCommentReq) (resp *types.SetInterviewCommentResp, err error) {
	//管理员认证
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}

	if req.Rev < 0 {
		return nil, errors.New("invalid rev")
	}
	if err := validateInterviewComment(req.Comment); err != nil {
		return nil, err
	}

	// 乐观锁写入：只有当前版本等于 req.Rev 才成功，避免基于旧版本的覆盖
	ret, err := l.svcCtx.EntryFormModel.SetInterviewComment(l.ctx, req.FormID, req.Comment, req.Rev)
	if err != nil {
		return nil, err
	}
	if ret.MatchedCount == 0 {
		// 未命中可能是版本冲突，也可能是报名表不存在，读一次加以区分
		_, findErr := l.svcCtx.EntryFormModel.FindOne(l.ctx, req.FormID)
		return nil, commentWriteError(findErr)
	}

	return &types.SetInterviewCommentResp{Flag: true, Rev: req.Rev + 1}, nil
}

// commentWriteError 把 CAS 未命中后的读回结果映射为对外错误：
// 文档不存在 -> entry form not found；读回本身出错 -> 原错误；其余 -> 版本冲突。
func commentWriteError(findErr error) error {
	switch {
	case errors.Is(findErr, model.ErrNotFound):
		return errors.New("entry form not found")
	case findErr != nil:
		return findErr
	default:
		return errors.New("comment has been modified, please refresh")
	}
}

// validateInterviewComment 校验面评正文字符数，空串合法（表示清空）。
func validateInterviewComment(comment string) error {
	if len([]rune(comment)) > maxInterviewCommentLen {
		return errors.New("interview comment too long")
	}
	return nil
}
