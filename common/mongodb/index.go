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
// 判定"已存在"必须比对**名字 + 键（含顺序）+ unique**，不能只比名字：
// 若同名索引的键或唯一性不符，直接跳过会让调用方误以为约束已建立，
// 进而在后续步骤（如删除旧的唯一索引）中丢掉唯一性保护。
func EnsureIndex(ctx context.Context, client *mongo.Client, db string, spec IndexSpec) error {
	coll := client.Database(db).Collection(spec.Collection)

	existing, err := findIndex(ctx, coll, spec.Name)
	if err != nil {
		return fmt.Errorf("list index %s: %w", spec.Name, err)
	}
	if existing != nil {
		if indexMatches(existing, spec) {
			logx.Infof("index %s already exists, skip", spec.Name)
			return nil
		}
		return fmt.Errorf("index %s exists with different definition (keys=%v unique=%v), want (keys=%v unique=%v)",
			spec.Name, existing.Key, existing.Unique, spec.Keys, spec.Unique)
	}

	opts := options.Index().SetName(spec.Name)
	if spec.Unique {
		opts = opts.SetUnique(true)
	}
	if spec.Sparse {
		opts = opts.SetSparse(true)
	}

	if _, err := coll.Indexes().CreateOne(ctx, mongo.IndexModel{Keys: spec.Keys, Options: opts}); err != nil {
		// 并发建索引等情况：重新核对是否为完全一致的索引，仅精确匹配才算成功
		if existing, lookErr := findIndex(ctx, coll, spec.Name); lookErr == nil && existing != nil && indexMatches(existing, spec) {
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

	info, err := findIndex(ctx, coll, name)
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

// findIndex 按名查找集合索引；不存在返回 (nil, nil)。
func findIndex(ctx context.Context, coll *mongo.Collection, name string) (*indexInfo, error) {
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

// indexMatches 报告实际索引是否与期望完全一致。
//
// 比较范围：名字 + 键及顺序 + unique + sparse + partialFilterExpression。
// 后三项决定唯一性约束覆盖哪些文档，任一不符都必须判为不一致——
// 否则同名但作用域不同的索引会被接受，调用方继而删除旧索引，导致约束名存实亡。
// （期望侧只声明 sparse；partialFilterExpression 期望为"不设"，故实际存在即视为不符。）
func indexMatches(info *indexInfo, spec IndexSpec) bool {
	return info.Name == spec.Name &&
		info.Unique == spec.Unique &&
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
