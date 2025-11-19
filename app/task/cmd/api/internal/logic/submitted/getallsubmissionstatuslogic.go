package submitted

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/submissionclient"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"
	"github.com/jinzhu/copier"

	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAllSubmissionStatusLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAllSubmissionStatusLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAllSubmissionStatusLogic {
	return &GetAllSubmissionStatusLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAllSubmissionStatusLogic) GetAllSubmissionStatus(req *types.GetAllSubmissionStatusReq) (resp *types.GetAllSubmissionStatusResp, err error) {
	//管理员身份认证
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.Admin && getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}
	//get
	getAllStatusResp, err := l.svcCtx.SubmissionClient.GetAllSubmissionStatus(l.ctx, &submissionclient.GetAllSubmissionStatusReq{
		AssignmentID: req.AssignmentID,
	})
	if err != nil {
		return nil, err
	}
	var completions []types.Completion
	copier.Copy(&completions, &getAllStatusResp.Completions)
	return &types.GetAllSubmissionStatusResp{
		Completions: completions,
	}, nil
}
