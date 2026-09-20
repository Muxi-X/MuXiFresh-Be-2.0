package logic

import (
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"context"
	"errors"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetUserInfoByEmailLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewGetUserInfoByEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetUserInfoByEmailLogic {
	return &GetUserInfoByEmailLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *GetUserInfoByEmailLogic) GetUserInfoByEmail(in *pb.GetUserInfoByEmailReq) (*pb.GetUserInfoByEmailResp, error) {
	if in.Email == "" {
		return nil, errors.New("email is empty")
	}

	userInfo, err := l.svcCtx.UserInfoModel.FindByEmail(l.ctx, in.Email)
	if err != nil {
		if errors.Is(err, model.ErrNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}

	return &pb.GetUserInfoByEmailResp{
		UserId:   userInfo.ID.Hex(),
		Avatar:   userInfo.Avatar,
		NickName: userInfo.NickName,
		Name:     userInfo.Name,
		Email:    userInfo.Email,
		UserType: userInfo.UserType,
	}, nil
}
