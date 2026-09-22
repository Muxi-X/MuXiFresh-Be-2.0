package mongodb

import (
	"context"
	"fmt"

	"github.com/zeromicro/go-zero/core/logx"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// IndexSpec 描述一个待确保的索引。
type IndexSpec struct {
	Collection string
	Name       string
	Unique     bool
	Sparse     bool
	Keys       bson.D
}

// indexInfo 是集合索引的实际定义；Key 用 bson.D 保序解码，顺序是索引语义的一部分。
// sparse / partialFilterExpression 会改变索引覆盖的文档范围，进而影响唯一性约束的
// 实际作用域，比对时必须纳入，否则同名的 sparse/partial 索引会被误判为等价。
type indexInfo struct {
	Name                    string `bson:"name"`
	Key                     bson.D `bson:"key"`
	Unique                  bool   `bson:"unique"`
	Sparse                  bool   `bson:"sparse"`
	PartialFilterExpression bson.M `bson:"partialFilterExpression"`
}

// Connect 建立 MongoDB 连接；调用方负责 Disconnect。
func Connect(ctx context.Context, url string) (*mongo.Client, error) {
	return mongo.Connect(ctx, options.Client().ApplyURI(url))
}

// EnsureIndexes 幂等地创建 specs 中声明的索引。
// 索引是库级对象，任意进程建一次即全局生效，重复执行安全。
func EnsureIndexes(ctx context.Context, client *mongo.Client, db string, specs ...IndexSpec) error {
	for _, spec := range specs {
		if err := EnsureIndex(ctx, client, db, spec); err != nil {
			return err
		}
	}
	return nil
}

// EnsureIndex 创建单个索引。
//
// 判定"已存在"比对的是**索引定义**（键及顺序 + unique + sparse + partial），
// 而不是索引名。名字只是标签：同名不同定义、或不同名相同定义都很常见
// （如历史手工建的 user_id_1 与期望的 schedule_user_id 定义完全相同）。
// 若按名判定，不同名同定义的索引会走 CreateOne 并因 IndexOptionsConflict 失败；
// 同名不同定义则会被误当成满足约束。
func EnsureIndex(ctx context.Context, client *mongo.Client, db string, spec IndexSpec) error {
	coll := client.Database(db).Collection(spec.Collection)

	existing, err := findIndexByDefinition(ctx, coll, spec)
	if err != nil {
		return fmt.Errorf("list index %s: %w", spec.Name, err)
	}
	if existing != nil {
		if existing.Name != spec.Name {
			// 定义一致但名字不同：等价于约束已建立，不重复建（Mongo 会报
			// IndexOptionsConflict），也不改名——改名需 drop+create，属运维动作。
			logx.Infof("index %s already present under name %q, skip", spec.Name, existing.Name)
		} else {
			logx.Infof("index %s already exists, skip", spec.Name)
		}
		return nil
	}

	opts := options.Index().SetName(spec.Name)
	if spec.Unique {
		opts = opts.SetUnique(true)
	}
	if spec.Sparse {
		opts = opts.SetSparse(true)
	}

	if _, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: spec.Keys, Options: opts}); err != nil {
		// 并发建索引等情况：重新按定义核对，命中才算成功
		if found, lookErr := findIndexByDefinition(ctx, coll, spec); lookErr == nil && found != nil {
			logx.Infof("index %s already exists, skip", spec.Name)
			return nil
		}
		return fmt.Errorf("create index %s: %w", spec.Name, err)
	}
	logx.Infof("index %s created", spec.Name)
	return nil
}

// DropIndexByKey 按名删除索引，但仅当其键与 want 完全一致（含顺序）时才删；
// 名字命中而键不同（可能被他人复用该名）不擅自删除，仅记日志。
func DropIndexByKey(ctx context.Context, client *mongo.Client, db, collection, name string, want bson.D) error {
	coll := client.Database(db).Collection(collection)

	info, err := findIndexByName(ctx, coll, name)
	if err != nil {
		return fmt.Errorf("list index %s: %w", name, err)
	}
	if info == nil {
		return nil
	}
	if !keysMatch(info.Key, want) {
		logx.Errorf("index %s has unexpected keys %v, skip drop", name, info.Key)
		return nil
	}
	if _, err := coll.Indexes().DropOne(ctx, name); err != nil {
		return fmt.Errorf("drop index %s: %w", name, err)
	}
	logx.Infof("index %s dropped", name)
	return nil
}

// findIndexByName 按名查找集合索引；不存在返回 (nil, nil)。
// 仅用于按名删除（删除必须指名，不能按键推断）。
func findIndexByName(ctx context.Context, coll *mongo.Collection, name string) (*indexInfo, error) {
	cur, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var info indexInfo
		if err := cur.Decode(&info); err != nil {
			return nil, err
		}
		if info.Name == name {
			return &info, nil
		}
	}
	return nil, cur.Err()
}

// findIndexByDefinition 按定义（键及顺序 + unique + sparse + partial）查找索引；
// 不存在返回 (nil, nil)。名字不参与匹配。
func findIndexByDefinition(ctx context.Context, coll *mongo.Collection, spec IndexSpec) (*indexInfo, error) {
	cur, err := coll.Indexes().List(ctx)
	if err != nil {
		return nil, err
	}
	defer cur.Close(ctx)

	for cur.Next(ctx) {
		var info indexInfo
		if err := cur.Decode(&info); err != nil {
			return nil, err
		}
		if indexMatches(&info, spec) {
			return &info, nil
		}
	}
	return nil, cur.Err()
}

// indexMatches 报告实际索引定义是否满足期望。
//
// 比较范围：键及顺序 + unique + sparse + partialFilterExpression。**名字不参与**——
// 名字只是标签，约束的语义完全由定义决定。partialFilterExpression 会改变唯一约束
// 覆盖的文档范围，期望侧不设该选项，故实际存在即视为不符（避免把作用域更窄的索引
// 误认为已满足约束，进而删除旧索引导致唯一性名存实亡）。
func indexMatches(info *indexInfo, spec IndexSpec) bool {
	return info.Unique == spec.Unique &&
		info.Sparse == spec.Sparse &&
		len(info.PartialFilterExpression) == 0 &&
		keysMatch(info.Key, spec.Keys)
}

// keysMatch 比较两个索引键定义，顺序敏感；数值按整型归一（驱动可能解出 int32/int64）。
func keysMatch(got, want bson.D) bool {
	if len(got) != len(want) {
		return false
	}
	for i := range want {
		if got[i].Key != want[i].Key || !numericEqual(got[i].Value, want[i].Value) {
			return false
		}
	}
	return true
}

func numericEqual(got, want interface{}) bool {
	g, ok1 := toInt64(got)
	w, ok2 := toInt64(want)
	if ok1 && ok2 {
		return g == w
	}
	return got == want
}

func toInt64(v interface{}) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int32:
		return int64(n), true
	case int64:
		return n, true
	case float32:
		return int64(n), true
	case float64:
		return int64(n), true
	default:
		return 0, false
	}
}
