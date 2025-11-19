package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"fmt"
	"github.com/jinzhu/copier"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetUserInfoLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoLogic {
	return &GetUserInfoLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetUserInfoLogic) GetUserInfo(req *types.GetUserInfoReq) (resp *types.GetUserInfoResp, err error) {

	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &pb.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})

	if err != nil {
		return nil, err
	}

	if getUserTypeResp.UserType != globalKey.SuperAdmin && getUserTypeResp.UserType != globalKey.Admin {
		return nil, fmt.Errorf("permission denied")
	}

	userInfo, err := l.svcCtx.UserClient.GetUserInfo(l.ctx, &pb.GetUserInfoReq{
		UserId: req.UserId,
	})

	if err != nil {
		return nil, err
	}
	var UserInfoResp types.GetUserInfoResp
	copier.Copy(&UserInfoResp, userInfo)
	return &UserInfoResp, nil
}
