package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetUserTypeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetUserTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetUserTypeLogic {
	return &SetUserTypeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetUserTypeLogic) SetUserType(req *types.SetUserTypeReq) (resp *types.SetUserTypeResp, err error) {

	//super_admin authorization
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, errors.New("permission denied")
	}

	setUserTypeResp, err := l.svcCtx.UserClient.SetUserType(l.ctx, &userclient.SetUserTypeReq{
		Email:    req.Email,
		UserType: req.UserType,
	})
	if err != nil {
		return nil, err
	}

	return &types.SetUserTypeResp{
		Flag: setUserTypeResp.Flag,
	}, nil
}
