package model

import (
	"context"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ InterviewCommentLogModel = (*customInterviewCommentLogModel)(nil)

type (
	// InterviewCommentLogModel 追加并查询 interview_comment_log（面评历史）。
	InterviewCommentLogModel interface {
		Append(ctx context.Context, log *InterviewCommentLog) error
		// ListByFormID 按 rev 升序返回该报名表的面评历史版本，最多 limit 条。
		ListByFormID(ctx context.Context, formID string, limit int64) ([]*InterviewCommentLog, error)
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

func (m *customInterviewCommentLogModel) ListByFormID(ctx context.Context, formID string, limit int64) ([]*InterviewCommentLog, error) {
	oid, err := primitive.ObjectIDFromHex(formID)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	// 倒序取最近 limit 版再反转：被 limit 截断时丢的是最旧的版本，而非最新。
	var logs []*InterviewCommentLog
	err = m.conn.Find(ctx, &logs, bson.M{"formId": oid},
		options.Find().SetSort(bson.D{{Key: "rev", Value: -1}}).SetLimit(limit))
	switch err {
	case nil:
		reverseLogs(logs)
		return logs, nil
	case mon.ErrNotFound:
		return []*InterviewCommentLog{}, nil
	default:
		return nil, err
	}
}

func reverseLogs(logs []*InterviewCommentLog) {
	for i, j := 0, len(logs)-1; i < j; i, j = i+1, j-1 {
		logs[i], logs[j] = logs[j], logs[i]
	}
}
