package logic

import (
	"context"
	"testing"

	formmodel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/rpc/pb"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc/metadata"
)

type fakeScheduleModel struct {
	model.ScheduleModel
	findOneFn func(ctx context.Context, id string) (*model.Schedule, error)
	insertFn  func(ctx context.Context, data *model.Schedule) (string, error)
}

func (f *fakeScheduleModel) FindOne(ctx context.Context, id string) (*model.Schedule, error) {
	return f.findOneFn(ctx, id)
}

func (f *fakeScheduleModel) InsertGetID(ctx context.Context, data *model.Schedule) (string, error) {
	return f.insertFn(ctx, data)
}

type fakeEntryFormModel struct {
	formmodel.EntryFormModel
	findByUserFn func(ctx context.Context, userId string) (*formmodel.EntryForm, error)
}

func (f *fakeEntryFormModel) FindOneByUserId(ctx context.Context, userId string) (*formmodel.EntryForm, error) {
	return f.findByUserFn(ctx, userId)
}

type fakeUserInfoModel struct {
	usermodel.UserInfoModel
	findOneFn func(ctx context.Context, id string) (*usermodel.UserInfo, error)
	updateFn  func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error)
}

func (f *fakeUserInfoModel) FindOne(ctx context.Context, id string) (*usermodel.UserInfo, error) {
	return f.findOneFn(ctx, id)
}

func (f *fakeUserInfoModel) Update(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
	return f.updateFn(ctx, data)
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
	empty := metadata.NewIncomingContext(context.Background(), metadata.Pairs(ctxData.CallerIDKey, ""))
	if _, err := callerIDFromCtx(empty); err == nil {
		t.Fatal("empty caller should be rejected")
	}
	if _, err := callerIDFromCtx(ctxWithCaller("not-an-objectid")); err == nil {
		t.Fatal("non-objectid caller should be rejected")
	}
}

// fail-closed：缺 metadata 直接拒绝，不触达任何 model（nil 接口会 panic）
func TestCheck_RejectsMissingMetadata(t *testing.T) {
	l := NewCheckLogic(context.Background(), &svc.ServiceContext{})
	if _, err := l.Check(&pb.CheckReq{UserId: primitive.NewObjectID().Hex()}); err == nil {
		t.Fatal("check without caller metadata should be rejected")
	}
}

// 入参 UserId 与调用者不一致时拒绝
func TestCheck_RejectsMismatchedCaller(t *testing.T) {
	caller := primitive.NewObjectID()
	other := primitive.NewObjectID()
	l := NewCheckLogic(ctxWithCaller(caller.Hex()), &svc.ServiceContext{
		ScheduleClient: &fakeScheduleModel{},
	})
	if _, err := l.Check(&pb.CheckReq{UserId: other.Hex(), ScheduleID: primitive.NewObjectID().Hex()}); err == nil {
		t.Fatal("check for other user should be rejected")
	}
}

// schedule 属主不是调用者时拒绝（即使 UserId 匹配）
func TestCheck_RejectsForeignSchedule(t *testing.T) {
	caller := primitive.NewObjectID()
	other := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		EntryFormClient: &fakeEntryFormModel{
			findByUserFn: func(ctx context.Context, userId string) (*formmodel.EntryForm, error) {
				return nil, formmodel.ErrNotFound
			},
		},
		ScheduleClient: &fakeScheduleModel{
			findOneFn: func(ctx context.Context, id string) (*model.Schedule, error) {
				return &model.Schedule{ID: sid, UserID: other}, nil
			},
		},
	}
	l := NewCheckLogic(ctxWithCaller(caller.Hex()), svcCtx)
	if _, err := l.Check(&pb.CheckReq{UserId: caller.Hex(), ScheduleID: sid.Hex()}); err == nil {
		t.Fatal("check foreign schedule should be rejected")
	}
}

// 合法路径：本人查自己的 schedule 成功
func TestCheck_OwnerSuccess(t *testing.T) {
	caller := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		EntryFormClient: &fakeEntryFormModel{
			findByUserFn: func(ctx context.Context, userId string) (*formmodel.EntryForm, error) {
				return &formmodel.EntryForm{Major: "CS", Group: "Backend"}, nil
			},
		},
		ScheduleClient: &fakeScheduleModel{
			findOneFn: func(ctx context.Context, id string) (*model.Schedule, error) {
				return &model.Schedule{ID: sid, UserID: caller, EntryFormStatus: "已提交", AdmissionStatus: "已报名"}, nil
			},
		},
		UserInfoClient: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{Name: "张三", School: "信管"}, nil
			},
		},
	}
	l := NewCheckLogic(ctxWithCaller(caller.Hex()), svcCtx)
	resp, err := l.Check(&pb.CheckReq{UserId: caller.Hex(), ScheduleID: sid.Hex()})
	if err != nil {
		t.Fatalf("owner check should succeed, got %v", err)
	}
	if resp.Name != "张三" || resp.AdmissionStatus != "已报名" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

// 未交表用户：EntryForm 缺失不应报错，应返回空表单字段
func TestCheck_NoEntryFormFallback(t *testing.T) {
	caller := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		EntryFormClient: &fakeEntryFormModel{
			findByUserFn: func(ctx context.Context, userId string) (*formmodel.EntryForm, error) {
				return nil, formmodel.ErrNotFound
			},
		},
		ScheduleClient: &fakeScheduleModel{
			findOneFn: func(ctx context.Context, id string) (*model.Schedule, error) {
				return &model.Schedule{ID: sid, UserID: caller, EntryFormStatus: "未提交", AdmissionStatus: "未报名"}, nil
			},
		},
		UserInfoClient: &fakeUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{Name: "李四"}, nil
			},
		},
	}
	l := NewCheckLogic(ctxWithCaller(caller.Hex()), svcCtx)
	resp, err := l.Check(&pb.CheckReq{UserId: caller.Hex(), ScheduleID: sid.Hex()})
	if err != nil {
		t.Fatalf("no entry form should fallback, got %v", err)
	}
	if resp.Major != "" || resp.Group != "" {
		t.Fatalf("expected empty form fields, got %+v", resp)
	}
}

// fail-closed：缺 metadata 直接拒绝（Create）
func TestCreate_RejectsMissingMetadata(t *testing.T) {
	l := NewCreateLogic(context.Background(), &svc.ServiceContext{})
	if _, err := l.Create(&pb.CreateReq{UserId: primitive.NewObjectID().Hex()}); err == nil {
		t.Fatal("create without caller metadata should be rejected")
	}
}

// Create 拒绝为他人创建进度
func TestCreate_RejectsMismatchedCaller(t *testing.T) {
	caller := primitive.NewObjectID()
	other := primitive.NewObjectID()
	l := NewCreateLogic(ctxWithCaller(caller.Hex()), &svc.ServiceContext{
		ScheduleClient: &fakeScheduleModel{},
	})
	if _, err := l.Create(&pb.CreateReq{UserId: other.Hex()}); err == nil {
		t.Fatal("create for other user should be rejected")
	}
}

// Create 合法路径：为本人创建成功
func TestCreate_OwnerSuccess(t *testing.T) {
	caller := primitive.NewObjectID()
	updated := false
	svcCtx := &svc.ServiceContext{
		ScheduleClient: &fakeScheduleModel{
			insertFn: func(ctx context.Context, data *model.Schedule) (string, error) {
				return primitive.NewObjectID().String(), nil
			},
		},
		UserInfoClient: &fakeUserInfoModel{
			updateFn: func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
				updated = true
				return &mongo.UpdateResult{MatchedCount: 1}, nil
			},
		},
	}
	l := NewCreateLogic(ctxWithCaller(caller.Hex()), svcCtx)
	if _, err := l.Create(&pb.CreateReq{UserId: caller.Hex()}); err != nil {
		t.Fatalf("owner create should succeed, got %v", err)
	}
	if !updated {
		t.Fatal("userinfo should be updated with schedule id")
	}
}
