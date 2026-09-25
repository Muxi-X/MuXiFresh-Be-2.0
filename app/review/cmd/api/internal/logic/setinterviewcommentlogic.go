package logic

import (
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

	if err := validateInterviewComment(req.Comment); err != nil {
		return nil, err
	}

	ret, err := l.svcCtx.EntryFormModel.SetInterviewComment(l.ctx, req.FormID, req.Comment)
	if err != nil {
		return nil, err
	}
	if ret.MatchedCount == 0 {
		return nil, errors.New("entry form not found")
	}

	return &types.SetInterviewCommentResp{Flag: true}, nil
}

// validateInterviewComment 校验面评正文字符数，空串合法（表示清空）。
func validateInterviewComment(comment string) error {
	if len([]rune(comment)) > maxInterviewCommentLen {
		return errors.New("interview comment too long")
	}
	return nil
}
