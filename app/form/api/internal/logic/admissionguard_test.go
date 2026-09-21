package logic

import (
	"context"
	"testing"

	scheduleModel "MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/common/globalKey"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestEnsureNotAdmittedMember(t *testing.T) {
	uid := primitive.NewObjectID()

	cases := []struct {
		name    string
		status  string
		wantErr bool
	}{
		{"已转正拒绝", globalKey.Formal, true},
		{"实习期拒绝", globalKey.Internship, true},
		{"已报名放行", globalKey.Registered, false},
		{"未报名放行", "未报名", false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			m := &fakeScheduleModel{
				findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
					return &scheduleModel.Schedule{ID: primitive.NewObjectID(), UserID: uid, AdmissionStatus: c.status}, nil
				},
			}
			err := ensureNotAdmittedMember(context.Background(), m, uid.Hex())
			if c.wantErr && err == nil {
				t.Fatalf("status %q should be rejected", c.status)
			}
			if !c.wantErr && err != nil {
				t.Fatalf("status %q should be allowed, got %v", c.status, err)
			}
		})
	}
}

// 无进度记录视为尚未报名，放行
func TestEnsureNotAdmittedMember_NoScheduleAllows(t *testing.T) {
	m := &fakeScheduleModel{
		findByUser: func(ctx context.Context, userId string) (*scheduleModel.Schedule, error) {
			return nil, scheduleModel.ErrNotFound
		},
	}
	if err := ensureNotAdmittedMember(context.Background(), m, primitive.NewObjectID().Hex()); err != nil {
		t.Fatalf("missing schedule should be allowed as not-yet-registered, got %v", err)
	}
}
