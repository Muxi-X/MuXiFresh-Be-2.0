package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

var _ EntryFormModel = (*customEntryFormModel)(nil)

type (
	// EntryFormModel is an interface to be customized, add more methods here,
	// and implement the added methods in customEntryFormModel.
	EntryFormModel interface {
		entryFormModel
		InsertReturnID(ctx context.Context, data *EntryForm) (interface{}, error)
		// FindOneByUserId 返回该用户最新一届（createAt 最大）的报名表。
		// 同一用户跨届可存在多份表，故必须排序后再取，否则可能取到往届。
		FindOneByUserId(ctx context.Context, userId string) (*EntryForm, error)
		// FindByUserIdAndCycle 返回指定用户在指定届次的报名表；无则 ErrNotFound。
		FindByUserIdAndCycle(ctx context.Context, userId, cycle string) (*EntryForm, error)
		FindByGroup(ctx context.Context, group string, school string, grade string, startDate time.Time, endDate time.Time) ([]*EntryForm, error)
		// SetInterviewComment 以乐观锁方式只更新面评字段与版本号，不触碰报名表其它字段；
		// 空串可清空面评。仅当当前版本等于 expectedRev 时才写入，命中 0 条表示版本冲突或文档不存在。
		SetInterviewComment(ctx context.Context, formID string, comment string, expectedRev int64) (*mongo.UpdateResult, error)
	}

	customEntryFormModel struct {
		*defaultEntryFormModel
	}
)

// NewEntryFormModel returns a model for the mongo.
func NewEntryFormModel(url, db, collection string) EntryFormModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customEntryFormModel{
		defaultEntryFormModel: newDefaultEntryFormModel(conn),
	}
}

func (m *defaultEntryFormModel) InsertReturnID(ctx context.Context, data *EntryForm) (interface{}, error) {
	if data.ID.IsZero() {
		now := time.Now()
		data.ID = primitive.NewObjectID()
		data.CreateAt = now
		data.UpdateAt = now
		// 届次与 CreateAt 必须同一时刻推导，否则跨越 7/1 分界时二者会不一致
		data.Cycle = CycleOf(now)
	}

	id, err := m.conn.InsertOne(ctx, data)
	if err != nil {
		return nil, err
	}
	return id.InsertedID, nil
}

func (m *customEntryFormModel) FindOneByUserId(ctx context.Context, userId string) (*EntryForm, error) {
	oid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	var data EntryForm

	// 跨届多表时取最新一届（createAt 倒序），保证作业组别校验/判交表口径基于当届
	err = m.conn.FindOne(ctx, &data, bson.M{"user_id": oid},
		options.FindOne().SetSort(bson.D{{Key: "createAt", Value: -1}}))
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByUserIdAndCycle 返回指定用户在指定届次的报名表；无则 ErrNotFound。
func (m *customEntryFormModel) FindByUserIdAndCycle(ctx context.Context, userId, cycle string) (*EntryForm, error) {
	oid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	var data EntryForm

	err = m.conn.FindOne(ctx, &data, bson.M{"user_id": oid, "cycle": cycle})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// FindByGroup 按时间窗口与可选条件查询报名表。
func (m *customEntryFormModel) FindByGroup(ctx context.Context, group string, school string, grade string, startDate time.Time, endDate time.Time) ([]*EntryForm, error) {
	var entryForms []*EntryForm
	filter := bson.D{}
	//必选
	filter = append(filter, bson.E{Key: "createAt", Value: bson.M{
		"$gte": startDate, // 大于等于起始时间
		"$lte": endDate,   // 小于等于结束时间
	}})
	//可选项
	if school != "" {
		filter = append(filter, bson.E{Key: "school", Value: school})
	}
	if grade != "" {
		filter = append(filter, bson.E{Key: "grade", Value: grade})
	}
	if group != "" {
		filter = append(filter, bson.E{Key: "group", Value: group})
	}
	err := m.conn.Find(ctx, &entryForms, filter, options.Find().SetSort(bson.D{{"name", 1}}))

	switch err {
	case nil:
		return entryForms, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

// SetInterviewComment 用显式 $set 只写面评字段，避免复用 Update 时把整个结构体
// 写回而误触其它字段；用 bson.M 而非结构体，空串也会被写入，所以可以清空面评。
// 过滤条件带上 interviewCommentRev 做 CAS，命中才 +1，从而拒绝基于旧版本的覆盖写。
// expectedRev 为 0 时要额外匹配「字段不存在」的老文档——Mongo 的 {field: 0} 不匹配缺失字段。
func (m *customEntryFormModel) SetInterviewComment(ctx context.Context, formID string, comment string, expectedRev int64) (*mongo.UpdateResult, error) {
	oid, err := primitive.ObjectIDFromHex(formID)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	filter := bson.M{"_id": oid}
	if expectedRev == 0 {
		filter["$or"] = []bson.M{
			{"interviewCommentRev": int64(0)},
			{"interviewCommentRev": bson.M{"$exists": false}},
		}
	} else {
		filter["interviewCommentRev"] = expectedRev
	}

	return m.conn.UpdateOne(ctx, filter,
		bson.M{
			"$set": bson.M{"interviewComment": comment},
			"$inc": bson.M{"interviewCommentRev": 1},
		})
}
