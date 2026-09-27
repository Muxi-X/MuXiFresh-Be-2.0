package svc

import (
	"context"
	"time"

	"MuXiFresh-Be-2.0/common/mongodb"

	"go.mongodb.org/mongo-driver/bson"
)

const recruitSettingCollection = "recruit_setting"

// EnsureRecruitSettingIndexes 确保 recruit_setting 的 cycle 唯一约束。
// 乐观锁首写用 upsert，并发下后到者会撞唯一键（归一为版本冲突），
// 没有该约束则可能插入重复届次。索引是库级对象，建一次即全局生效。
func EnsureRecruitSettingIndexes(url, db string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongodb.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	return mongodb.EnsureIndex(ctx, client, db, mongodb.IndexSpec{
		Collection: recruitSettingCollection,
		Name:       "recruit_setting_cycle",
		Unique:     true,
		Keys:       bson.D{{Key: "cycle", Value: 1}},
	})
}
