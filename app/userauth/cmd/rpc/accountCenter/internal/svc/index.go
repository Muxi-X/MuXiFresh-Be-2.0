package svc

import (
	"context"
	"time"

	"MuXiFresh-Be-2.0/common/mongodb"

	"go.mongodb.org/mongo-driver/bson"
)

// EnsureAccountIndexes 确保账户身份两个集合的邮箱唯一约束（账户数据归属方
// accountCenter 负责）。邮箱在应用层已统一规范化（域名小写），故普通 unique
// 索引即可；sparse 允许"未设邮箱"的文档共存（如存量治理中被摘除邮箱的旧账号）。
//
// 创建失败即启动失败（fail closed）：静默失败会让同邮箱多账号继续产生，且无
// 任何外部迹象。
func EnsureAccountIndexes(url, db string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := mongodb.Connect(ctx, url)
	if err != nil {
		return err
	}
	defer client.Disconnect(context.Background())

	specs := []mongodb.IndexSpec{
		{
			Collection: "userinfo",
			Name:       "userinfo_email_unique",
			Unique:     true,
			Sparse:     true,
			Keys:       bson.D{{Key: "email", Value: 1}},
		},
		{
			Collection: "userauth",
			Name:       "userauth_email_unique",
			Unique:     true,
			Sparse:     true,
			Keys:       bson.D{{Key: "email", Value: 1}},
		},
	}
	if err := mongodb.EnsureIndexes(ctx, client, db, specs...); err != nil {
		return err
	}

	// 历史遗留的 `email_1`（非 sparse unique）会把多条"无邮箱"记录视为同一 null
	// 而互相冲突。新索引建好后删除它，按名+键精确匹配，不误删其他索引。
	for _, coll := range []string{"userinfo", "userauth"} {
		if err := mongodb.DropIndexByKey(ctx, client, db, coll, "email_1",
			bson.D{{Key: "email", Value: 1}}); err != nil {
			return err
		}
	}
	return nil
}
