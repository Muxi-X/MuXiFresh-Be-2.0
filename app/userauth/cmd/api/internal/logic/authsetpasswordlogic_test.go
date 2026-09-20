package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func newAuthSetPasswordLogic(t *testing.T) *AuthSetPasswordLogic {
	t.Helper()

	server := miniredis.RunT(t)
	cfg := config.Config{
		EmailCodeExpired: 10,
		CaptchaConf:      &config.CaptchaConf{},
	}
	cfg.JwtAuthChPass.AccessSecret = "chpass-secret"
	cfg.JwtAuthChPass.AccessExpire = 3600

	ctx := &svc.ServiceContext{
		Config:      cfg,
		RedisClient: redis.MustNewRedis(redis.RedisConf{Host: server.Addr(), Type: "node"}),
	}
	code.Load(cfg, ctx)

	return NewAuthSetPasswordLogic(context.Background(), ctx)
}

func TestAuthSetPasswordRestoresCodeWhenSigningFails(t *testing.T) {
	l := newAuthSetPasswordLogic(t)

	const email = "user@example.com"
	if err := code.SetEmailCode(globalKey.SetPassword, email, "ABCDEF"); err != nil {
		t.Fatalf("set code: %v", err)
	}

	previousSign := signAuthSetPasswordToken
	signAuthSetPasswordToken = func(secretKey string, iat, seconds int64, email string) (string, error) {
		return "", errors.New("sign failed")
	}
	t.Cleanup(func() { signAuthSetPasswordToken = previousSign })

	if _, err := l.AuthSetPassword(&types.AuthSetPasswordReq{Email: email, VerifyCode: "ABCDEF"}); err == nil {
		t.Fatal("expected signing failure to surface as an error")
	}

	// 签发失败后验证码必须已恢复，用户可以拿同一个码重试。
	signAuthSetPasswordToken = previousSign
	resp, err := l.AuthSetPassword(&types.AuthSetPasswordReq{Email: email, VerifyCode: "ABCDEF"})
	if err != nil {
		t.Fatalf("retry with the same code must succeed, got %v", err)
	}
	if resp.AuthSetPasswordToken == "" {
		t.Fatal("expected a non-empty auth set password token")
	}
}
