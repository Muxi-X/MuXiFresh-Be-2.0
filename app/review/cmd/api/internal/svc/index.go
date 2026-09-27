package svc

import (
	"context"
	"time"

	"MuXiFresh-Be-2.0/common/mongodb"

	"go.mongodb.org/mongo-driver/bson"
)

const interviewCommentLogCollection = "interview_comment_log"

// EnsureInterviewCommentLogIndexes 确保面评历史可按 {formId, rev} 高效查询。
// 该索引只影响查询性能，不影响正确性，故由调用方按 best-effort 处理、不阻塞启动。
func EnsureInterviewCommentLogIndexes(url, db string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongodb.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	return mongodb.EnsureIndex(ctx, client, db, mongodb.IndexSpec{
		Collection: interviewCommentLogCollection,
		Name:       "interview_comment_log_formId_rev",
		Keys:       bson.D{{Key: "formId", Value: 1}, {Key: "rev", Value: 1}},
	})
}
