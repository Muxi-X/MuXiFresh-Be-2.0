package logic

import (
	"context"
	"testing"
	"time"

	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	usermodel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/globalKey"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type fakeHistoryUserInfoModel struct {
	usermodel.UserInfoModel
	findByUserIdsFn func(ctx context.Context, ids []string) ([]*usermodel.UserInfo, error)
}

func (f *fakeHistoryUserInfoModel) FindByUserIds(ctx context.Context, ids []string) ([]*usermodel.UserInfo, error) {
	return f.findByUserIdsFn(ctx, ids)
}

func historySvc(userType string, log model.InterviewCommentLogModel, info usermodel.UserInfoModel) *svc.ServiceContext {
	return &svc.ServiceContext{
		UserClient:               &fakeUserClient{userType: userType},
		InterviewCommentLogModel: log,
		UserInfoModel:            info,
	}
}

func TestGetInterviewCommentHistory_RowsWithOperatorNames(t *testing.T) {
	op1 := primitive.NewObjectID()
	op2 := primitive.NewObjectID()
	formID := primitive.NewObjectID()

	var gotFormID string
	var gotLimit int64
	logModel := &fakeCommentLogModel{
		listFn: func(ctx context.Context, fID string, limit int64) ([]*model.InterviewCommentLog, error) {
			gotFormID, gotLimit = fID, limit
			return []*model.InterviewCommentLog{
				{FormID: formID, Rev: 1, Comment: "初稿", OperatorID: op1, OperatorType: globalKey.Admin,
					OperatedAt: time.Date(2026, time.October, 5, 6, 3, 7, 0, time.UTC)},
				{FormID: formID, Rev: 2, Comment: "二稿", OperatorID: op2, OperatorType: globalKey.SuperAdmin},
			}, nil
		},
	}
	infoModel := &fakeHistoryUserInfoModel{
		findByUserIdsFn: func(ctx context.Context, ids []string) ([]*usermodel.UserInfo, error) {
			return []*usermodel.UserInfo{
				{ID: op1, Name: "张三"},
				{ID: op2, Name: "李四"},
			}, nil
		},
	}
	l := NewGetInterviewCommentHistoryLogic(commentCtx(primitive.NewObjectID().Hex()), historySvc(globalKey.Admin, logModel, infoModel))

	resp, err := l.GetInterviewCommentHistory(&types.GetInterviewCommentHistoryReq{FormID: formID.Hex(), Limit: 10})
	if err != nil {
		t.Fatalf("admin should read history, got %v", err)
	}
	if gotFormID != formID.Hex() || gotLimit != 10 {
		t.Fatalf("model called with formID=%q limit=%d", gotFormID, gotLimit)
	}
	if len(resp.Rows) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(resp.Rows))
	}
	if resp.Rows[0].Rev != 1 || resp.Rows[0].Comment != "初稿" || resp.Rows[0].OperatorName != "张三" || resp.Rows[0].OperatorType != globalKey.Admin {
		t.Fatalf("unexpected row 0 %+v", resp.Rows[0])
	}
	if resp.Rows[0].OperatedAt != "2026-10-05 14:03:07" {
		t.Fatalf("operated_at = %q, want 2026-10-05 14:03:07", resp.Rows[0].OperatedAt)
	}
	if resp.Rows[1].Rev != 2 || resp.Rows[1].OperatorName != "李四" {
		t.Fatalf("unexpected row 1 %+v", resp.Rows[1])
	}
}

func TestGetInterviewCommentHistory_EmptyReturnsEmptyRows(t *testing.T) {
	logModel := &fakeCommentLogModel{
		listFn: func(ctx context.Context, fID string, limit int64) ([]*model.InterviewCommentLog, error) {
			return nil, nil
		},
	}
	l := NewGetInterviewCommentHistoryLogic(commentCtx(primitive.NewObjectID().Hex()), historySvc(globalKey.Admin, logModel, nil))

	resp, err := l.GetInterviewCommentHistory(&types.GetInterviewCommentHistoryReq{FormID: primitive.NewObjectID().Hex()})
	if err != nil {
		t.Fatalf("empty history should not error, got %v", err)
	}
	if resp.Rows == nil || len(resp.Rows) != 0 {
		t.Fatalf("empty history should return empty non-nil rows, got %#v", resp.Rows)
	}
}

func TestGetInterviewCommentHistory_UnknownOperatorNameBlank(t *testing.T) {
	logModel := &fakeCommentLogModel{
		listFn: func(ctx context.Context, fID string, limit int64) ([]*model.InterviewCommentLog, error) {
			return []*model.InterviewCommentLog{
				{Rev: 1, Comment: "x", OperatorID: primitive.NewObjectID(), OperatorType: globalKey.Admin},
			}, nil
		},
	}
	infoModel := &fakeHistoryUserInfoModel{
		findByUserIdsFn: func(ctx context.Context, ids []string) ([]*usermodel.UserInfo, error) {
			return nil, nil
		},
	}
	l := NewGetInterviewCommentHistoryLogic(commentCtx(primitive.NewObjectID().Hex()), historySvc(globalKey.Admin, logModel, infoModel))

	resp, err := l.GetInterviewCommentHistory(&types.GetInterviewCommentHistoryReq{FormID: primitive.NewObjectID().Hex()})
	if err != nil {
		t.Fatalf("missing operator userinfo should not error, got %v", err)
	}
	if len(resp.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(resp.Rows))
	}
	if resp.Rows[0].OperatorName != "" {
		t.Fatalf("unknown operator should have blank name, got %q", resp.Rows[0].OperatorName)
	}
}

func TestGetInterviewCommentHistory_LimitClamp(t *testing.T) {
	cases := []struct {
		in   int64
		want int64
	}{
		{0, defaultHistoryLimit},
		{-5, defaultHistoryLimit},
		{10, 10},
		{9999, maxHistoryLimit},
	}
	for _, c := range cases {
		var gotLimit int64
		logModel := &fakeCommentLogModel{
			listFn: func(ctx context.Context, fID string, limit int64) ([]*model.InterviewCommentLog, error) {
				gotLimit = limit
				return nil, nil
			},
		}
		l := NewGetInterviewCommentHistoryLogic(commentCtx(primitive.NewObjectID().Hex()), historySvc(globalKey.Admin, logModel, nil))
		_, err := l.GetInterviewCommentHistory(&types.GetInterviewCommentHistoryReq{FormID: primitive.NewObjectID().Hex(), Limit: c.in})
		if err != nil {
			t.Fatalf("in=%d should be accepted, got %v", c.in, err)
		}
		if gotLimit != c.want {
			t.Fatalf("limit in=%d: got %d, want %d", c.in, gotLimit, c.want)
		}
	}
}

func TestGetInterviewCommentHistory_NonAdminRejected(t *testing.T) {
	called := false
	logModel := &fakeCommentLogModel{
		listFn: func(ctx context.Context, fID string, limit int64) ([]*model.InterviewCommentLog, error) {
			called = true
			return nil, nil
		},
	}
	l := NewGetInterviewCommentHistoryLogic(commentCtx(primitive.NewObjectID().Hex()), historySvc(globalKey.Freshman, logModel, nil))

	_, err := l.GetInterviewCommentHistory(&types.GetInterviewCommentHistoryReq{FormID: primitive.NewObjectID().Hex()})
	if err == nil || err.Error() != "permission denied" {
		t.Fatalf("non-admin should be rejected, got %v", err)
	}
	if called {
		t.Fatal("non-admin must not query history")
	}
}

func TestGetInterviewCommentHistory_InvalidFormID(t *testing.T) {
	l := NewGetInterviewCommentHistoryLogic(commentCtx(primitive.NewObjectID().Hex()), historySvc(globalKey.Admin, &fakeCommentLogModel{}, nil))

	_, err := l.GetInterviewCommentHistory(&types.GetInterviewCommentHistoryReq{FormID: "bad"})
	if err == nil || err.Error() != "invalid form id" {
		t.Fatalf("invalid form id should be rejected, got %v", err)
	}
}
