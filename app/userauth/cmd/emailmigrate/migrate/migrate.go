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
// 返回错误代表数据库操作失败，或治理后仍存在规范化重复（末尾只读自检失败）——
// 后者绝不能静默放过，否则 accountCenter 启动建唯一索引会 panic。
//
// 两个集合各自收敛重复，且每个非 keeper 账号都先摘 userauth 再摘 userinfo：
// 中途失败后重跑，两集合的扫描都能独立发现并继续收敛，命令是幂等且可自愈的。
func Run(ctx context.Context, client *mongo.Client, dbName string, apply bool, out io.Writer) error {
	userinfo := client.Database(dbName).Collection("userinfo")
	userauth := client.Database(dbName).Collection("userauth")

	if err := flattenDomains(ctx, userinfo, "userinfo", apply, out); err != nil {
		return err
	}
	if err := flattenDomains(ctx, userauth, "userauth", apply, out); err != nil {
		return err
	}
	if err := resolveUserinfoDuplicates(ctx, userinfo, userauth, apply, out); err != nil {
		return err
	}
	if err := resolveUserauthDuplicates(ctx, userinfo, userauth, apply, out); err != nil {
		return err
	}

	if apply {
		if err := verifyNoDuplicates(ctx, userinfo, userauth, out); err != nil {
			return err
		}
	} else {
		fmt.Fprintln(out, "[dry-run] 未做任何改动；加 -apply 才会写入。")
	}
	return nil
}

// scanEmailGroups 返回集合中按规范化 email 分组的 account（_id），只含 email 为
// 非空字符串的文档。
func scanEmailGroups(ctx context.Context, coll *mongo.Collection) (map[string][]primitive.ObjectID, error) {
	cur, err := coll.Find(ctx, bson.M{"email": bson.M{"$type": "string"}})
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	groups := map[string][]primitive.ObjectID{}
	for cur.Next(ctx) {
		var doc struct {
			ID    primitive.ObjectID `bson:"_id"`
			Email string             `bson:"email"`
		}
		if err := cur.Decode(&doc); err != nil {
			return nil, err
		}
		key := tool.NormalizeEmail(doc.Email)
		if key == "" {
			continue
		}
		groups[key] = append(groups[key], doc.ID)
	}
	if err := cur.Err(); err != nil {
		return nil, err
	}
	return groups, nil
}

// newerFirst 降序排序：ObjectID 越大越新（高位含时间戳），首元素为 keeper。
func newerFirst(ids []primitive.ObjectID) {
	sort.Slice(ids, func(i, j int) bool { return ids[i].Hex() > ids[j].Hex() })
}

// resolveUserinfoDuplicates 以 userinfo 分组为准收敛重复：keeper 取较新账号；
// 非 keeper 的学号迁到 keeper 后，先摘 userauth 邮箱、再摘 userinfo 邮箱。
func resolveUserinfoDuplicates(ctx context.Context, userinfo, userauth *mongo.Collection, apply bool, out io.Writer) error {
	groups, err := scanEmailGroups(ctx, userinfo)
	if err != nil {
		return fmt.Errorf("列出 userinfo 邮箱: %w", err)
	}

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ids := groups[key]
		if len(ids) < 2 {
			continue
		}
		newerFirst(ids)
		keeper := ids[0]
		fmt.Fprintf(out, "[duplicate] userinfo %q count=%d keeper=%s\n", key, len(ids), keeper.Hex())

		for _, old := range ids[1:] {
			if err := migrateStudentID(ctx, userinfo, keeper, old, apply, out); err != nil {
				return err
			}
			if err := unsetEmails(ctx, userinfo, userauth, old, apply, out); err != nil {
				return err
			}
		}
	}
	return nil
}

// resolveUserauthDuplicates 收敛 userauth 自身的规范化重复（如孤儿 userauth、或
// 与 userinfo 不一致的数据）。keeper 优先取"其 userinfo 仍持有该邮箱"的账号，
// 否则取较新一条；其余摘除 userauth 邮箱。此步保证 userauth 唯一索引也能建起来。
func resolveUserauthDuplicates(ctx context.Context, userinfo, userauth *mongo.Collection, apply bool, out io.Writer) error {
	groups, err := scanEmailGroups(ctx, userauth)
	if err != nil {
		return fmt.Errorf("列出 userauth 邮箱: %w", err)
	}

	keys := make([]string, 0, len(groups))
	for k := range groups {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for _, key := range keys {
		ids := groups[key]
		if len(ids) < 2 {
			continue
		}
		newerFirst(ids)
		keeper := ids[0]
		// 优先保留 userinfo 侧仍是该邮箱的账号，避免 keeper 与 userinfo 错位。
		for _, id := range ids {
			owner, err := userinfoEmailOwner(ctx, userinfo, id)
			if err != nil {
				return err
			}
			if owner == key {
				keeper = id
				break
			}
		}
		fmt.Fprintf(out, "[duplicate] userauth %q count=%d keeper=%s\n", key, len(ids), keeper.Hex())

		for _, id := range ids {
			if id == keeper {
				continue
			}
			fmt.Fprintf(out, "  unset email: userauth _id=%s\n", id.Hex())
			if apply {
				if _, err := userauth.UpdateOne(ctx, bson.M{"_id": id}, bson.M{"$unset": bson.M{"email": ""}}); err != nil {
					return fmt.Errorf("摘除 userauth/%s 邮箱: %w", id.Hex(), err)
				}
			}
		}
	}
	return nil
}

// userinfoEmailOwner 返回 userinfo 中该账号当前的规范化邮箱（无该文档或无邮箱
// 返回空串）。
func userinfoEmailOwner(ctx context.Context, userinfo *mongo.Collection, id primitive.ObjectID) (string, error) {
	var doc struct {
		Email string `bson:"email"`
	}
	if err := userinfo.FindOne(ctx, bson.M{"_id": id}).Decode(&doc); err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		return "", err
	}
	return tool.NormalizeEmail(doc.Email), nil
}

// unsetEmails 摘除非 keeper 账号的邮箱：先 userauth（可自愈，失败重跑时 userinfo
// 分组仍能发现残留），后 userinfo。
func unsetEmails(ctx context.Context, userinfo, userauth *mongo.Collection, old primitive.ObjectID, apply bool, out io.Writer) error {
	fmt.Fprintf(out, "  unset email: userauth(userInfoID=%s), userinfo/%s\n", old.Hex(), old.Hex())
	if !apply {
		return nil
	}
	if _, err := userauth.UpdateMany(ctx, bson.M{"userInfoID": old}, bson.M{"$unset": bson.M{"email": ""}}); err != nil {
		return fmt.Errorf("摘除 userauth(userInfoID=%s) 邮箱: %w", old.Hex(), err)
	}
	if _, err := userinfo.UpdateOne(ctx, bson.M{"_id": old}, bson.M{"$unset": bson.M{"email": ""}}); err != nil {
		return fmt.Errorf("摘除 userinfo/%s 邮箱: %w", old.Hex(), err)
	}
	return nil
}

// verifyNoDuplicates 治理后只读自检：两集合任一仍有规范化重复即报错，阻止在
// 未清理干净时盲目重启服务（建唯一索引会 panic）。
func verifyNoDuplicates(ctx context.Context, userinfo, userauth *mongo.Collection, out io.Writer) error {
	for _, c := range []struct {
		name string
		coll *mongo.Collection
	}{{"userinfo", userinfo}, {"userauth", userauth}} {
		// 显式存储的空串不被 scanEmailGroups 覆盖，但会参与 sparse 唯一索引，
		// 多条空串仍会互相冲突，故单独拒绝。
		emptyCount, err := c.coll.CountDocuments(ctx, bson.M{"email": ""})
		if err != nil {
			return err
		}
		if emptyCount > 1 {
			return fmt.Errorf("治理后 %s 仍有空邮箱字段 (count=%d)，请人工排查", c.name, emptyCount)
		}
		groups, err := scanEmailGroups(ctx, c.coll)
		if err != nil {
			return err
		}
		for key, ids := range groups {
			if len(ids) > 1 {
				return fmt.Errorf("治理后 %s 仍有规范化重复邮箱 %q (count=%d)，请人工排查", c.name, key, len(ids))
			}
		}
	}
	fmt.Fprintln(out, "[verify] 两集合均无规范化重复邮箱。")
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
		if norm == "" {
			// 纯空白会被规范化为空串；空串在 sparse 唯一索引下仍参与唯一性，
			// 多条空串会互相冲突。故直接摘除字段（等价于"未设邮箱"）。
			fmt.Fprintf(out, "[flatten] %s _id=%s empty email -> unset\n", name, doc.ID.Hex())
			if apply {
				if _, err := coll.UpdateOne(ctx, bson.M{"_id": doc.ID}, bson.M{"$unset": bson.M{"email": ""}}); err != nil {
					return fmt.Errorf("摘除 %s/%s 空邮箱: %w", name, doc.ID.Hex(), err)
				}
			}
			continue
		}
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
