package svc

import (
	"context"
	"time"

	"MuXiFresh-Be-2.0/common/mongodb"

	"go.mongodb.org/mongo-driver/bson"
)

// EnsureScheduleIndexes 确保 schedule 的索引（由 schedule 服务负责，数据归属方）。
// user_id 唯一索引用于兜底 UpsertByUserId 的并发双写。
func EnsureScheduleIndexes(url, db string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongodb.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	return mongodb.EnsureIndex(ctx, client, db, mongodb.IndexSpec{
		Collection: "schedule",
		Name:       "schedule_user_id",
		Unique:     true,
		Keys:       bson.D{{Key: "user_id", Value: 1}},
	})
}
