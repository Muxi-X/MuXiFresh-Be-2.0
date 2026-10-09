// Package migrate 提供一次性邮箱存量治理逻辑：把规范化后重复的账号收敛为唯一
// keeper，并摘除其余账号的邮箱、迁移其学号；随后把全部邮箱域名拍平为小写。
package migrate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sort"

	"MuXiFresh-Be-2.0/common/tool"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// Run 治理存量邮箱。apply=false 为 dry-run，只输出计划不改动。
//
// 返回错误仅代表数据库操作失败；业务层冲突（如 keeper 已绑定不同学号）以警告
// 输出、不中止。核心不变量：结束后同一邮箱（域名小写）至多对应一个账号。
func Run(ctx context.Context, client *mongo.Client, dbName string, apply bool, out io.Writer) error {
	userinfo := client.Database(dbName).Collection("userinfo")
	userauth := client.Database(dbName).Collection("userauth")

	// 先拍平域名，再处理重复：3 组重复仅域名大小写不同，拍平后才成为完全相同
	// 的 email，后续按规范化值分组可一并命中。
	if err := flattenDomains(ctx, userinfo, "userinfo", apply, out); err != nil {
		return err
	}
	if err := flattenDomains(ctx, userauth, "userauth", apply, out); err != nil {
		return err
	}
	if err := resolveDuplicates(ctx, userinfo, userauth, apply, out); err != nil {
		return err
	}

	if !apply {
		fmt.Fprintln(out, "[dry-run] 未做任何改动；加 -apply 才会写入。")
	}
	return nil
}

// flattenDomains 把集合中所有 email 的域名部分规范化为小写（本地部分与整串
// 已规范的保持不变）。
func flattenDomains(ctx context.Context, coll *mongo.Collection, name string, apply bool, out io.Writer) error {
	cur, err := coll.Find(ctx, bson.M{"email": bson.M{"$type": "string"}})
	if err != nil {
		return fmt.Errorf("列出 %s 邮箱: %w", name, err)
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var doc struct {
			ID    primitive.ObjectID `bson:"_id"`
			Email string             `bson:"email"`
		}
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		norm := tool.NormalizeEmail(doc.Email)
		if norm == doc.Email {
			continue
		}
		fmt.Fprintf(out, "[flatten] %s _id=%s %q -> %q\n", name, doc.ID.Hex(), doc.Email, norm)
		if apply {
			if _, err := coll.UpdateOne(ctx, bson.M{"_id": doc.ID}, bson.M{"$set": bson.M{"email": norm}}); err != nil {
				return fmt.Errorf("拍平 %s/%s: %w", name, doc.ID.Hex(), err)
			}
		}
	}
	return cur.Err()
}

// resolveDuplicates 对 userinfo 按规范化 email 分组，每组保留 _id 最大（较新）
// 的一条为 keeper；其余账号的 student_id 迁移到 keeper 后，摘除其 email。
func resolveDuplicates(ctx context.Context, userinfo, userauth *mongo.Collection, apply bool, out io.Writer) error {
	cur, err := userinfo.Find(ctx, bson.M{"email": bson.M{"$type": "string"}})
	if err != nil {
		return fmt.Errorf("列出 userinfo 邮箱: %w", err)
	}
	defer cur.Close(ctx)

	groups := map[string][]primitive.ObjectID{}
	var order []string
	for cur.Next(ctx) {
		var doc struct {
			ID    primitive.ObjectID `bson:"_id"`
			Email string             `bson:"email"`
		}
		if err := cur.Decode(&doc); err != nil {
			return err
		}
		key := tool.NormalizeEmail(doc.Email)
		if _, ok := groups[key]; !ok {
			order = append(order, key)
		}
		groups[key] = append(groups[key], doc.ID)
	}
	if err := cur.Err(); err != nil {
		return err
	}
	sort.Strings(order)

	for _, key := range order {
		ids := groups[key]
		if len(ids) < 2 {
			continue
		}
		// 较新 = ObjectID 较大（其高位含时间戳），降序取首为 keeper。
		sort.Slice(ids, func(i, j int) bool { return ids[i].Hex() > ids[j].Hex() })
		keeper := ids[0]
		fmt.Fprintf(out, "[duplicate] %q count=%d keeper=%s raw=%s\n", key, len(ids), keeper.Hex(), key)

		for _, id := range ids[1:] {
			if err := migrateStudentID(ctx, userinfo, keeper, id, apply, out); err != nil {
				return err
			}
			fmt.Fprintf(out, "  unset email: userinfo/%s, userauth(userInfoID=%s)\n", id.Hex(), id.Hex())
			if apply {
				if _, err := userinfo.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$unset": bson.M{"email": ""}}); err != nil {
					return fmt.Errorf("摘除 userinfo/%s 邮箱: %w", id.Hex(), err)
				}
				if _, err := userauth.UpdateMany(ctx, bson.M{"userInfoID": id}, bson.M{"$unset": bson.M{"email": ""}}); err != nil {
					return fmt.Errorf("摘除 userauth(userInfoID=%s) 邮箱: %w", id.Hex(), err)
				}
			}
		}
	}
	return nil
}

// migrateStudentID 把旧账号的学号迁移到 keeper：
//   - keeper 未绑定 → 写入 keeper；
//   - keeper 已绑同一学号 → 无需搬；
//   - keeper 已绑不同学号 → 保留 keeper 值并警告，不搬。
//
// 旧账号若无学号则无操作。
func migrateStudentID(ctx context.Context, userinfo *mongo.Collection, keeper, old primitive.ObjectID, apply bool, out io.Writer) error {
	var oldDoc struct {
		StudentID string `bson:"student_id"`
	}
	if err := userinfo.FindOne(ctx, bson.M{"_id": old}).Decode(&oldDoc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}
	if oldDoc.StudentID == "" {
		return nil
	}

	var keeperDoc struct {
		StudentID string `bson:"student_id"`
	}
	if err := userinfo.FindOne(ctx, bson.M{"_id": keeper}).Decode(&keeperDoc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}
		return err
	}

	switch {
	case keeperDoc.StudentID == "":
		fmt.Fprintf(out, "  migrate student_id %q: %s -> keeper %s (unset from old)\n", oldDoc.StudentID, old.Hex(), keeper.Hex())
		if apply {
			if _, err := userinfo.UpdateOne(ctx, bson.M{"_id": keeper}, bson.M{"$set": bson.M{"student_id": oldDoc.StudentID}}); err != nil {
				return fmt.Errorf("迁移学号到 keeper %s: %w", keeper.Hex(), err)
			}
			if _, err := userinfo.UpdateOne(ctx, bson.M{"_id": old}, bson.M{"$unset": bson.M{"student_id": ""}}); err != nil {
				return fmt.Errorf("摘除旧账号 %s 学号: %w", old.Hex(), err)
			}
		}
	case keeperDoc.StudentID == oldDoc.StudentID:
		fmt.Fprintf(out, "  keeper %s already bound to student_id %q, unset from old\n", keeper.Hex(), oldDoc.StudentID)
		if apply {
			if _, err := userinfo.UpdateOne(ctx, bson.M{"_id": old}, bson.M{"$unset": bson.M{"student_id": ""}}); err != nil {
				return fmt.Errorf("摘除旧账号 %s 学号: %w", old.Hex(), err)
			}
		}
	default:
		// 同人不同学号罕见：保留 keeper 值、旧账号原样不搬，留人工确认，避免
		// 单方面摘除导致无法回溯。
		fmt.Fprintf(out, "  [warn] keeper %s student_id=%q, old %s student_id=%q; keep both, manual review\n",
			keeper.Hex(), keeperDoc.StudentID, old.Hex(), oldDoc.StudentID)
	}
	return nil
}
