package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"errors"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"

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
		return nil, errors.New("verify code failed")
	}
	SetEmailResp, err := l.svcCtx.AccountCenterClient.SetEmail(l.ctx, &accountcenterclient.SetEmailReq{
		Email:  req.Email,
		UserId: ctxData.GetUserIdFromCtx(l.ctx),
	})
	if err != nil {
		l.restoreCode(req.Email, req.VerifyCode)
		return nil, err
	}

	return &types.SetEmailResp{
		Flag: SetEmailResp.Flag,
	}, nil
}

// restoreCode puts the consumed code back when the follow-up step failed, so a
// retry with the same code still works.
func (l *SetEmailLogic) restoreCode(email, verifyCode string) {
	if err := code.RestoreEmailCode(globalKey.SetEmail, email, verifyCode); err != nil {
		l.Errorf("restore set-email code for %s failed: %v", email, err)
	}
}
