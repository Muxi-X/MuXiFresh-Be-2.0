package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"github.com/zeromicro/go-zero/core/logx"
)

type GetAdminListLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetAdminListLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetAdminListLogic {
	return &GetAdminListLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetAdminListLogic) GetAdminList(in *pb.GetAdminListReq) (*pb.GetAdminListResp, error) {

	var userInfos []*model.UserInfo

	userType := in.UserType
	var err error

	if userType == globalKey.SuperAdmin || userType == globalKey.Admin {
		userInfos, err = l.svcCtx.UserInfoModel.FindByUserType(l.ctx, in.UserType)
		if err != nil {
			return nil, err
		}
	} else {
		userInfos, err = l.svcCtx.UserInfoModel.FindByUserType(l.ctx, globalKey.Freshman)
		if err != nil {
			return nil, err
		}
		tmpInfos, err := l.svcCtx.UserInfoModel.FindByUserType(l.ctx, globalKey.Normal)
		if err != nil {
			return nil, err
		}
		userInfos = append(userInfos, tmpInfos...)
	}

	var list []*pb.One
	for _, userInfo := range userInfos {
		list = append(list, &pb.One{
			UserId:   userInfo.ID.String()[10:34],
			Nickname: userInfo.NickName,
			Avatar:   userInfo.Avatar,
			Name:     userInfo.Name,
			Email:    userInfo.Email,
		})
	}
	return &pb.GetAdminListResp{
		List: list,
	}, nil
}
