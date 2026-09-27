package logic

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/internal/svc"
)

type fakeRecruitSettingModel struct {
	model.RecruitSettingModel
	getFn func(ctx context.Context, cycle string) (*model.RecruitSetting, error)
}

func (f *fakeRecruitSettingModel) GetByCycle(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
	return f.getFn(ctx, cycle)
}

func svcWithSetting(getFn func(ctx context.Context, cycle string) (*model.RecruitSetting, error)) *svc.ServiceContext {
	return &svc.ServiceContext{RecruitSettingModel: &fakeRecruitSettingModel{getFn: getFn}}
}

func TestEnsureWithinDeadline_StoredDeadline(t *testing.T) {
	now := time.Now()
	past, _ := model.ParseDeadline(model.FormatDeadline(now.Add(-time.Hour)))
	future, _ := model.ParseDeadline(model.FormatDeadline(now.Add(time.Hour)))

	if err := ensureWithinDeadline(context.Background(), svcWithSetting(func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
		return &model.RecruitSetting{Cycle: cycle, Deadline: past}, nil
	}), "2026autumn", now); err == nil || err.Error() != "未在报名时间内" {
		t.Fatalf("past stored deadline should reject with 未在报名时间内, got %v", err)
	}

	if err := ensureWithinDeadline(context.Background(), svcWithSetting(func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
		return &model.RecruitSetting{Cycle: cycle, Deadline: future}, nil
	}), "2026autumn", now); err != nil {
		t.Fatalf("future stored deadline should allow, got %v", err)
	}
}

func TestEnsureWithinDeadline_AutumnDefault(t *testing.T) {
	now := time.Now()
	notFound := func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
		return nil, model.ErrNotFound
	}

	pastCycle := fmt.Sprintf("%dautumn", now.Year()-1)
	if err := ensureWithinDeadline(context.Background(), svcWithSetting(notFound), pastCycle, now); err == nil {
		t.Fatalf("unset past autumn (%s) should fall back to past default and reject", pastCycle)
	}

	futureCycle := fmt.Sprintf("%dautumn", now.Year()+1)
	if err := ensureWithinDeadline(context.Background(), svcWithSetting(notFound), futureCycle, now); err != nil {
		t.Fatalf("unset future autumn (%s) should fall back to future default and allow, got %v", futureCycle, err)
	}
}

func TestEnsureWithinDeadline_SpringOpen(t *testing.T) {
	now := time.Now()
	notFound := func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
		return nil, model.ErrNotFound
	}
	springCycle := fmt.Sprintf("%dspring", now.Year())
	if err := ensureWithinDeadline(context.Background(), svcWithSetting(notFound), springCycle, now); err != nil {
		t.Fatalf("unset spring should be open, got %v", err)
	}
}

func TestEnsureWithinDeadline_ModelErrorPassthrough(t *testing.T) {
	sentinel := errors.New("db down")
	err := ensureWithinDeadline(context.Background(), svcWithSetting(func(ctx context.Context, cycle string) (*model.RecruitSetting, error) {
		return nil, sentinel
	}), "2026autumn", time.Now())
	if !errors.Is(err, sentinel) {
		t.Fatalf("model error should pass through, got %v", err)
	}
}
