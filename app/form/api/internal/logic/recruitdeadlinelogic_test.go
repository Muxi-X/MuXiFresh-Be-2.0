package logic

import (
	"context"
	"testing"
	"time"

	"MuXiFresh-Be-2.0/app/form/api/internal/svc"
	"MuXiFresh-Be-2.0/app/form/api/internal/types"
	"MuXiFresh-Be-2.0/app/form/model"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeRecruitSettingModel struct {
	model.RecruitSettingModel
	getFn func(ctx context.Context, cycle string) (*model.RecruitSetting, error)
	setFn func(ctx context.Context, cycle string, deadline time.Time, expectedRev int64, updateBy primitive.ObjectID, operatorType string) (int64, error)
}

func (f *fakeRecruitSettingModel) GetByCycle(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
	return f.getFn(ctx, cycle)
}

func (f *fakeRecruitSettingModel) SetDeadline(ctx context.Context, cycle string, deadline time.Time, expectedRev int64, updateBy primitive.ObjectID, operatorType string) (int64, error) {
	return f.setFn(ctx, cycle, deadline, expectedRev, updateBy, operatorType)
}

func recruitSvc(userType string, setting model.RecruitSettingModel) *svc.ServiceContext {
	return &svc.ServiceContext{
		RecruitSettingModel: setting,
		UserInfoModelClient: &fakeFormUserInfoModel{
			findOneFn: func(ctx context.Context, id string) (*usermodel.UserInfo, error) {
				return &usermodel.UserInfo{ID: primitive.NewObjectID(), UserType: userType}, nil
			},
		},
	}
}

func TestGetRecruitDeadline_StoredValue(t *testing.T) {
	stored, _ := model.ParseDeadline("2026-10-08 23:59:00")
	setting := &fakeRecruitSettingModel{
		getFn: func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
			return &model.RecruitSetting{Cycle: cycle, Deadline: stored, Rev: 3}, nil
		},
	}
	l := NewGetRecruitDeadlineLogic(context.Background(), recruitSvc(globalKey.Freshman, setting))

	resp, err := l.GetRecruitDeadline(&types.GetRecruitDeadlineReq{Cycle: "2026autumn"})
	if err != nil {
		t.Fatalf("stored value should be returned, got %v", err)
	}
	if resp.Deadline != "2026-10-08 23:59:00" || resp.Rev != 3 || resp.Cycle != "2026autumn" {
		t.Fatalf("unexpected resp %+v", resp)
	}
}

func TestGetRecruitDeadline_AutumnDefault(t *testing.T) {
	setting := &fakeRecruitSettingModel{
		getFn: func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
			return nil, model.ErrNotFound
		},
	}
	l := NewGetRecruitDeadlineLogic(context.Background(), recruitSvc(globalKey.Freshman, setting))

	resp, err := l.GetRecruitDeadline(&types.GetRecruitDeadlineReq{Cycle: "2026autumn"})
	if err != nil {
		t.Fatalf("unset autumn should fall back to default, got %v", err)
	}
	if resp.Deadline != "2026-10-06 23:59:00" || resp.Rev != 0 {
		t.Fatalf("unexpected autumn default %+v", resp)
	}
}

func TestGetRecruitDeadline_SpringUnset(t *testing.T) {
	setting := &fakeRecruitSettingModel{
		getFn: func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
			return nil, model.ErrNotFound
		},
	}
	l := NewGetRecruitDeadlineLogic(context.Background(), recruitSvc(globalKey.Freshman, setting))

	resp, err := l.GetRecruitDeadline(&types.GetRecruitDeadlineReq{Cycle: "2026spring"})
	if err != nil {
		t.Fatalf("unset spring should be returned as open, got %v", err)
	}
	if resp.Deadline != "" || resp.Rev != 0 || resp.Cycle != "2026spring" {
		t.Fatalf("unexpected spring unset %+v", resp)
	}
}

func TestGetRecruitDeadline_InvalidCycle(t *testing.T) {
	called := false
	setting := &fakeRecruitSettingModel{
		getFn: func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
			called = true
			return nil, model.ErrNotFound
		},
	}
	l := NewGetRecruitDeadlineLogic(context.Background(), recruitSvc(globalKey.Freshman, setting))

	if _, err := l.GetRecruitDeadline(&types.GetRecruitDeadlineReq{Cycle: "bad"}); err == nil {
		t.Fatal("invalid cycle should be rejected")
	}
	if called {
		t.Fatal("model must not be queried for invalid cycle")
	}
}

func TestGetRecruitDeadline_DefaultsToCurrentCycle(t *testing.T) {
	setting := &fakeRecruitSettingModel{
		getFn: func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
			return nil, model.ErrNotFound
		},
	}
	l := NewGetRecruitDeadlineLogic(context.Background(), recruitSvc(globalKey.Freshman, setting))

	resp, err := l.GetRecruitDeadline(&types.GetRecruitDeadlineReq{})
	if err != nil {
		t.Fatalf("empty cycle should use current cycle, got %v", err)
	}
	if resp.Cycle != model.CycleOf(time.Now()) {
		t.Fatalf("cycle = %q, want current cycle", resp.Cycle)
	}
}

func TestSetRecruitDeadline_NonAdminRejected(t *testing.T) {
	setCalled := false
	setting := &fakeRecruitSettingModel{
		setFn: func(ctx context.Context, cycle string, deadline time.Time, expectedRev int64, updateBy primitive.ObjectID, operatorType string) (int64, error) {
			setCalled = true
			return 0, nil
		},
	}
	l := NewSetRecruitDeadlineLogic(formCtxWithUser(primitive.NewObjectID().Hex()), recruitSvc(globalKey.Freshman, setting))

	_, err := l.SetRecruitDeadline(&types.SetRecruitDeadlineReq{Cycle: "2026autumn", Deadline: "2026-10-06 23:59:00"})
	if err == nil || err.Error() != "permission denied" {
		t.Fatalf("non-admin should be rejected, got %v", err)
	}
	if setCalled {
		t.Fatal("no write should happen for non-admin")
	}
}

func TestSetRecruitDeadline_Success(t *testing.T) {
	adminID := primitive.NewObjectID()
	setting := &fakeRecruitSettingModel{
		setFn: func(ctx context.Context, cycle string, deadline time.Time, expectedRev int64, updateBy primitive.ObjectID, operatorType string) (int64, error) {
			if cycle != "2026autumn" || expectedRev != 0 || updateBy != adminID || operatorType != globalKey.Admin {
				t.Fatalf("set called with unexpected args cycle=%q rev=%d updateBy=%v type=%q", cycle, expectedRev, updateBy, operatorType)
			}
			return 1, nil
		},
	}
	l := NewSetRecruitDeadlineLogic(formCtxWithUser(adminID.Hex()), recruitSvc(globalKey.Admin, setting))

	resp, err := l.SetRecruitDeadline(&types.SetRecruitDeadlineReq{Cycle: "2026autumn", Deadline: "2026-10-06 23:59:00", Rev: 0})
	if err != nil {
		t.Fatalf("admin write should succeed, got %v", err)
	}
	if resp.Rev != 1 || resp.Deadline != "2026-10-06 23:59:00" || resp.Cycle != "2026autumn" {
		t.Fatalf("unexpected resp %+v", resp)
	}
}

func TestSetRecruitDeadline_VersionConflict(t *testing.T) {
	setting := &fakeRecruitSettingModel{
		setFn: func(ctx context.Context, cycle string, deadline time.Time, expectedRev int64, updateBy primitive.ObjectID, operatorType string) (int64, error) {
			return 0, model.ErrVersionConflict
		},
	}
	l := NewSetRecruitDeadlineLogic(formCtxWithUser(primitive.NewObjectID().Hex()), recruitSvc(globalKey.Admin, setting))

	_, err := l.SetRecruitDeadline(&types.SetRecruitDeadlineReq{Cycle: "2026autumn", Deadline: "2026-10-06 23:59:00", Rev: 1})
	if err == nil || err.Error() != "deadline has been modified, please refresh" {
		t.Fatalf("conflict should map to refresh message, got %v", err)
	}
}

func TestSetRecruitDeadline_Validations(t *testing.T) {
	setCalled := false
	setting := &fakeRecruitSettingModel{
		setFn: func(ctx context.Context, cycle string, deadline time.Time, expectedRev int64, updateBy primitive.ObjectID, operatorType string) (int64, error) {
			setCalled = true
			return 0, nil
		},
	}
	cases := []struct {
		name string
		req  *types.SetRecruitDeadlineReq
		want string
	}{
		{"届次非法", &types.SetRecruitDeadlineReq{Cycle: "bad", Deadline: "2026-10-06 23:59:00"}, "invalid cycle"},
		{"截止时间格式非法", &types.SetRecruitDeadlineReq{Cycle: "2026autumn", Deadline: "2026/10/06"}, "invalid deadline"},
		{"版本号为负", &types.SetRecruitDeadlineReq{Cycle: "2026autumn", Deadline: "2026-10-06 23:59:00", Rev: -1}, "invalid rev"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			setCalled = false
			l := NewSetRecruitDeadlineLogic(formCtxWithUser(primitive.NewObjectID().Hex()), recruitSvc(globalKey.Admin, setting))
			_, err := l.SetRecruitDeadline(c.req)
			if err == nil || err.Error() != c.want {
				t.Fatalf("want %q, got %v", c.want, err)
			}
			if setCalled {
				t.Fatal("model must not be written when validation fails")
			}
		})
	}
}
