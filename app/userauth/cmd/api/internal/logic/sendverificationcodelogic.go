package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/email"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type SendVerificationCodeLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSendVerificationCodeLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SendVerificationCodeLogic {
	return &SendVerificationCodeLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SendVerificationCodeLogic) SendVerificationCode(req *types.SendEmailCodeReq) (resp *types.SendEmailCodeResp, err error) {
	existEmailResp, err := l.svcCtx.AccountCenterClient.ExistEmail(l.ctx, &accountcenterclient.ExistEmailReq{
		Email: req.Email,
	})
	if err != nil {
		return nil, err
	}

	switch req.Type {
	case globalKey.SetPassword:
		if !existEmailResp.Exist {
			return nil, xerr.ErrEmailHasNotBeenUsed
		}
	default:
		if existEmailResp.Exist {
			return nil, xerr.ErrEmailHasBeenUsed
		}
	}

	msg := &email.Msg{
		Email: req.Email,
		Type:  req.Type,
	}
	err = email.Push(msg)
	if err != nil {
		return nil, err
	}
	return &types.SendEmailCodeResp{Flag: true}, nil
}
