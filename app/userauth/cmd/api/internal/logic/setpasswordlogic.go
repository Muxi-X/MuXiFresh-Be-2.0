package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/tool"
	"context"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type SetPasswordLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewSetPasswordLogic(ctx context.Context, svcCtx *svc.ServiceContext) *SetPasswordLogic {
	return &SetPasswordLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *SetPasswordLogic) SetPassword(req *types.SetPasswordReq) (resp *types.SetPasswordResp, err error) {

	setPasswordResp, err := l.svcCtx.AccountCenterClient.SetPassword(l.ctx, &accountcenterclient.SetPasswordReq{
		Email:    tool.NormalizeEmail(ctxData.GetEmailFromCtx(l.ctx)),
		Password: tool.EncryptedPasswordMD5(req.Password),
	})
	if err != nil {
		return nil, err
	}
	return &types.SetPasswordResp{Flag: setPasswordResp.Flag}, nil
}
