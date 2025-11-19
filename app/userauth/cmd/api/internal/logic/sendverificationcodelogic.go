package logic

import (
	"context"
	"encoding/json"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"

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
	body, _ := json.Marshal(req)
	if err = l.svcCtx.KqPusher.Push(string(body)); err != nil {
		return nil, err
	}
	return &types.SendEmailCodeResp{Flag: true}, nil
}
