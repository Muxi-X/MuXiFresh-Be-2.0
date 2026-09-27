package logic

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/svc"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/types"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/ctxData"
	"MuXiFresh-Be-2.0/common/globalKey"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"google.golang.org/grpc"
)

type fakeEntryFormModel struct {
	model.EntryFormModel
	setFn     func(ctx context.Context, formID, comment string, expectedRev int64, operatorID primitive.ObjectID, at time.Time) (int64, bool, error)
	findOneFn func(ctx context.Context, id string) (*model.EntryForm, error)
}

func (f *fakeEntryFormModel) SetInterviewComment(ctx context.Context, formID, comment string, expectedRev int64, operatorID primitive.ObjectID, at time.Time) (int64, bool, error) {
	return f.setFn(ctx, formID, comment, expectedRev, operatorID, at)
}

func (f *fakeEntryFormModel) FindOne(ctx context.Context, id string) (*model.EntryForm, error) {
	if f.findOneFn == nil {
		return nil, mon.ErrNotFound
	}
	return f.findOneFn(ctx, id)
}

type fakeCommentLogModel struct {
	model.InterviewCommentLogModel
	err      error
	appended []*model.InterviewCommentLog
	listFn   func(ctx context.Context, formID string, limit int64) ([]*model.InterviewCommentLog, error)
}

func (f *fakeCommentLogModel) Append(ctx context.Context, log *model.InterviewCommentLog) error {
	f.appended = append(f.appended, log)
	return f.err
}

func (f *fakeCommentLogModel) ListByFormID(ctx context.Context, formID string, limit int64) ([]*model.InterviewCommentLog, error) {
	if f.listFn == nil {
		return nil, nil
	}
	return f.listFn(ctx, formID, limit)
}

type fakeUserClient struct {
	userclient.UserClient
	userType string
}

func (f *fakeUserClient) GetUserType(ctx context.Context, in *userclient.GetUserTypeReq, opts ...grpc.CallOption) (*userclient.GetUserTypeResp, error) {
	return &userclient.GetUserTypeResp{UserType: f.userType}, nil
}

func commentCtx(uid string) context.Context {
	return context.WithValue(context.Background(), ctxData.CtxKeyJwtUserID, uid)
}

func commentSvc(userType string, form model.EntryFormModel, log model.InterviewCommentLogModel) *svc.ServiceContext {
	return &svc.ServiceContext{
		EntryFormModel:           form,
		UserClient:               &fakeUserClient{userType: userType},
		InterviewCommentLogModel: log,
	}
}

func TestSetInterviewComment_AppendsAuditLog(t *testing.T) {
	adminID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	var writtenAt time.Time
	form := &fakeEntryFormModel{
		setFn: func(ctx context.Context, gotFormID, comment string, expectedRev int64, operatorID primitive.ObjectID, at time.Time) (int64, bool, error) {
			if gotFormID != formID.Hex() || comment != "一面：基础扎实" || expectedRev != 2 || operatorID != adminID || at.IsZero() {
				t.Fatalf("set called with unexpected args formID=%q comment=%q rev=%d operator=%v at=%v", gotFormID, comment, expectedRev, operatorID, at)
			}
			writtenAt = at
			return 7, true, nil
		},
	}
	logModel := &fakeCommentLogModel{}
	l := NewSetInterviewCommentLogic(commentCtx(adminID.Hex()), commentSvc(globalKey.Admin, form, logModel))

	resp, err := l.SetInterviewComment(&types.SetInterviewCommentReq{FormID: formID.Hex(), Comment: "一面：基础扎实", Rev: 2})
	if err != nil {
		t.Fatalf("admin write should succeed, got %v", err)
	}
	// rev 应来自模型返回值（7），而非逻辑层自行 +1（会是 3），以验出硬编码回归
	if resp.Rev != 7 {
		t.Fatalf("resp rev = %d, want 7", resp.Rev)
	}

	if len(logModel.appended) != 1 {
		t.Fatalf("expected one history entry, got %d", len(logModel.appended))
	}
	entry := logModel.appended[0]
	if entry.FormID != formID || entry.Rev != 7 || entry.Comment != "一面：基础扎实" ||
		entry.OperatorID != adminID || entry.OperatorType != globalKey.Admin {
		t.Fatalf("unexpected history entry %+v", entry)
	}
	// 历史时间应与面评写入时间一致（同一个 now）
	if entry.OperatedAt.IsZero() || !entry.OperatedAt.Equal(writtenAt) {
		t.Fatalf("history operatedAt %v should equal write time %v", entry.OperatedAt, writtenAt)
	}
}

func TestSetInterviewComment_ConflictWritesNoLog(t *testing.T) {
	formID := primitive.NewObjectID()
	form := &fakeEntryFormModel{
		setFn: func(ctx context.Context, gotFormID, comment string, expectedRev int64, operatorID primitive.ObjectID, at time.Time) (int64, bool, error) {
			return 0, false, nil
		},
		findOneFn: func(ctx context.Context, id string) (*model.EntryForm, error) {
			return &model.EntryForm{ID: formID}, nil
		},
	}
	logModel := &fakeCommentLogModel{}
	l := NewSetInterviewCommentLogic(commentCtx(primitive.NewObjectID().Hex()), commentSvc(globalKey.Admin, form, logModel))

	_, err := l.SetInterviewComment(&types.SetInterviewCommentReq{FormID: formID.Hex(), Comment: "x", Rev: 1})
	if err == nil || err.Error() != "comment has been modified, please refresh" {
		t.Fatalf("conflict should be reported, got %v", err)
	}
	if len(logModel.appended) != 0 {
		t.Fatal("no history should be written on conflict")
	}
}

func TestSetInterviewComment_NonAdminWritesNoLog(t *testing.T) {
	setCalled := false
	form := &fakeEntryFormModel{
		setFn: func(ctx context.Context, gotFormID, comment string, expectedRev int64, operatorID primitive.ObjectID, at time.Time) (int64, bool, error) {
			setCalled = true
			return 1, true, nil
		},
	}
	logModel := &fakeCommentLogModel{}
	l := NewSetInterviewCommentLogic(commentCtx(primitive.NewObjectID().Hex()), commentSvc(globalKey.Freshman, form, logModel))

	_, err := l.SetInterviewComment(&types.SetInterviewCommentReq{FormID: primitive.NewObjectID().Hex(), Comment: "x", Rev: 0})
	if err == nil || err.Error() != "permission denied" {
		t.Fatalf("non-admin should be rejected, got %v", err)
	}
	if setCalled || len(logModel.appended) != 0 {
		t.Fatal("non-admin must not write comment or history")
	}
}

func TestSetInterviewComment_LogFailureStillSucceeds(t *testing.T) {
	adminID := primitive.NewObjectID()
	formID := primitive.NewObjectID()
	form := &fakeEntryFormModel{
		setFn: func(ctx context.Context, gotFormID, comment string, expectedRev int64, operatorID primitive.ObjectID, at time.Time) (int64, bool, error) {
			return 1, true, nil
		},
	}
	logModel := &fakeCommentLogModel{err: errors.New("db down")}
	l := NewSetInterviewCommentLogic(commentCtx(adminID.Hex()), commentSvc(globalKey.Admin, form, logModel))

	resp, err := l.SetInterviewComment(&types.SetInterviewCommentReq{FormID: formID.Hex(), Comment: "x", Rev: 0})
	if err != nil {
		t.Fatalf("log failure must not fail the comment write, got %v", err)
	}
	if resp.Rev != 1 {
		t.Fatalf("resp rev = %d, want 1", resp.Rev)
	}
}

func TestSetInterviewComment_OverLimitWritesNoLog(t *testing.T) {
	setCalled := false
	form := &fakeEntryFormModel{
		setFn: func(ctx context.Context, gotFormID, comment string, expectedRev int64, operatorID primitive.ObjectID, at time.Time) (int64, bool, error) {
			setCalled = true
			return 1, true, nil
		},
	}
	logModel := &fakeCommentLogModel{}
	l := NewSetInterviewCommentLogic(commentCtx(primitive.NewObjectID().Hex()), commentSvc(globalKey.Admin, form, logModel))

	tooLong := strings.Repeat("好", maxInterviewCommentLen+1)
	_, err := l.SetInterviewComment(&types.SetInterviewCommentReq{FormID: primitive.NewObjectID().Hex(), Comment: tooLong, Rev: 0})
	if err == nil {
		t.Fatal("over-limit comment should be rejected")
	}
	if setCalled || len(logModel.appended) != 0 {
		t.Fatal("validation failure must not write comment or history")
	}
}

func TestSetInterviewComment_InvalidOperatorRejected(t *testing.T) {
	form := &fakeEntryFormModel{}
	logModel := &fakeCommentLogModel{}
	l := NewSetInterviewCommentLogic(commentCtx("not-an-objectid"), commentSvc(globalKey.Admin, form, logModel))

	if _, err := l.SetInterviewComment(&types.SetInterviewCommentReq{FormID: primitive.NewObjectID().Hex(), Comment: "x"}); err == nil {
		t.Fatal("invalid operator id should be rejected")
	}
	if len(logModel.appended) != 0 {
		t.Fatal("no history should be written for invalid operator")
	}
}

func TestSetInterviewComment_ModelErrorPassthrough(t *testing.T) {
	sentinel := errors.New("db down")
	form := &fakeEntryFormModel{
		setFn: func(ctx context.Context, gotFormID, comment string, expectedRev int64, operatorID primitive.ObjectID, at time.Time) (int64, bool, error) {
			return 0, false, sentinel
		},
	}
	logModel := &fakeCommentLogModel{}
	l := NewSetInterviewCommentLogic(commentCtx(primitive.NewObjectID().Hex()), commentSvc(globalKey.Admin, form, logModel))

	_, err := l.SetInterviewComment(&types.SetInterviewCommentReq{FormID: primitive.NewObjectID().Hex(), Comment: "x", Rev: 0})
	if !errors.Is(err, sentinel) {
		t.Fatalf("model error should pass through, got %v", err)
	}
	if len(logModel.appended) != 0 {
		t.Fatal("no history should be written on model error")
	}
}
