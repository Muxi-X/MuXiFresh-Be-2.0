package model

import (
	"MuXiFresh-Be-2.0/common/globalKey"
	"context"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var _ ScheduleModel = (*customScheduleModel)(nil)

type (
	// ScheduleModel is an interface to be customized, add more methods here,
	// and implement the added methods in customScheduleModel.
	ScheduleModel interface {
		scheduleModel
		FindOneByUserId(ctx context.Context, userId string) (*Schedule, error)
		FindByUserIds(ctx context.Context, userIds []string) ([]*Schedule, error)
		UpdateByUserId(ctx context.Context, data *Schedule) (*mongo.UpdateResult, error)
		UpsertByUserId(ctx context.Context, data *Schedule) (*mongo.UpdateResult, error)
		InsertGetID(ctx context.Context, data *Schedule) (string, error)
	}

	customScheduleModel struct {
		*defaultScheduleModel
	}
)

// NewScheduleModel returns a model for the mongo.
func NewScheduleModel(url, db, collection string) ScheduleModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customScheduleModel{
		defaultScheduleModel: newDefaultScheduleModel(conn),
	}
}

func (m *customScheduleModel) FindOneByUserId(ctx context.Context, userId string) (*Schedule, error) {
	uid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	var data Schedule

	err = m.conn.FindOne(ctx, &data, bson.M{"user_id": uid})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customScheduleModel) FindByUserIds(ctx context.Context, userIds []string) ([]*Schedule, error) {
	oids := make([]primitive.ObjectID, 0, len(userIds))
	for _, id := range userIds {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, ErrInvalidObjectId
		}
		oids = append(oids, oid)
	}

	var schedules []*Schedule
	err := m.conn.Find(ctx, &schedules, bson.M{"user_id": bson.M{"$in": oids}})
	switch err {
	case nil:
		return schedules, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultScheduleModel) UpdateByUserId(ctx context.Context, data *Schedule) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()

	res, err := m.conn.UpdateOne(ctx, bson.M{"user_id": data.UserID}, bson.M{"$set": data})
	return res, err
}

// UpsertByUserId 原子地按 user_id 插入或更新 schedule，
// 配合 user_id 唯一索引可避免并发下的双写（需在 MongoDB 建 unique 索引）。
//
// 匹配条件排除已录取状态（实习期/已转正）：这类成员的录取状态归管理员所有，
// 报名流程不得改写。守卫是入口预检，存在"查完→管理员改状态→开始写"的并发窗口，
// 故必须在此用过滤条件兜底。
//
// 三种结局：
//   - 无记录 / 状态为未报名、已报名 → 命中并写入（未报名被升级为已报名），
//     无记录时由 upsert 插入；
//   - 状态为实习期/已转正 → 过滤不命中，upsert 尝试插入却撞 user_id 唯一索引，
//     返回重复键错误。调用方需读回判别：录取态拒绝、并发插入放行。
//
// 更新只刷新状态与 updateAt，createAt/user_id 仅插入时写入（$setOnInsert）。
func (m *defaultScheduleModel) UpsertByUserId(ctx context.Context, data *Schedule) (*mongo.UpdateResult, error) {
	now := time.Now()

	res, err := m.conn.UpdateOne(ctx,
		bson.M{
			"user_id": data.UserID,
			"admission_status": bson.M{
				"$nin": []string{globalKey.Internship, globalKey.Formal},
			},
		},
		bson.M{
			"$set": bson.M{
				"entry_form_status": data.EntryFormStatus,
				"admission_status":  data.AdmissionStatus,
				"updateAt":          now,
			},
			"$setOnInsert": bson.M{
				"user_id":  data.UserID,
				"createAt": now,
			},
		},
		options.Update().SetUpsert(true))
	return res, err
}

func (m *defaultScheduleModel) InsertGetID(ctx context.Context, data *Schedule) (string, error) {
	if data.ID.IsZero() {
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	result, err := m.conn.InsertOne(ctx, data)
	if err != nil {
		return "", err
	}
	return fmt.Sprint(result.InsertedID), nil
}
