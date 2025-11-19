package logic

import (
	"context"

	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserTypeLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserTypeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserTypeLogic {
	return &GetUserTypeLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserTypeLogic) GetUserType(in *pb.GetUserTypeReq) (*pb.GetUserTypeResp, error) {

	userInfo, err := l.svcCtx.UserInfoModel.FindOne(l.ctx, in.UserId)
	if err != nil {
		return nil, err
	}
	return &pb.GetUserTypeResp{
		UserType: userInfo.UserType,
	}, nil
}
