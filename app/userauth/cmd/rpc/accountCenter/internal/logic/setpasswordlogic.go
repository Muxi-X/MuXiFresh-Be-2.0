package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/model"
	"context"

	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetPasswordLogic struct {
	ctx    context.Context
	svcCtx *svc.ServiceContext
	logx.Logger
}

func NewSetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPasswordLogic {
	return &SetPasswordLogic{
		ctx:    ctx,
		svcCtx: svcCtx,
		Logger: logx.WithContext(ctx),
	}
}

func (l *SetPasswordLogic) SetPassword(in *pb.SetPasswordReq) (*pb.SetPasswordResp, error) {

	_, err := l.svcCtx.UserAuthClient.UpdateByEmail(l.ctx, &model.UserAuth{
		Email:    in.Email,
		Password: in.Password,
	})
	if err != nil {
		return nil, err
	}
	return &pb.SetPasswordResp{
		Flag: true,
	}, nil
}
