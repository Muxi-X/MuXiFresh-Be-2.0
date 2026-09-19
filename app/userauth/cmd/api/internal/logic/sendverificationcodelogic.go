package logic

import (
	"context"
	"encoding/json"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/tool"

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

// pushMessage is what the email consumer receives. It carries the already
// generated code so delivery stays best effort and never decides code validity.
type pushMessage struct {
	Email    string `json:"email"`
	Type     string `json:"type"`
	RandCode string `json:"rand_code"`
}

func (l *SendVerificationCodeLogic) SendVerificationCode(req *types.SendEmailCodeReq) (resp *types.SendEmailCodeResp, err error) {
	// Generate and persist the code synchronously so that a successful response
	// guarantees the code can be verified, independent of Kafka/SMTP delivery.
	randCode := tool.RandStringBytes(6)
	if err = code.SetEmailCode(req.Type, req.Email, randCode); err != nil {
		return nil, err
	}

	body, err := json.Marshal(pushMessage{
		Email:    req.Email,
		Type:     req.Type,
		RandCode: randCode,
	})
	if err != nil {
		l.rollback(req, randCode)
		return nil, err
	}

	if err = l.svcCtx.KqPusher.Push(l.ctx, string(body)); err != nil {
		l.rollback(req, randCode)
		return nil, err
	}

	return &types.SendEmailCodeResp{Flag: true}, nil
}

// rollback removes the freshly stored code when it can no longer be delivered,
// so users are not left with a valid code they will never receive. It only
// deletes its own value, never a newer code written by a concurrent resend.
func (l *SendVerificationCodeLogic) rollback(req *types.SendEmailCodeReq, randCode string) {
	if err := code.DelEmailCodeIfMatch(req.Type, req.Email, randCode); err != nil {
		l.Errorf("rollback verification code for %s failed: %v", req.Email, err)
	}
}
