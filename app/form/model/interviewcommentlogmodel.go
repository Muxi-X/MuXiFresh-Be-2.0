package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var _ InterviewCommentLogModel = (*customInterviewCommentLogModel)(nil)

type (
	// InterviewCommentLogModel 只追加 interview_comment_log（面评历史，暂不对外提供查询）。
	InterviewCommentLogModel interface {
		Append(ctx context.Context, log *InterviewCommentLog) error
	}

	customInterviewCommentLogModel struct {
		conn *mon.Model
	}
)

// NewInterviewCommentLogModel returns a model for the mongo.
func NewInterviewCommentLogModel(url, db, collection string) InterviewCommentLogModel {
	return &customInterviewCommentLogModel{conn: mon.MustNewModel(url, db, collection)}
}

func (m *customInterviewCommentLogModel) Append(ctx context.Context, log *InterviewCommentLog) error {
	if log.ID.IsZero() {
		log.ID = primitive.NewObjectID()
	}
	if log.OperatedAt.IsZero() {
		log.OperatedAt = time.Now()
	}

	_, err := m.conn.InsertOne(ctx, log)
	return err
}
