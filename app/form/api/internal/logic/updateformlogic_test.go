package logic

import (
	"context"
	"errors"
	"testing"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	formModel "MuXiFresh-Be-2.0/app/form/model"
	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// ownerForm 返回一个归属 userId 的报名表（带本届 cycle），供归属校验通过
func ownerForm(userID, formID primitive.ObjectID) *formModel.EntryForm {
	return &formModel.EntryForm{ID: formID, UserId: userID}
}

func updateSvc(userID primitive.ObjectID, fc *fakeCreateFormClient, efm *fakeEntryFormModel, sm *fakeScheduleModel) *svc.ServiceContext {
	return &svc.ServiceContext{
		UserInfoModelClient: &fakeFormUserInfoModel{},
		FormClient:          fc,
		EntryFormModel:      efm,
		ScheduleModel:       sm,
	}
}

func TestUpdateForm_OwnIdPasses(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	fc := &fakeCreateFormClient{}
	svcCtx := updateSvc(userID, fc,
		&fakeEntryFormModel{
			findOneFn: func(ctx context.Context, id string) (*formModel.EntryForm, error) {
				return ownerForm(userID, formID), nil
			},
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return &formModel.EntryForm{ID: formID, UserId: userID, Cycle: cycle}, nil
			},
		}, &fakeScheduleModel{
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID}, nil
			},
		})
	l := NewUpdateFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	if _, err := l.UpdateForm(&types.CreateReq{FormId: formID.Hex()}); err != nil {
		t.Fatalf("own form id should pass, got %v", err)
	}
	if !fc.updateCalled {
		t.Fatal("own current-cycle form should be updated")
	}
}

// 他人的表：拒绝
func TestUpdateForm_OtherIdRejected(t *testing.T) {
	userID := primitive.NewObjectID()
	otherID := primitive.NewObjectID()
	attacker := primitive.NewObjectID()
	fc := &fakeCreateFormClient{}
	svcCtx := updateSvc(attacker, fc,
		&fakeEntryFormModel{
			findOneFn: func(ctx context.Context, id string) (*formModel.EntryForm, error) {
				return &formModel.EntryForm{ID: otherID, UserId: userID}, nil
			},
		}, &fakeScheduleModel{})
	l := NewUpdateFormLogic(formCtxWithUser(attacker.Hex()), svcCtx)

	_, err := l.UpdateForm(&types.CreateReq{FormId: otherID.Hex()})
	if err == nil || err.Error() != "无权修改该报名表" {
		t.Fatalf("other's form should be rejected, got: %v", err)
	}
	if fc.updateCalled || fc.createCalled {
		t.Fatal("no write should happen for unauthorized access")
	}
}

// 表不存在：归一为无权，不泄露存在性
func TestUpdateForm_MissingFormRejected(t *testing.T) {
	userID := primitive.NewObjectID()
	fc := &fakeCreateFormClient{}
	svcCtx := updateSvc(userID, fc,
		&fakeEntryFormModel{
			findOneFn: func(ctx context.Context, id string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
		}, &fakeScheduleModel{})
	l := NewUpdateFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	_, err := l.UpdateForm(&types.CreateReq{FormId: primitive.NewObjectID().Hex()})
	if err == nil || err.Error() != "无权修改该报名表" {
		t.Fatalf("missing form should be normalized, got: %v", err)
	}
}

func TestUpdateForm_EmptyUserId(t *testing.T) {
	l := NewUpdateFormLogic(formCtxWithUser(""), &svc.ServiceContext{})
	if _, err := l.UpdateForm(&types.CreateReq{FormId: primitive.NewObjectID().Hex()}); err == nil || err.Error() != "身份缺失" {
		t.Fatalf("empty user id should be rejected, got: %v", err)
	}
}

// 往届表重报：走公共流程建本届表（本届无表时新建）
func TestUpdateForm_PastCycleCreatesCurrentCycleForm(t *testing.T) {
	userID := primitive.NewObjectID()
	pastFormID := primitive.NewObjectID()
	newFormID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	fc := &fakeCreateFormClient{formID: newFormID.Hex()}
	svcCtx := updateSvc(userID, fc,
		&fakeEntryFormModel{
			findOneFn: func(ctx context.Context, id string) (*formModel.EntryForm, error) {
				return ownerForm(userID, pastFormID), nil
			},
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return nil, formModel.ErrNotFound
			},
		},
		&fakeScheduleModel{
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID}, nil
			},
		})
	svcCtx.UserInfoModelClient.(*fakeFormUserInfoModel).updateFn = func(ctx context.Context, data *usermodel.UserInfo) (*mongo.UpdateResult, error) {
		if data.EntryFormID != newFormID {
			t.Fatalf("userinfo should point at new form, got %s", data.EntryFormID)
		}
		return &mongo.UpdateResult{MatchedCount: 1}, nil
	}
	l := NewUpdateFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	if _, err := l.UpdateForm(&types.CreateReq{FormId: pastFormID.Hex()}); err != nil {
		t.Fatalf("past cycle re-submit should pass, got %v", err)
	}
	if !fc.createCalled {
		t.Fatal("past cycle should create a current-cycle form")
	}
}

// 重报成功后客户端仍持旧 form_id 再次提交：归属校验放行，幂等落到本届表
func TestUpdateForm_StaleFormIdAfterResubmitStillAccepted(t *testing.T) {
	userID := primitive.NewObjectID()
	staleFormID := primitive.NewObjectID()
	currentFormID := primitive.NewObjectID()
	sid := primitive.NewObjectID()
	fc := &fakeCreateFormClient{}
	svcCtx := updateSvc(userID, fc,
		&fakeEntryFormModel{
			// 旧 id 仍是本人所有
			findOneFn: func(ctx context.Context, id string) (*formModel.EntryForm, error) {
				return ownerForm(userID, staleFormID), nil
			},
			// 本届表已存在（重报已产生），应更新它而非新建
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return &formModel.EntryForm{ID: currentFormID, UserId: userID, Cycle: cycle}, nil
			},
		},
		&fakeScheduleModel{
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: sid, UserID: userID}, nil
			},
		})
	l := NewUpdateFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	if _, err := l.UpdateForm(&types.CreateReq{FormId: staleFormID.Hex()}); err != nil {
		t.Fatalf("stale own form id should still be accepted, got %v", err)
	}
	if fc.createCalled {
		t.Fatal("should update existing current-cycle form, not create")
	}
	if fc.updateFormID != currentFormID.Hex() {
		t.Fatalf("write should target current-cycle form, got %q", fc.updateFormID)
	}
}

// 已转正成员不得报名
func TestUpdateForm_AdmittedMemberRejected(t *testing.T) {
	userID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	fc := &fakeCreateFormClient{}
	svcCtx := updateSvc(userID, fc,
		&fakeEntryFormModel{
			findOneFn: func(ctx context.Context, id string) (*formModel.EntryForm, error) {
				return ownerForm(userID, formID), nil
			},
		}, &fakeScheduleModel{
			findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
				return &scheduleModel.Schedule{ID: primitive.NewObjectID(), UserID: userID, AdmissionStatus: globalKey.Formal}, nil
			},
		})
	l := NewUpdateFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	if _, err := l.UpdateForm(&types.CreateReq{FormId: formID.Hex()}); err == nil {
		t.Fatal("admitted member should not re-submit")
	}
	if fc.createCalled || fc.updateCalled {
		t.Fatal("no write should happen for admitted member")
	}
}

// 关联失败：复用既有表时不得回滚用户已提交的表
func TestUpdateForm_ReusedFormSurvivesAssociationFailure(t *testing.T) {
	userID := primitive.NewObjectID()
	pastFormID := primitive.NewObjectID()
	currentFormID := primitive.NewObjectID()
	upsertErr := errors.New("upsert boom")
	fc := &fakeCreateFormClient{}
	svcCtx := updateSvc(userID, fc,
		&fakeEntryFormModel{
			findOneFn: func(ctx context.Context, id string) (*formModel.EntryForm, error) {
				return ownerForm(userID, pastFormID), nil
			},
			findByCycleFn: func(ctx context.Context, userId, cycle string) (*formModel.EntryForm, error) {
				return &formModel.EntryForm{ID: currentFormID, UserId: userID, Cycle: cycle}, nil
			},
			deleteFn: func(ctx context.Context, id string) (int64, error) {
				t.Fatalf("must not delete reused current-cycle form %s", id)
				return 0, nil
			},
		},
		&fakeScheduleModel{
			upsertFn: func(ctx context.Context, data *scheduleModel.Schedule) (*mongo.UpdateResult, error) {
				return nil, upsertErr
			},
		})
	l := NewUpdateFormLogic(formCtxWithUser(userID.Hex()), svcCtx)

	if _, err := l.UpdateForm(&types.CreateReq{FormId: pastFormID.Hex()}); err == nil {
		t.Fatal("association failure should propagate")
	}
}
