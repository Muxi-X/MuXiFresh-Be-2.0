package logic

import (
	"context"

	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type LoginLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewLoginLogic(ctx context.Context, svcCtx *svc.ServiceContext) *LoginLogic {
	return &LoginLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *LoginLogic) Login(in *pb.LoginVerifyReq) (*pb.LoginVerifyResp, error) {
	userAuth, err := l.svcCtx.UserAuthClient.FindOneByEmailAndPassword(l.ctx, in.Email, in.Password)
	if err != nil {
		return nil, err
	}
	return &pb.LoginVerifyResp{
		ID: userAuth.UserInfoID.String()[10:34],
	}, nil
}
