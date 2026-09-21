package logic

import (
	"context"
	"errors"
	"testing"
	"time"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	formModel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"google.golang.org/grpc"
)

var errSchedule = errors.New("schedule upsert boom")

type fakeCreateFormClient struct {
	entryformclient.EntryFormClient
	formID        string
	createErr     error
	createCalled  bool
	updateCalled  bool
	updateFormID  string
	createReqSeen *entryformclient.CreateReq
}

func (f *fakeCreateFormClient) CreateForm(ctx context.Context, in *entryformclient.CreateReq, opts ...grpc.CallOption) (*entryformclient.CreateResp, error) {
	f.createCalled = true
	f.createReqSeen = in
	if f.createErr != nil {
		return nil, f.createErr
	}
	return &entryformclient.CreateResp{FormID: f.formID}, nil
}

func (f *fakeCreateFormClient) UpdateForm(ctx context.Context, in *entryformclient.CreateReq, opts ...grpc.CallOption) (*entryformclient.CreateResp, error) {
	f.updateCalled = true
	f.updateFormID = in.FormId
	return &entryformclient.CreateResp{}, nil
}

type fakeEntryFormModel struct {
	formModel.EntryFormModel
	deleteFn      func(ctx context.Context, id string) (int64, error)
	findOneFn     func(ctx context.Context, id string) (*formModel.EntryForm, error)
	findByCycleFn func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error)
	findByUserFn  func(ctx context.Context, userId string) (*formModel.EntryForm, error)
}

func (f *fakeEntryFormModel) Delete(ctx context.Context, id string) (int64, error) {
	if f.deleteFn == nil {
		return 1, nil
	}
	return f.deleteFn(ctx, id)
}

func (f *fakeEntryFormModel) FindOne(ctx context.Context, id string) (*formModel.EntryForm, error) {
	return f.findOneFn(ctx, id)
}

func (f *fakeEntryFormModel) FindByUserIdAndCycle(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
	return f.findByCycleFn(ctx, userId, cycle)
}

func (f *fakeEntryFormModel) FindOneByUserId(ctx context.Context, userId string) (*formModel.EntryForm, error) {
	if f.findByUserFn == nil {
		return nil, formModel.ErrNotFound
	}
	return f.findByUserFn(ctx, userId)
}

type fakeScheduleModel struct {
	scheduleModel.ScheduleModel
	upsertFn   func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error)
	findByUser func(ctx context.Context, userId string) (*scheduleModel.Schedule, error)
}

func (f *fakeScheduleModel) UpsertByUserId(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
	if f.upsertFn == nil {
		return &mongo.UpdateResult{}, nil
	}
	return f.upsertFn(ctx, data)
}

func (f *fakeScheduleModel) FindOneByUserId(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
	if f.findByUser == nil {
		// 默认视为尚无进度记录（未报名用户），放行录取状态校验
		return nil, scheduleModel.ErrNotFound
	}
	return f.findByUser(ctx, userId)
}

type fakeUserInfoUpdateModel struct {
	usermodel.UserInfoModel
	updateFn func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error)
	findFn   func(ctx context.Context, id string) (*usermodel.UserInfo, error)
}

func (f *fakeUserInfoUpdateModel) Update(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
	return f.updateFn(ctx, data)
}

func (f *fakeUserInfoUpdateModel) FindOne(ctx context.Context, id string) (*usermodel.UserInfo, error) {
	if f.findFn == nil {
		// 默认返回"未关联任何报名表"，使回滚可执行
		return &usermodel.UserInfo{ID: primitive.NewObjectID()}, nil
	}
	return f.findFn(ctx, id)
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

// 本届无表：新建，并关联 schedule/userinfo
func TestCreateForm_NewCurrentCycleForm(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	var deleted bool
	fc := &fakeCreateFormClient{formID: formID.Hex()}
	svcCtx := &svc.ServiceContext{
		FormClient: fc,
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deleted = true
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID}, nil
			},
		},
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			updateFn: func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
				if data.EntryFormID != formID {
					t.Fatalf("userinfo should be linked to new form, got %s", data.EntryFormID)
				}
				return &mongo.UpdateResult{MatchedCount: 1}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); err != nil {
		t.Fatalf("new form should succeed, got %v", err)
	}
	if !fc.createCalled {
		t.Fatal("should create a new form when none exists this cycle")
	}
	if fc.updateCalled {
		t.Fatal("should not update when no current-cycle form exists")
	}
	if deleted {
		t.Fatal("successful create must not roll back")
	}
}

// 本届已有表：覆盖更新（幂等），不新建
func TestCreateForm_CurrentCycleExistingUpdates(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	fc := &fakeCreateFormClient{}
	svcCtx := &svc.ServiceContext{
		FormClient: fc,
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return &formModel.EntryForm{ID: formID, UserId: userID, Cycle: cycle}, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
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
		t.Fatalf("duplicate submit should be idempotent, got %v", err)
	}
	if fc.createCalled {
		t.Fatal("existing current-cycle form should not be re-created")
	}
	if !fc.updateCalled || fc.updateFormID != formID.Hex() {
		t.Fatalf("existing form should be updated in place, got %q", fc.updateFormID)
	}
}

// 并发双击：查不到 → 建失败 → 回查复用已存在的表
func TestCreateForm_ConcurrentCreateReusesExisting(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	lookups := 0
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{createErr: errors.New("rpc error: duplicate key")},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				lookups++
				if lookups == 1 {
					return nil, formModel.ErrNotFound
				}
				return &formModel.EntryForm{ID: formID, UserId: userID, Cycle: cycle}, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
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
		t.Fatalf("concurrent create should self-heal, got %v", err)
	}
}

// 建表失败且本届确实无表：错误上抛，不静默成功
func TestCreateForm_CreateFailurePropagates(t *testing.T) {
	userID := primitive.NewObjectID()
	bomErr := errors.New("rpc error: mongo down")
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{createErr: bomErr},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
		},
		ScheduleModel: &fakeScheduleModel{},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); !errors.Is(err, bomErr) {
		t.Fatalf("create failure without existing form should propagate, got %v", err)
	}
}

// schedule 关联失败时回滚本次新建的表
func TestCreateForm_RollbackOnScheduleFailure(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	deletedID := ""
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deletedID = id
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return nil, errSchedule
			},
		},
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			findFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); !errors.Is(err, errSchedule) {
		t.Fatalf("schedule failure should propagate, got %v", err)
	}
	if deletedID != formID.Hex() {
		t.Fatalf("newly created form should be rolled back, got %q", deletedID)
	}
}

// 守卫之后管理员改状态导致 upsert 命中唯一索引：读回是录取态则拒绝并回滚
func TestCreateForm_AdmittedDuringRaceRejected(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	deletedID := ""
	dupErr := mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 11000}}}
	// 第 1 次读是入口守卫（此时尚未录取，放行）；之后读回时管理员已置为已转正
	reads := 0
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deletedID = id
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return nil, dupErr
			},
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				reads++
				if reads == 1 {
					return &scheduleModel.Schedule{ID: primitive.NewObjectID(), UserID: userID, AdmissionStatus: globalKey.Registered}, nil
				}
				return &scheduleModel.Schedule{ID: primitive.NewObjectID(), UserID: userID, AdmissionStatus: globalKey.Formal}, nil
			},
		},
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			findFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	_, err := l.CreateForm(newCreateReq())
	if err == nil || err.Error() != "已是正式成员，无需重复报名" {
		t.Fatalf("admitted-during-race should be rejected, got %v", err)
	}
	if deletedID != formID.Hex() {
		t.Fatalf("newly created form should be rolled back, got %q", deletedID)
	}
}

// 并发插入导致 upsert 命中唯一索引：读回非录取态则放行（不误判为已录取）
func TestCreateForm_DuplicateKeyFromConcurrencyAllowed(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	dupErr := mongo.WriteException{WriteErrors: []mongo.WriteError{{Code: 11000}}}
	deleted := false
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deleted = true
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return nil, dupErr
			},
			// 读回是并发请求刚插入的"已报名"，不是录取态
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID, AdmissionStatus: globalKey.Registered}, nil
			},
		},
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			updateFn: func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
				return &mongo.UpdateResult{MatchedCount: 1}, nil
			},
			findFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); err != nil {
		t.Fatalf("duplicate key from concurrency should be allowed, got %v", err)
	}
	if deleted {
		t.Fatal("form should not be rolled back on concurrent-insert duplicate key")
	}
}

// 并发场景：表已被其他请求关联为 entry_form_id 时，回滚不得删除它
func TestCreateForm_NoRollbackWhenFormAlreadyAssociated(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	deleted := false
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deleted = true
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return nil, errSchedule
			},
		},
		// userinfo 已把 entry_form_id 指向本次新建的表（另一请求已关联）
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			findFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID, EntryFormID: formID}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); err == nil {
		t.Fatal("schedule failure should propagate")
	}
	if deleted {
		t.Fatal("form already associated by a concurrent request must not be deleted")
	}
}

// 幂等更新既有表时，后续失败不得删除用户已提交的表
func TestCreateForm_NoRollbackWhenReusingExisting(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	deleted := false
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return &formModel.EntryForm{ID: formID, UserId: userID, Cycle: cycle}, nil
			},
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deleted = true
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return nil, errSchedule
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); err == nil {
		t.Fatal("schedule failure should propagate")
	}
	if deleted {
		t.Fatal("existing (reused) form must not be deleted on later failure")
	}
}

// userinfo 关联缺失（MatchedCount=0）时回滚
func TestCreateForm_RollbackOnMissingUserInfo(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	deletedID := ""
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				deletedID = id
				return 1, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID}, nil
			},
		},
		UserInfoModelClient: &fakeUserInfoUpdateModel{
			updateFn: func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
				return &mongo.UpdateResult{MatchedCount: 0}, nil
			},
			findFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: userID}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); err == nil {
		t.Fatal("missing userinfo should fail")
	}
	if deletedID != formID.Hex() {
		t.Fatalf("form should be rolled back when userinfo missing, got %q", deletedID)
	}
}

// 无 schedule 的老用户重报：upsert 会创建，不再直接失败
func TestCreateForm_MissingScheduleIsCreated(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	findCalls := 0
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: formID.Hex()},
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
		},
		ScheduleModel: &fakeScheduleModel{
			// 第一次（录取守卫）无记录 → 放行；upsert 后第二次查得到
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				findCalls++
				if findCalls == 1 {
					return nil, scheduleModel.ErrNotFound
				}
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
		t.Fatalf("missing schedule should be created via upsert, got %v", err)
	}
}

// 已转正成员不得重复报名
func TestCreateForm_AdmittedMemberRejected(t *testing.T) {
	userID := primitive.NewObjectID()
	svcCtx := &svc.ServiceContext{
		FormClient: &fakeCreateFormClient{formID: primitive.NewObjectID().Hex()},
		ScheduleModel: &fakeScheduleModel{
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: primitive.NewObjectID(), UserID: userID, AdmissionStatus: globalKey.Formal}, nil
			},
		},
	}
	l := NewCreateFormLogic(createCtx(userID.Hex()), svcCtx)
	if _, err := l.CreateForm(newCreateReq()); err == nil {
		t.Fatal("admitted member should be rejected")
	}
	if svcCtx.FormClient.(*fakeCreateFormClient).createCalled {
		t.Fatal("no form write should happen for admitted member")
	}
}

// 空身份拒绝
func TestCreateForm_EmptyUserId(t *testing.T) {
	l := NewCreateFormLogic(createCtx(""), &svc.ServiceContext{})
	if _, err := l.CreateForm(newCreateReq()); err == nil || err.Error() != "身份缺失" {
		t.Fatalf("empty user id should be rejected, got %v", err)
	}
}

// 存量表无 cycle 但 createAt 落在本届：视为本届，覆盖而非重复建表
func TestCreateForm_LegacyCurrentCycleFormIsUpdated(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	fc := &fakeCreateFormClient{}
	svcCtx := &svc.ServiceContext{
		FormClient: fc,
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
			findByUserFn: func(ctx context.Context, userId string) (*formModel.EntryForm, error) {
				// 无 cycle 字段，createAt 为当前时间 → EffectiveCycle 为本届
				return &formModel.EntryForm{ID: formID, UserId: userID, CreateAt: time.Now()}, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
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
		t.Fatalf("legacy current-cycle form should be updated, got %v", err)
	}
	if fc.createCalled {
		t.Fatal("legacy current-cycle form must not be duplicated")
	}
	if !fc.updateCalled || fc.updateFormID != formID.Hex() {
		t.Fatalf("legacy form should be updated in place, got %q", fc.updateFormID)
	}
}

// 存量表无 cycle 且 createAt 为往届：视为往届，新建本届表
func TestCreateForm_LegacyPastCycleFormCreatesNew(t *testing.T) {
	userID := primitive.NewObjectID()
	pastFormID := primitive.NewObjectID()
	newFormID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	fc := &fakeCreateFormClient{formID: newFormID.Hex()}
	svcCtx := &svc.ServiceContext{
		FormClient: fc,
		EntryFormModel: &fakeEntryFormModel{
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
			findByUserFn: func(ctx context.Context, userId string) (*formModel.EntryForm, error) {
				return &formModel.EntryForm{ID: pastFormID, UserId: userID,
					CreateAt: time.Date(2020, 9, 1, 0, 0, 0, 0, time.UTC)}, nil
			},
		},
		ScheduleModel: &fakeScheduleModel{
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
		t.Fatalf("legacy past-cycle should create new form, got %v", err)
	}
	if !fc.createCalled {
		t.Fatal("legacy past-cycle form should trigger a new current-cycle form")
	}
}

var _ = mon.ErrNotFound
