package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"context"
	"fmt"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type VerifyCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewVerifyCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *VerifyCaptchaLogic {
	return &VerifyCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *VerifyCaptchaLogic) VerifyCaptcha(req *types.VerifyCaptchaReq) (resp *types.VerifyCaptchaResp, err error) {
	match := code.NewCaptcha().VerifyCaptcha(req.ImageID, req.VerifyCode)
	if !match {
		return nil, fmt.Errorf("do not match")
	}
	return &types.VerifyCaptchaResp{
		Flag: true,
	}, nil
}
