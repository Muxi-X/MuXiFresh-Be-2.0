package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"
	"github.com/jinzhu/copier"

	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetAdminListLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetAdminListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminListLogic {
	return &GetAdminListLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetAdminListLogic) GetAdminList(req *types.GetAdminListReq) (resp *types.GetAdminListResp, err error) {

	//super_admin authorization
	getUserTypeResp, err := l.svcCtx.UserClient.GetUserType(l.ctx, &userclient.GetUserTypeReq{
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}
	if getUserTypeResp.UserType != globalKey.SuperAdmin {
		return nil, xerr.ErrPermissionDenied
	}

	getAdminListResp, err := l.svcCtx.UserClient.GetAdminList(l.ctx, &userclient.GetAdminListReq{
		UserType: req.UserType,
	})
	if err != nil {
		return nil, err
	}
	var list []types.One
	copier.Copy(&list, &getAdminListResp.List)
	return &types.GetAdminListResp{
		List: list,
	}, nil
}
