package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
		FindOneByUserId(ctx context.Context, userId string) (*EntryForm, error)
		FindByGroup(ctx context.Context, group string, school string, grade string, startDate time.Time, endDate time.Time) ([]*EntryForm, error)
		CountByGroup(ctx context.Context, group string, startDate time.Time, endDate time.Time) (int64, error)
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
		data.ID = primitive.NewObjectID()
		data.CreateAt = time.Now()
		data.UpdateAt = time.Now()
	}

	id, err := m.conn.InsertOne(ctx, data)
	return id.InsertedID, err
}

func (m *customEntryFormModel) FindOneByUserId(ctx context.Context, userId string) (*EntryForm, error) {
	oid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, ErrInvalidObjectId
	}

	var data EntryForm

	err = m.conn.FindOne(ctx, &data, bson.M{"user_id": oid})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

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

func (m *customEntryFormModel) CountByGroup(ctx context.Context, group string, startDate time.Time, endDate time.Time) (int64, error) {
	number, err := m.conn.CountDocuments(ctx, bson.M{
		"group": group,
		"createAt": bson.M{
			"$gte": startDate, // 大于等于起始时间
			"$lte": endDate,   // 小于等于结束时间
		},
	})
	if err != nil {
		return 0, err
	}
	return number, err
}
