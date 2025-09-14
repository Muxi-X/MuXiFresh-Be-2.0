package submitted

import (
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/submissionclient"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"
	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetSubmissionInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetSubmissionInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetSubmissionInfoLogic {
	return &GetSubmissionInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetSubmissionInfoLogic) GetSubmissionInfo(req *types.GetSubmissionInfoReq) (resp *types.GetSubmissionInfoResp, err error) {
	var userId string
	if req.UserId != globalKey.Myself {
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
		userId = req.UserId
	} else {
		userId = ctxData.GetUserIdFromCtx(l.ctx)
	}

	getSubmissionInfoResp, err := l.svcCtx.SubmissionClient.GetSubmissionInfo(l.ctx, &submissionclient.GetSubmissionInfoReq{
		AssignmentID: req.AssignmentID,
		UserId:       userId,
	})
	if err != nil {
		return nil, err
	}
	var submissioninfos []types.SubmissionInfo
	copier.Copy(&submissioninfos, &getSubmissionInfoResp.SubmissionInfos)
	return &types.GetSubmissionInfoResp{
		SubmissionInfos: submissioninfos,
	}, nil
}
