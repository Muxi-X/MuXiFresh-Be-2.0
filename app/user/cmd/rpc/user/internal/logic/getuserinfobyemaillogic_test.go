package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/internal/svc"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/pb"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeUserInfoModel struct {
	usermodel.UserInfoModel
	findByEmailFn func(ctx context.Context, email string) (*usermodel.UserInfo, error)
}

func (f *fakeUserInfoModel) FindByEmail(ctx context.Context, email string) (*usermodel.UserInfo, error) {
	return f.findByEmailFn(ctx, email)
}

func newLogic(findByEmailFn func(ctx context.Context, email string) (*usermodel.UserInfo, error)) *GetUserInfoByEmailLogic {
	return NewGetUserInfoByEmailLogic(context.Background(), &svc.ServiceContext{
		UserInfoModel: &fakeUserInfoModel{findByEmailFn: findByEmailFn},
	})
}

func TestGetUserInfoByEmail(t *testing.T) {
	oid := primitive.NewObjectID()

	t.Run("empty email rejected", func(t *testing.T) {
		l := newLogic(func(ctx context.Context, email string) (*usermodel.UserInfo, error) {
			t.Fatal("model should not be called for empty email")
			return nil, nil
		})
		resp, err := l.GetUserInfoByEmail(&pb.GetUserInfoByEmailReq{Email: ""})
		if err == nil {
			t.Fatal("expected error for empty email")
		}
		if resp != nil {
			t.Fatalf("expected nil resp, got %+v", resp)
		}
	})

	t.Run("not found returns friendly error", func(t *testing.T) {
		l := newLogic(func(ctx context.Context, email string) (*usermodel.UserInfo, error) {
			return nil, usermodel.ErrNotFound
		})
		_, err := l.GetUserInfoByEmail(&pb.GetUserInfoByEmailReq{Email: "nobody@example.com"})
		if err == nil || err.Error() != "user not found" {
			t.Fatalf("expected 'user not found', got %v", err)
		}
	})

	t.Run("model error propagated", func(t *testing.T) {
		boom := errors.New("db down")
		l := newLogic(func(ctx context.Context, email string) (*usermodel.UserInfo, error) {
			return nil, boom
		})
		_, err := l.GetUserInfoByEmail(&pb.GetUserInfoByEmailReq{Email: "a@b.com"})
		if !errors.Is(err, boom) {
			t.Fatalf("expected original error, got %v", err)
		}
	})

	t.Run("success maps fields with hex id", func(t *testing.T) {
		l := newLogic(func(ctx context.Context, email string) (*usermodel.UserInfo, error) {
			return &usermodel.UserInfo{
				ID:       oid,
				Avatar:   "https://cdn/a.png",
				NickName: "小明",
				Name:     "张三",
				Email:    email,
				UserType: "freshman",
			}, nil
		})
		resp, err := l.GetUserInfoByEmail(&pb.GetUserInfoByEmailReq{Email: "a@b.com"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if resp.UserId != oid.Hex() {
			t.Errorf("UserId = %q, want %q", resp.UserId, oid.Hex())
		}
		if resp.Avatar != "https://cdn/a.png" || resp.NickName != "小明" || resp.Name != "张三" ||
			resp.Email != "a@b.com" || resp.UserType != "freshman" {
			t.Errorf("fields not mapped correctly: %+v", resp)
		}
	})
}
