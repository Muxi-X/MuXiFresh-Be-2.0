package logic

import (
	"context"
	"testing"

	taskmodel "MuXiFresh-Be-2.0/app/task/model"

	"google.golang.org/grpc/metadata"
)

// fakeCommentModel：内嵌接口，只重写测试路径用到的方法（FindOne/Insert/Delete）
type fakeCommentModel struct {
	taskmodel.CommentModel
	findOneFn            func(ctx context.Context, id string) (*taskmodel.Comment, error)
	findBySubmissionIDFn func(ctx context.Context, submissionID string) ([]*taskmodel.Comment, error)
	insertFn             func(ctx context.Context, data *taskmodel.Comment) error
	deleteFn             func(ctx context.Context, id string) (int64, error)
}

func (f *fakeCommentModel) FindOne(ctx context.Context, id string) (*taskmodel.Comment, error) {
	return f.findOneFn(ctx, id)
}

func (f *fakeCommentModel) FindBySubmissionID(ctx context.Context, submissionID string) ([]*taskmodel.Comment, error) {
	if f.findBySubmissionIDFn != nil {
		return f.findBySubmissionIDFn(ctx, submissionID)
	}
	return []*taskmodel.Comment{}, nil
}

func (f *fakeCommentModel) Insert(ctx context.Context, data *taskmodel.Comment) error {
	if f.insertFn != nil {
		return f.insertFn(ctx, data)
	}
	return nil
}

func (f *fakeCommentModel) Delete(ctx context.Context, id string) (int64, error) {
	if f.deleteFn != nil {
		return f.deleteFn(ctx, id)
	}
	return 1, nil
}

// callerCtx 构造携带调用者身份的 incoming context（与 API 层注入的 metadata 同键）
func callerCtx(t *testing.T, callerID string) context.Context {
	t.Helper()
	return metadata.NewIncomingContext(context.Background(), metadata.Pairs(callerIDKey, callerID))
}
