package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/svc"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/pb"
	"MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/xerr"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// registerCfg 返回带默认用户信息的配置，供 Register 测试构造 ServiceContext。
func registerCfg() config.Config {
	cfg := config.Config{}
	cfg.DefaultUserInfo.Avatar = "a"
	cfg.DefaultUserInfo.NickName = "n"
	return cfg
}

// duplicateKeyErr 构造驱动可识别的 E11000 唯一键冲突错误。
func duplicateKeyErr() error {
	return mongo.WriteException{WriteErrors: mongo.WriteErrors{
		{Code: 11000, Message: "E11000 duplicate key error"},
	}}
}

type fakeSetEmailUserInfo struct {
	model.UserInfoModel
	findByEmailFn func(ctx context.Context, email string) (*model.UserInfo, error)
	updateFn      func(ctx context.Context, data *model.UserInfo) (*mongo.UpdateResult, error)
}

func (f *fakeSetEmailUserInfo) FindByEmail(ctx context.Context, email string) (*model.UserInfo, error) {
	return f.findByEmailFn(ctx, email)
}

func (f *fakeSetEmailUserInfo) Update(ctx context.Context, data *model.UserInfo) (*mongo.UpdateResult, error) {
	return f.updateFn(ctx, data)
}

type fakeSetEmailUserAuth struct {
	model.UserAuthModel
	findByUserIDFn func(ctx context.Context, userId primitive.ObjectID) (*model.UserAuth, error)
	updateFn       func(ctx context.Context, data *model.UserAuth) (*mongo.UpdateResult, error)
}

func (f *fakeSetEmailUserAuth) FindOneByUserId(ctx context.Context, userId primitive.ObjectID) (*model.UserAuth, error) {
	return f.findByUserIDFn(ctx, userId)
}

func (f *fakeSetEmailUserAuth) UpdateByUserId(ctx context.Context, data *model.UserAuth) (*mongo.UpdateResult, error) {
	return f.updateFn(ctx, data)
}

func newSetEmailSvc(ui *fakeSetEmailUserInfo, ua *fakeSetEmailUserAuth) *svc.ServiceContext {
	return &svc.ServiceContext{UserInfoClient: ui, UserAuthClient: ua}
}

func TestSetEmail_RejectsOccupiedEmail(t *testing.T) {
	otherID := primitive.NewObjectID()
	uid := primitive.NewObjectID()
	ui := &fakeSetEmailUserInfo{
		findByEmailFn: func(context.Context, string) (*model.UserInfo, error) {
			return &model.UserInfo{ID: otherID}, nil
		},
		updateFn: func(context.Context, *model.UserInfo) (*mongo.UpdateResult, error) {
			t.Fatal("userinfo update must not run when email is occupied")
			return nil, nil
		},
	}
	ua := &fakeSetEmailUserAuth{
		findByUserIDFn: func(context.Context, primitive.ObjectID) (*model.UserAuth, error) {
			t.Fatal("userauth lookup must not run when email is occupied")
			return nil, nil
		},
		updateFn: func(context.Context, *model.UserAuth) (*mongo.UpdateResult, error) {
			t.Fatal("userauth update must not run when email is occupied")
			return nil, nil
		},
	}
	l := NewSetEmailLogic(context.Background(), newSetEmailSvc(ui, ua))
	if _, err := l.SetEmail(&pb.SetEmailReq{UserId: uid.Hex(), Email: "a@b.com"}); err == nil {
		t.Fatal("expected occupied email to be rejected")
	}
}

func TestSetEmail_UpdatesBothCollections(t *testing.T) {
	uid := primitive.NewObjectID()
	var authEmail, infoEmail string
	ui := &fakeSetEmailUserInfo{
		findByEmailFn: func(context.Context, string) (*model.UserInfo, error) {
			return nil, model.ErrNotFound
		},
		updateFn: func(_ context.Context, data *model.UserInfo) (*mongo.UpdateResult, error) {
			infoEmail = data.Email
			return &mongo.UpdateResult{}, nil
		},
	}
	ua := &fakeSetEmailUserAuth{
		findByUserIDFn: func(context.Context, primitive.ObjectID) (*model.UserAuth, error) {
			return &model.UserAuth{Email: "old@x.com"}, nil
		},
		updateFn: func(_ context.Context, data *model.UserAuth) (*mongo.UpdateResult, error) {
			authEmail = data.Email
			return &mongo.UpdateResult{}, nil
		},
	}
	l := NewSetEmailLogic(context.Background(), newSetEmailSvc(ui, ua))
	if _, err := l.SetEmail(&pb.SetEmailReq{UserId: uid.Hex(), Email: "New@QQ.COM"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if authEmail != "New@qq.com" || infoEmail != "New@qq.com" {
		t.Fatalf("both collections must get normalized email, got auth=%q info=%q", authEmail, infoEmail)
	}
}

func TestSetEmail_RollsBackUserAuthOnUserInfoFail(t *testing.T) {
	uid := primitive.NewObjectID()
	var authUpdates []string
	ui := &fakeSetEmailUserInfo{
		findByEmailFn: func(context.Context, string) (*model.UserInfo, error) {
			return nil, model.ErrNotFound
		},
		updateFn: func(context.Context, *model.UserInfo) (*mongo.UpdateResult, error) {
			return nil, errors.New("userinfo down")
		},
	}
	ua := &fakeSetEmailUserAuth{
		findByUserIDFn: func(context.Context, primitive.ObjectID) (*model.UserAuth, error) {
			return &model.UserAuth{Email: "old@x.com"}, nil
		},
		updateFn: func(_ context.Context, data *model.UserAuth) (*mongo.UpdateResult, error) {
			authUpdates = append(authUpdates, data.Email)
			return &mongo.UpdateResult{}, nil
		},
	}
	l := NewSetEmailLogic(context.Background(), newSetEmailSvc(ui, ua))
	if _, err := l.SetEmail(&pb.SetEmailReq{UserId: uid.Hex(), Email: "new@qq.com"}); err == nil {
		t.Fatal("expected userinfo failure to surface")
	}
	if len(authUpdates) != 2 || authUpdates[0] != "new@qq.com" || authUpdates[1] != "old@x.com" {
		t.Fatalf("userauth should be rolled back to old email, got %v", authUpdates)
	}
}

func TestRegister_DuplicateEmailReturnsRegistered(t *testing.T) {
	ui := &fakeRegisterUserInfoModel{
		insertFn: func(context.Context, *model.UserInfo) error {
			return duplicateKeyErr()
		},
		deleteFn: func(context.Context, string) (int64, error) { return 1, nil },
	}
	ua := &fakeRegisterUserAuthModel{
		insertFn: func(context.Context, *model.UserAuth) error {
			t.Fatal("userauth insert must not run when userinfo already duplicated")
			return nil
		},
	}
	svcCtx := &svc.ServiceContext{
		Config:         registerCfg(),
		UserInfoClient: ui,
		UserAuthClient: ua,
	}
	l := NewRegisterLogic(context.Background(), svcCtx)
	_, err := l.Register(&pb.RegisterDataReq{Email: "a@b.com", Password: "p"})
	var ce *xerr.CodeError
	if !errors.As(err, &ce) || ce.GetErrCode() != xerr.EMAIL_REGISTERED_ERROR {
		t.Fatalf("expected EMAIL_REGISTERED_ERROR, got %v", err)
	}
}

func TestRegister_UserAuthDuplicateRollsBackAndReturnsRegistered(t *testing.T) {
	var deleted bool
	ui := &fakeRegisterUserInfoModel{
		insertFn: func(_ context.Context, data *model.UserInfo) error {
			data.ID = primitive.NewObjectID()
			return nil
		},
		deleteFn: func(context.Context, string) (int64, error) {
			deleted = true
			return 1, nil
		},
	}
	ua := &fakeRegisterUserAuthModel{
		insertFn: func(context.Context, *model.UserAuth) error {
			return duplicateKeyErr()
		},
	}
	svcCtx := &svc.ServiceContext{
		Config:         registerCfg(),
		UserInfoClient: ui,
		UserAuthClient: ua,
	}
	l := NewRegisterLogic(context.Background(), svcCtx)
	_, err := l.Register(&pb.RegisterDataReq{Email: "a@b.com", Password: "p"})
	var ce *xerr.CodeError
	if !errors.As(err, &ce) || ce.GetErrCode() != xerr.EMAIL_REGISTERED_ERROR {
		t.Fatalf("expected EMAIL_REGISTERED_ERROR, got %v", err)
	}
	if !deleted {
		t.Fatal("userinfo must be rolled back on userauth duplicate key")
	}
}
