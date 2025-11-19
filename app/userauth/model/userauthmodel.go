package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"time"
)

var _ UserAuthModel = (*customUserAuthModel)(nil)

type (
	// UserAuthModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserAuthModel.
	UserAuthModel interface {
		userAuthModel
		FindOneByEmailAndPassword(ctx context.Context, Email string, Password string) (*UserAuth, error)
		UpdateByEmail(ctx context.Context, data *UserAuth) (*mongo.UpdateResult, error)
		UpdateByUserId(ctx context.Context, data *UserAuth) (*mongo.UpdateResult, error)
	}

	customUserAuthModel struct {
		*defaultUserAuthModel
	}
)

// NewUserAuthModel returns a model for the mongo.
func NewUserAuthModel(url, db, collection string) UserAuthModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customUserAuthModel{
		defaultUserAuthModel: newDefaultUserAuthModel(conn),
	}
}

func (m *customUserAuthModel) UpdateByUserId(ctx context.Context, data *UserAuth) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()

	res, err := m.conn.UpdateOne(ctx, bson.M{"userInfoID": data.UserInfoID}, bson.M{"$set": data})

	return res, err
}

func (m *customUserAuthModel) FindOneByEmailAndPassword(ctx context.Context, Email string, Password string) (*UserAuth, error) {
	var data UserAuth

	err := m.conn.FindOne(ctx, &data, bson.M{"email": Email, "password": Password})

	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserAuthModel) UpdateByEmail(ctx context.Context, data *UserAuth) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()

	res, err := m.conn.UpdateOne(ctx, bson.M{"email": data.Email}, bson.M{"$set": data})

	return res, err
}
