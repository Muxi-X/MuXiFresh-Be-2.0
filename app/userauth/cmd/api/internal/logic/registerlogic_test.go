package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"MuXiFresh-Be-2.0/common/globalKey"

	"google.golang.org/grpc"
)

// fakeAccountCenter records calls and returns the configured error, so the
// caller-level failure/retry behaviour can be exercised without a live RPC.
type fakeAccountCenter struct {
	registerErr error
	setEmailErr error
	registered  int
	setEmail    int
}

func (f *fakeAccountCenter) Register(context.Context, *accountcenterclient.RegisterDataReq, ...grpc.CallOption) (*accountcenterclient.RegisterDataResp, error) {
	f.registered++
	if f.registerErr != nil {
		return nil, f.registerErr
	}
	return &accountcenterclient.RegisterDataResp{ID: "507f1f77bcf86cd799439011"}, nil
}

func (f *fakeAccountCenter) Login(context.Context, *accountcenterclient.LoginVerifyReq, ...grpc.CallOption) (*accountcenterclient.LoginVerifyResp, error) {
	return &accountcenterclient.LoginVerifyResp{}, nil
}

func (f *fakeAccountCenter) SetPassword(context.Context, *accountcenterclient.SetPasswordReq, ...grpc.CallOption) (*accountcenterclient.SetPasswordResp, error) {
	return &accountcenterclient.SetPasswordResp{}, nil
}

func (f *fakeAccountCenter) CcnuLogin(context.Context, *accountcenterclient.CcnuLoginReq, ...grpc.CallOption) (*accountcenterclient.CcnuLoginResp, error) {
	return &accountcenterclient.CcnuLoginResp{}, nil
}

func (f *fakeAccountCenter) SetStudentID(context.Context, *accountcenterclient.SetStudentIDReq, ...grpc.CallOption) (*accountcenterclient.SetStudentIDResp, error) {
	return &accountcenterclient.SetStudentIDResp{}, nil
}

func (f *fakeAccountCenter) SetEmail(context.Context, *accountcenterclient.SetEmailReq, ...grpc.CallOption) (*accountcenterclient.SetEmailResp, error) {
	f.setEmail++
	if f.setEmailErr != nil {
		return nil, f.setEmailErr
	}
	return &accountcenterclient.SetEmailResp{Flag: true}, nil
}

func stubRegisterSeams(t *testing.T, verify bool) *[]restoreCall {
	t.Helper()

	previousVerify := verifyRegisterCode
	previousRestore := restoreRegisterCode

	restores := &[]restoreCall{}
	verifyRegisterCode = func(prefix, key, value string) bool { return verify }
	restoreRegisterCode = func(prefix, key, value string) error {
		*restores = append(*restores, restoreCall{prefix: prefix, key: key, value: value})
		return nil
	}
	t.Cleanup(func() {
		verifyRegisterCode = previousVerify
		restoreRegisterCode = previousRestore
	})

	return restores
}

func stubSetEmailSeams(t *testing.T, verify bool) *[]restoreCall {
	t.Helper()

	previousVerify := verifySetEmailCode
	previousRestore := restoreSetEmailCode

	restores := &[]restoreCall{}
	verifySetEmailCode = func(prefix, key, value string) bool { return verify }
	restoreSetEmailCode = func(prefix, key, value string) error {
		*restores = append(*restores, restoreCall{prefix: prefix, key: key, value: value})
		return nil
	}
	t.Cleanup(func() {
		verifySetEmailCode = previousVerify
		restoreSetEmailCode = previousRestore
	})

	return restores
}

func registerServiceContext(client accountcenterclient.AccountCenterClient) *svc.ServiceContext {
	cfg := config.Config{}
	cfg.JwtAuth.AccessSecret = "register-secret"
	cfg.JwtAuth.AccessExpire = 3600

	return &svc.ServiceContext{Config: cfg, AccountCenterClient: client}
}

func TestRegisterRestoresCodeWhenRPCFailsThenRetries(t *testing.T) {
	restores := stubRegisterSeams(t, true)

	client := &fakeAccountCenter{registerErr: errors.New("register rpc failed")}
	l := NewRegisterLogic(context.Background(), registerServiceContext(client))

	req := &types.RegisterReq{Email: "user@example.com", Password: "pw", VerifyCode: "ABCDEF"}
	if _, err := l.Register(req); err == nil {
		t.Fatal("expected the RPC failure to surface as an error")
	}
	if len(*restores) != 1 || (*restores)[0] != (restoreCall{prefix: globalKey.Register, key: "user@example.com", value: "ABCDEF"}) {
		t.Fatalf("expected the code to be restored once, got %v", *restores)
	}

	// 修复后重试：同一个码应当能再次走通完整流程。这里 verify/restore 均被 stub，
	// 只验证「失败后仍能再次到达 RPC」；验证码恢复后真实可用由 code 包测试覆盖。
	client.registerErr = nil
	if _, err := l.Register(req); err != nil {
		t.Fatalf("retry with the same code must succeed, got %v", err)
	}
	if client.registered != 2 {
		t.Fatalf("expected the retry to reach the RPC, got %d calls", client.registered)
	}
}

func TestSetEmailRestoresCodeWhenRPCFailsThenRetries(t *testing.T) {
	restores := stubSetEmailSeams(t, true)

	client := &fakeAccountCenter{setEmailErr: errors.New("set email rpc failed")}
	l := NewSetEmailLogic(context.Background(), registerServiceContext(client))

	req := &types.SetEmailReq{Email: "user@example.com", VerifyCode: "ABCDEF"}
	if _, err := l.SetEmail(req); err == nil {
		t.Fatal("expected the RPC failure to surface as an error")
	}
	if len(*restores) != 1 || (*restores)[0] != (restoreCall{prefix: globalKey.SetEmail, key: "user@example.com", value: "ABCDEF"}) {
		t.Fatalf("expected the code to be restored once, got %v", *restores)
	}

	client.setEmailErr = nil
	if _, err := l.SetEmail(req); err != nil {
		t.Fatalf("retry with the same code must succeed, got %v", err)
	}
	if client.setEmail != 2 {
		t.Fatalf("expected the retry to reach the RPC, got %d calls", client.setEmail)
	}
}
