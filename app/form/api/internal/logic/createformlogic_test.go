package logic

import (
	"context"
	"testing"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	formModel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

type fakeCreateFormClient struct {
	entryformclient.EntryFormClient
	formID string
}

func (f *fakeCreateFormClient) CreateForm(ctx context.Context, in *entryformclient.CreateReq, opts ...grpc.CallOption) (*entryformclient.CreateResp, error) {
	return &entryformclient.CreateResp{FormID: f.formID}, nil
}

type fakeEntryFormModel struct {
	formModel.EntryFormModel
	deleteFn func(ctx context.Context, id string) (int64, error)
}

func (f *fakeEntryFormModel) Delete(ctx context.Context, id string) (int64, error) {
	return f.deleteFn(ctx, id)
}

type fakeScheduleModel struct {
	scheduleModel.ScheduleModel
	upsertFn   func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error)
	findByUser func(ctx context.Context, userId string) (*scheduleModel.Schedule, error)
}

func (f *fakeScheduleModel) UpsertByUserId(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
	return f.upsertFn(ctx, data)
}

func (f *fakeScheduleModel) FindOneByUserId(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
	return f.findByUser(ctx, userId)
}

type fakeUserInfoUpdateModel struct {
	usermodel.UserInfoModel
	updateFn func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error)
}

func (f *fakeUserInfoUpdateModel) Update(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
	return f.updateFn(ctx, data)
}

func createCtx(userID string) context.Context {
	return context.WithValue(context.Background(), ctxData.CtxKeyJwtUserID, userID)
}

func newCreateReq() *types.CreateReq {
	return &types.CreateReq{
		Avatar: "a", Major: "CS", Grade: "大一", Gender: "male",
		Phone: "1", Group: "Backend", Reason: "r", Knowledge: "k",
		SelfIntro: "s", ExtraQuestion: "e",
	}
}

// 第 2 步（schedule upsert）失败时应删除已插入的 entry_form 并返回原始错误
func TestCreateForm_RollbackOnScheduleFailure(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	deleted := ""
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deleted = id
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return nil, context.DeadlineExceeded
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	_, err := l.CreateForm(newCreateReq())
	if err == nil {
		t.Fatal("schedule failure should return error")
	}
	if deleted != formID.Hex() {
		t.Fatalf("entry_form should be rolled back with form id, got %q", deleted)
	}
}

// 第 3 步（userinfo 回写）失败时同样应删除 entry_form
func TestCreateForm_RollbackOnUserInfoFailure(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	deleted := ""
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deleted = id
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return &mongo.UpdateResult{}, nil
			},
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID}, nil
			},
		},
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			updateFn: func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
				return nil, mon.ErrNotFound
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	_, err := l.CreateForm(newCreateReq())
	if err == nil {
		t.Fatal("userinfo failure should return error")
	}
	if deleted != formID.Hex() {
		t.Fatalf("entry_form should be rolled back with form id, got %q", deleted)
	}
}

// userinfo 不存在（MatchedCount=0）时也应回滚
func TestCreateForm_RollbackOnMissingUserInfo(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	deleted := ""
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deleted = id
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return &mongo.UpdateResult{}, nil
			},
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID}, nil
			},
		},
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			updateFn: func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
				return &mongo.UpdateResult{MatchedCount: 0}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	_, err := l.CreateForm(newCreateReq())
	if err == nil {
		t.Fatal("missing userinfo should return error")
	}
	if deleted != formID.Hex() {
		t.Fatalf("entry_form should be rolled back when userinfo missing, got %q", deleted)
	}
}

// 成功路径不应触发删除
func TestCreateForm_SuccessNoRollback(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	deleted := false
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deleted = true
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return &mongo.UpdateResult{}, nil
			},
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID}, nil
			},
		},
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			updateFn: func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
				return &mongo.UpdateResult{MatchedCount: 1}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); err != nil {
		t.Fatalf("create should succeed, got %v", err)
	}
	if deleted {
		t.Fatal("success path should not delete entry_form")
	}
}
