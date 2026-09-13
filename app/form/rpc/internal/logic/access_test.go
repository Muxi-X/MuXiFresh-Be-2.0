package logic

import (
	"context"
	"testing"

	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/form/rpc/pb"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc/metadata"
)

type fakeEntryFormModel struct {
	model.EntryFormModel
	findOneFn func(ctx context.Context, id string) (*model.EntryForm, error)
}

func (f *fakeEntryFormModel) FindOne(ctx context.Context, id string) (*model.EntryForm, error) {
	return f.findOneFn(ctx, id)
}

type fakeUserInfoModel struct {
	usermodel.UserInfoModel
	findOneFn func(ctx context.Context, id string) (*usermodel.UserInfo, error)
}

func (f *fakeUserInfoModel) FindOne(ctx context.Context, id string) (*usermodel.UserInfo, error) {
	return f.findOneFn(ctx, id)
}

func ctxWithCaller(uid string) context.Context {
	return metadata.NewIncomingContext(context.Background(),
		metadata.Pairs(ctxData.CallerIDKey, uid))
}

func TestCallerIDFromCtx(t *testing.T) {
	uid := primitive.NewObjectID().Hex()
	if got, err := callerIDFromCtx(ctxWithCaller(uid)); err != nil || got != uid {
		t.Fatalf("valid caller should parse, got %q err %v", got, err)
	}
	if _, err := callerIDFromCtx(context.Background()); err == nil {
		t.Fatal("missing metadata should be rejected")
	}
	if _, err := callerIDFromCtx(ctxWithCaller("not-an-objectid")); err == nil {
		t.Fatal("non-objectid caller should be rejected")
	}
}

func TestCheckEntryFormReadAccess(t *testing.T) {
	owner := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	form := &model.EntryForm{ID: formID, UserId: owner}

	newSvc := func(userType string) *svc.ServiceContext {
		return &svc.ServiceContext{
			FormClient: &fakeEntryFormModel{
				findOneFn: func(ctx context.Context, id string) (*model.EntryForm, error) {
					if id == formID.Hex() {
						return form, nil
					}
					return nil, mon.ErrNotFound
				},
			},
			UserInfoModel: &fakeUserInfoModel{
				findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
					return &usermodel.UserInfo{ID: primitive.NewObjectID(), UserType: userType}, nil
				},
			},
		}
	}

	ctx := context.Background()
	if err := checkEntryFormReadAccess(ctx, newSvc(globalKey.Freshman), owner.Hex(), formID.Hex()); err != nil {
		t.Fatalf("owner should read own form, got %v", err)
	}
	if err := checkEntryFormReadAccess(ctx, newSvc(globalKey.Admin), primitive.NewObjectID().Hex(), formID.Hex()); err != nil {
		t.Fatalf("admin should read others form, got %v", err)
	}
	if err := checkEntryFormReadAccess(ctx, newSvc(globalKey.SuperAdmin), primitive.NewObjectID().Hex(), formID.Hex()); err != nil {
		t.Fatalf("super admin should read others form, got %v", err)
	}
	errOther := checkEntryFormReadAccess(ctx, newSvc(globalKey.Freshman), primitive.NewObjectID().Hex(), formID.Hex())
	if errOther == nil || errOther.Error() != "无权查看该报名表" {
		t.Fatalf("freshman should be rejected with normalized error, got %v", errOther)
	}
	errMissing := checkEntryFormReadAccess(ctx, newSvc(globalKey.Admin), owner.Hex(), primitive.NewObjectID().Hex())
	if errMissing == nil || errMissing.Error() != "无权查看该报名表" {
		t.Fatalf("missing form should reuse same error, got %v", errMissing)
	}
	if errOther.Error() != errMissing.Error() {
		t.Fatalf("unauthorized and missing should be indistinguishable: %q vs %q", errOther, errMissing)
	}
}

func TestCheckEntryFormWriteAccess(t *testing.T) {
	owner := primitive.NewObjectID()
	formID := primitive.NewObjectID()

	svcCtx := &svc.ServiceContext{
		FormClient: &fakeEntryFormModel{
			findOneFn: func(ctx context.Context, id string) (*model.EntryForm, error) {
				if id == formID.Hex() {
					return &model.EntryForm{ID: formID, UserId: owner}, nil
				}
				return nil, mon.ErrNotFound
			},
		},
	}
	ctx := context.Background()
	if err := checkEntryFormWriteAccess(ctx, svcCtx, owner.Hex(), formID.Hex()); err != nil {
		t.Fatalf("owner should write own form, got %v", err)
	}
	// 管理员也不能改他人报名表（审阅无需改表）
	errNonOwner := checkEntryFormWriteAccess(ctx, svcCtx, primitive.NewObjectID().Hex(), formID.Hex())
	if errNonOwner == nil || errNonOwner.Error() != "无权修改该报名表" {
		t.Fatalf("non-owner should be rejected with normalized error, got %v", errNonOwner)
	}
	errMissing := checkEntryFormWriteAccess(ctx, svcCtx, owner.Hex(), primitive.NewObjectID().Hex())
	if errMissing == nil || errMissing.Error() != "无权修改该报名表" {
		t.Fatalf("missing form should reuse same error, got %v", errMissing)
	}
	if errNonOwner.Error() != errMissing.Error() {
		t.Fatalf("unauthorized and missing should be indistinguishable: %q vs %q", errNonOwner, errMissing)
	}
}

func TestCallerIDFromCtx_MissingAndEmpty(t *testing.T) {
	if _, err := callerIDFromCtx(context.Background()); err == nil {
		t.Fatal("missing metadata should be rejected")
	}
	empty := metadata.NewIncomingContext(context.Background(), metadata.Pairs(ctxData.CallerIDKey, ""))
	if _, err := callerIDFromCtx(empty); err == nil {
		t.Fatal("empty caller should be rejected")
	}
}

// 入口级 fail-closed：CreateForm 拒绝 metadata 中的 callerID 与入参 UserId 不一致（防代他人建表）
func TestCreateForm_RejectsMismatchedCaller(t *testing.T) {
	caller := primitive.NewObjectID()
	other := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{FormClient: &fakeEntryFormModel{}}
	l := NewCreateFormLogic(ctxWithCaller(caller.Hex()), svcCtx)

	_, err := l.CreateForm(&pb.CreateReq{UserId: other.Hex()})
	if err == nil {
		t.Fatal("create form for other user should be rejected")
	}
}

// 入口级 fail-closed：CheckForm 缺 metadata 直接拒绝
func TestCheckForm_RejectsMissingMetadata(t *testing.T) {
	svcCtx := &svc.ServiceContext{FormClient: &fakeEntryFormModel{}}
	l := NewCheckFormLogic(context.Background(), svcCtx)

	_, err := l.CheckForm(&pb.CheckReq{EntryFormID: primitive.NewObjectID().Hex()})
	if err == nil {
		t.Fatal("check form without caller metadata should be rejected")
	}
}

// 入口级校验：UpdateForm 拒绝非本人修改
func TestUpdateForm_RejectsNonOwner(t *testing.T) {
	owner := primitive.NewObjectID()
	attacker := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeEntryFormModel{
			findOneFn: func(ctx context.Context, id string) (*model.EntryForm, error) {
				return &model.EntryForm{ID: formID, UserId: owner}, nil
			},
		},
	}
	l := NewUpdateFormLogic(ctxWithCaller(attacker.Hex()), svcCtx)

	_, err := l.UpdateForm(&pb.CreateReq{FormId: formID.Hex(), UserId: owner.Hex()})
	if err == nil {
		t.Fatal("update others form should be rejected")
	}
}
