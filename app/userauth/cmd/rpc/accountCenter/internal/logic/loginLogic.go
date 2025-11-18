package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/xerr"
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
		if err == model.ErrNotFound {
			return nil, xerr.ErrEmailOrPasswordIsWrong.Status()
		}
		return nil, xerr.NewErrCode(xerr.DB_ERROR).Status()
	}
	return &pb.LoginVerifyResp{
		ID: userAuth.UserInfoID.String()[10:34],
	}, nil
}
