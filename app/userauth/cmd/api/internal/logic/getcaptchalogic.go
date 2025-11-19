package logic

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"context"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"

	"github.com/zeromicro/go-zero/core/logx"
)

type GetCaptchaLogic struct {
	logx.Logger
	ctx    context.Context
	svcCtx *svc.ServiceContext
}

func NewGetCaptchaLogic(ctx context.Context, svcCtx *svc.ServiceContext) *GetCaptchaLogic {
	return &GetCaptchaLogic{
		Logger: logx.WithContext(ctx),
		ctx:    ctx,
		svcCtx: svcCtx,
	}
}

func (l *GetCaptchaLogic) GetCaptcha() (resp *types.GetCaptchaResp, err error) {
	id, base64, err := code.NewCaptcha().GenerateCaptcha()
	if err != nil {
		return nil, err
	}
	return &types.GetCaptchaResp{
		ImageBase64: base64,
		ImageID:     id,
	}, nil
}
