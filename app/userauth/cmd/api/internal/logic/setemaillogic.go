package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"MuXiFresh-Be-2.0/common/xerr"
	"context"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetEmailLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetEmailLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetEmailLogic {
	return &SetEmailLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetEmailLogic) SetEmail(req *types.SetEmailReq) (resp *types.SetEmailResp, err error) {

	if ok := code.VerifyEmailCode(globalKey.SetEmail, req.Email, req.VerifyCode); !ok {
		return nil, xerr.ErrEmailVerificationFailed
	}
	SetEmailResp, err := l.svcCtx.AccountCenterClient.SetEmail(l.ctx, &accountcenterclient.SetEmailReq{
		Email:  req.Email,
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		return nil, err
	}

	return &types.SetEmailResp{
		Flag: SetEmailResp.Flag,
	}, nil
}
