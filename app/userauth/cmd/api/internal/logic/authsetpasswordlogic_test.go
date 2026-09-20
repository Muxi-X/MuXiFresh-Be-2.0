package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/common/globalKey"
)

func stubSetPasswordSeams(t *testing.T, verify bool) *[]restoreCall {
	t.Helper()

	previousVerify := verifySetPasswordCode
	previousRestore := restoreSetPasswordCode
	previousSign := signAuthSetPasswordToken

	restores := &[]restoreCall{}
	verifySetPasswordCode = func(prefix, key, value string) bool { return verify }
	restoreSetPasswordCode = func(prefix, key, value string) error {
		*restores = append(*restores, restoreCall{prefix: prefix, key: key, value: value})
		return nil
	}
	// 默认签发成功，个别用例再覆盖。
	signAuthSetPasswordToken = func(secretKey string, iat, seconds int64, email string) (string, error) {
		return "signed-token", nil
	}

	t.Cleanup(func() {
		verifySetPasswordCode = previousVerify
		restoreSetPasswordCode = previousRestore
		signAuthSetPasswordToken = previousSign
	})

	return restores
}

type restoreCall struct {
	prefix string
	key    string
	value  string
}

func newAuthSetPasswordLogic() *AuthSetPasswordLogic {
	cfg := config.Config{}
	cfg.JwtAuthChPass.AccessSecret = "chpass-secret"
	cfg.JwtAuthChPass.AccessExpire = 3600

	return NewAuthSetPasswordLogic(context.Background(), &svc.ServiceContext{Config: cfg})
}

func TestAuthSetPasswordRestoresCodeWhenSigningFails(t *testing.T) {
	restores := stubSetPasswordSeams(t, true)
	signAuthSetPasswordToken = func(secretKey string, iat, seconds int64, email string) (string, error) {
		return "", errors.New("sign failed")
	}

	l := newAuthSetPasswordLogic()
	if _, err := l.AuthSetPassword(&types.AuthSetPasswordReq{Email: "user@example.com", VerifyCode: "ABCDEF"}); err == nil {
		t.Fatal("expected signing failure to surface as an error")
	}

	if len(*restores) != 1 {
		t.Fatalf("expected the consumed code to be restored once, got %d", len(*restores))
	}
	got := (*restores)[0]
	if got.prefix != globalKey.SetPassword || got.key != "user@example.com" || got.value != "ABCDEF" {
		t.Fatalf("unexpected restore call: %+v", got)
	}
}

func TestAuthSetPasswordDoesNotRestoreOnSuccess(t *testing.T) {
	restores := stubSetPasswordSeams(t, true)

	l := newAuthSetPasswordLogic()
	resp, err := l.AuthSetPassword(&types.AuthSetPasswordReq{Email: "user@example.com", VerifyCode: "ABCDEF"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AuthSetPasswordToken == "" {
		t.Fatal("expected a non-empty auth set password token")
	}
	if len(*restores) != 0 {
		t.Fatalf("a successful flow must not restore the code, got %d calls", len(*restores))
	}
}

func TestAuthSetPasswordRejectsWrongCodeWithoutRestore(t *testing.T) {
	restores := stubSetPasswordSeams(t, false)

	l := newAuthSetPasswordLogic()
	if _, err := l.AuthSetPassword(&types.AuthSetPasswordReq{Email: "user@example.com", VerifyCode: "WRONG1"}); err == nil {
		t.Fatal("expected an invalid code to be rejected")
	}
	if len(*restores) != 0 {
		t.Fatalf("a rejected code must not trigger restore, got %d calls", len(*restores))
	}
}
