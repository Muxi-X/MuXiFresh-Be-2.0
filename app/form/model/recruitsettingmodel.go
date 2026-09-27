package model

import (
	"context"
	"errors"
	"time"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ RecruitSettingModel = (*customRecruitSettingModel)(nil)

type (
	// RecruitSettingModel 读写 recruit_setting（每届一条）。
	RecruitSettingModel interface {
		// GetByCycle 返回指定届次的配置；不存在返回 ErrNotFound。
		GetByCycle(ctx context.Context, cycle string) (*RecruitSetting, error)
		// SetDeadline 以乐观锁写入指定届次的截止时间。
		// expectedRev==0 表示首次写入（记录不存在则创建，存在则要求其 rev 也为 0）；
		// 其余情况要求现有 rev 与 expectedRev 相等，否则返回 ErrVersionConflict。
		// 变更记录与其值在同一次更新中原子落库；返回写入后的版本号。
		SetDeadline(ctx context.Context, cycle string, deadline time.Time, expectedRev int64, updateBy primitive.ObjectID, operatorType string) (int64, error)
	}

	customRecruitSettingModel struct {
		conn *mon.Model
	}
)

// NewRecruitSettingModel returns a model for the mongo.
func NewRecruitSettingModel(url, db, collection string) RecruitSettingModel {
	return &customRecruitSettingModel{conn: mon.MustNewModel(url, db, collection)}
}

func (m *customRecruitSettingModel) GetByCycle(ctx context.Context, cycle string) (*RecruitSetting, error) {
	var data RecruitSetting

	err := m.conn.FindOne(ctx, &data, bson.M{"cycle": cycle})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// SetDeadline 用显式 $set 只写 deadline 相关字段，过滤条件带上 rev 做 CAS，
// 命中才 $inc +1，从而拒绝基于旧版本的覆盖写；expectedRev==0 时允许 upsert 创建。
// 变更记录通过 $push 与值写入同一次原子更新，避免写入成功而审计缺失。
// cycle 上有唯一索引：并发首写时后到者会撞唯一键，归一为版本冲突而非插入重复届次。
func (m *customRecruitSettingModel) SetDeadline(ctx context.Context, cycle string, deadline time.Time, expectedRev int64, updateBy primitive.ObjectID, operatorType string) (int64, error) {
	var old *time.Time

	current, err := m.GetByCycle(ctx, cycle)
	switch {
	case err == nil:
		if current.Rev != expectedRev {
			return 0, ErrVersionConflict
		}
		d := current.Deadline
		old = &d
	case errors.Is(err, ErrNotFound):
		if expectedRev != 0 {
			return 0, ErrVersionConflict
		}
	default:
		return 0, err
	}

	newRev := expectedRev + 1
	now := time.Now()
	// expectedRev==0 时额外匹配「rev 字段缺失」的老文档——Mongo 的 {rev:0} 不匹配缺失字段。
	filter := bson.M{"cycle": cycle}
	if expectedRev == 0 {
		filter["$or"] = []bson.M{
			{"rev": int64(0)},
			{"rev": bson.M{"$exists": false}},
		}
	} else {
		filter["rev"] = expectedRev
	}

	res, err := m.conn.UpdateOne(ctx,
		filter,
		bson.M{
			"$set": bson.M{"deadline": deadline, "updateBy": updateBy, "updateAt": now},
			"$inc": bson.M{"rev": 1},
			"$push": bson.M{"history": RecruitSettingChange{
				OldDeadline:  old,
				NewDeadline:  deadline,
				Rev:          newRev,
				OperatorID:   updateBy,
				OperatorType: operatorType,
				OperatedAt:   now,
			}},
			"$setOnInsert": bson.M{"createAt": now},
		},
		options.Update().SetUpsert(expectedRev == 0))
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return 0, ErrVersionConflict
		}
		return 0, err
	}
	if res.MatchedCount == 0 && res.UpsertedCount == 0 {
		return 0, ErrVersionConflict
	}

	return newRev, nil
}
