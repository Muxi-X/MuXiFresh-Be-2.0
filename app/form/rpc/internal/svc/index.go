package svc

import (
	"context"
	"fmt"
	"time"

	"MuXiFresh-Be-2.0/common/mongodb"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

const entryFormCollection = "entry_form"

// EnsureEntryFormIndexes 确保 entry_form 的索引与唯一约束（由 form 服务负责，
// 因为 entry_form 的写方是 form）。索引是库级对象，任意进程建一次即全局生效。
//
// 跨届重报要求唯一约束从 user_id 改为 (user_id, cycle)：
//   - 先建新唯一索引，确认建好后才删旧 user_id_1。绝不"先删旧再建新"：新索引若因
//     存量重复键建不起来，先删旧会让 entry_form 彻底失去唯一约束，并发下产生同届
//     重复表且再无 DB 兜底。建不起来则保留旧索引并返回错误（fail safe）。
//   - 旧的 user_id_1 会把同一用户的往届表与本届重报表判为冲突，必须删除。
func EnsureEntryFormIndexes(url, db string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongodb.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	// review 组合查询使用：createAt 前缀 + group/school/grade 可选过滤
	if err := mongodb.EnsureIndex(ctx, client, db, mongodb.IndexSpec{
		Collection: entryFormCollection,
		Name:       "entry_form_createAt_group_school_grade",
		Keys: bson.D{{Key: "createAt", Value: 1}, {Key: "group", Value: 1},
			{Key: "school", Value: 1}, {Key: "grade", Value: 1}},
	}); err != nil {
		return err
	}

	if err := mongodb.EnsureIndex(ctx, client, db, mongodb.IndexSpec{
		Collection: entryFormCollection,
		Name:       "entry_form_user_id_cycle",
		Unique:     true,
		Keys:       bson.D{{Key: "user_id", Value: 1}, {Key: "cycle", Value: 1}},
	}); err != nil {
		return fmt.Errorf("entry_form_user_id_cycle unavailable, keeping legacy index: %w", err)
	}

	return dropLegacyEntryFormUserIndex(ctx, client, db)
}

// dropLegacyEntryFormUserIndex 删除 entry_form 的历史 user_id 唯一索引（名为 user_id_1，
// 键为 {user_id: 1}），它是早期手工建立的。仅在键确实匹配时才删；索引不存在时空操作。
func dropLegacyEntryFormUserIndex(ctx context.Context, client *mongo.Client, db string) error {
	return mongodb.DropIndexByKey(ctx, client, db, entryFormCollection, "user_id_1",
		bson.D{{Key: "user_id", Value: 1}})
}
