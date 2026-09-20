package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"

	"google.golang.org/grpc"
)

type fakeUserClient struct {
	userclient.UserClient
	getUserTypeFn        func(ctx context.Context, in *userclient.GetUserTypeReq) (*userclient.GetUserTypeResp, error)
	getUserInfoByEmailFn func(ctx context.Context, in *userclient.GetUserInfoByEmailReq) (*userclient.GetUserInfoByEmailResp, error)
}

func (f *fakeUserClient) GetUserType(ctx context.Context, in *userclient.GetUserTypeReq, _ ...grpc.CallOption) (*userclient.GetUserTypeResp, error) {
	return f.getUserTypeFn(ctx, in)
}

func (f *fakeUserClient) GetUserInfoByEmail(ctx context.Context, in *userclient.GetUserInfoByEmailReq, _ ...grpc.CallOption) (*userclient.GetUserInfoByEmailResp, error) {
	return f.getUserInfoByEmailFn(ctx, in)
}

func ctxWithUser(userId string) context.Context {
	return context.WithValue(context.Background(), ctxData.CtxKeyJwtUserID, userId)
}

func TestPreviewUserByEmail(t *testing.T) {
	t.Run("get user type error propagated", func(t *testing.T) {
		boom := errors.New("rpc down")
		cli := &fakeUserClient{
			getUserTypeFn: func(ctx context.Context, in *userclient.GetUserTypeReq) (*userclient.GetUserTypeResp, error) {
				return nil, boom
			},
			getUserInfoByEmailFn: func(ctx context.Context, in *userclient.GetUserInfoByEmailReq) (*userclient.GetUserInfoByEmailResp, error) {
				t.Fatal("preview should not be called when auth check fails")
				return nil, nil
			},
		}
		l := NewPreviewUserByEmailLogic(ctxWithUser("u1"), &svc.ServiceContext{UserClient: cli})
		_, err := l.PreviewUserByEmail(&types.PreviewUserByEmailReq{Email: "a@b.com"})
		if !errors.Is(err, boom) {
			t.Fatalf("expected auth error, got %v", err)
		}
	})

	t.Run("non super_admin denied", func(t *testing.T) {
		cli := &fakeUserClient{
			getUserTypeFn: func(ctx context.Context, in *userclient.GetUserTypeReq) (*userclient.GetUserTypeResp, error) {
				return &userclient.GetUserTypeResp{UserType: "admin"}, nil
			},
			getUserInfoByEmailFn: func(ctx context.Context, in *userclient.GetUserInfoByEmailReq) (*userclient.GetUserInfoByEmailResp, error) {
				t.Fatal("preview should not be called for non super_admin")
				return nil, nil
			},
		}
		l := NewPreviewUserByEmailLogic(ctxWithUser("u1"), &svc.ServiceContext{UserClient: cli})
		_, err := l.PreviewUserByEmail(&types.PreviewUserByEmailReq{Email: "a@b.com"})
		if err == nil || err.Error() != "permission denied" {
			t.Fatalf("expected permission denied, got %v", err)
		}
	})

	t.Run("empty email rejected", func(t *testing.T) {
		cli := &fakeUserClient{
			getUserTypeFn: func(ctx context.Context, in *userclient.GetUserTypeReq) (*userclient.GetUserTypeResp, error) {
				return &userclient.GetUserTypeResp{UserType: "super_admin"}, nil
			},
			getUserInfoByEmailFn: func(ctx context.Context, in *userclient.GetUserInfoByEmailReq) (*userclient.GetUserInfoByEmailResp, error) {
				t.Fatal("preview should not be called for empty email")
				return nil, nil
			},
		}
		l := NewPreviewUserByEmailLogic(ctxWithUser("u1"), &svc.ServiceContext{UserClient: cli})
		_, err := l.PreviewUserByEmail(&types.PreviewUserByEmailReq{Email: ""})
		if err == nil || err.Error() != "email is empty" {
			t.Fatalf("expected email is empty, got %v", err)
		}
	})

	t.Run("rpc error propagated", func(t *testing.T) {
		boom := errors.New("user not found")
		cli := &fakeUserClient{
			getUserTypeFn: func(ctx context.Context, in *userclient.GetUserTypeReq) (*userclient.GetUserTypeResp, error) {
				return &userclient.GetUserTypeResp{UserType: "super_admin"}, nil
			},
			getUserInfoByEmailFn: func(ctx context.Context, in *userclient.GetUserInfoByEmailReq) (*userclient.GetUserInfoByEmailResp, error) {
				return nil, boom
			},
		}
		l := NewPreviewUserByEmailLogic(ctxWithUser("u1"), &svc.ServiceContext{UserClient: cli})
		_, err := l.PreviewUserByEmail(&types.PreviewUserByEmailReq{Email: "a@b.com"})
		if !errors.Is(err, boom) {
			t.Fatalf("expected rpc error, got %v", err)
		}
	})

	t.Run("super_admin success maps fields", func(t *testing.T) {
		cli := &fakeUserClient{
			getUserTypeFn: func(ctx context.Context, in *userclient.GetUserTypeReq) (*userclient.GetUserTypeResp, error) {
				return &userclient.GetUserTypeResp{UserType: "super_admin"}, nil
			},
			getUserInfoByEmailFn: func(ctx context.Context, in *userclient.GetUserInfoByEmailReq) (*userclient.GetUserInfoByEmailResp, error) {
				if in.Email != "a@b.com" {
					t.Errorf("email not passed through: %q", in.Email)
				}
				return &userclient.GetUserInfoByEmailResp{
					UserId:   "abc",
					Avatar:   "https://cdn/a.png",
					NickName: "小明",
					Name:     "张三",
					Email:    "a@b.com",
					UserType: "normal",
				}, nil
			},
		}
		l := NewPreviewUserByEmailLogic(ctxWithUser("u1"), &svc.ServiceContext{UserClient: cli})
		resp, err := l.PreviewUserByEmail(&types.PreviewUserByEmailReq{Email: "a@b.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.UserId != "abc" || resp.Avatar != "https://cdn/a.png" || resp.NickName != "小明" ||
			resp.Name != "张三" || resp.Email != "a@b.com" || resp.UserType != "normal" {
			t.Errorf("fields not mapped correctly: %+v", resp)
		}
	})
}
