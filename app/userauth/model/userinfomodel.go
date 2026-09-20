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

var _ UserInfoModel = (*customUserInfoModel)(nil)

type (
	// UserInfoModel is an interface to be customized, add more methods here,
	// and implement the added methods in customUserInfoModel.
	UserInfoModel interface {
		userInfoModel
		FindByStudentID(ctx context.Context, studentID string) (*UserInfo, error)
		FindByUserIds(ctx context.Context, userIds []string) ([]*UserInfo, error)
		FindByEmail(ctx context.Context, email string) (*UserInfo, error)
		UpdateByEmail(ctx context.Context, data *UserInfo) (*mongo.UpdateResult, error)
		FindByUserType(ctx context.Context, userType string) ([]*UserInfo, error)
	}

	customUserInfoModel struct {
		*defaultUserInfoModel
	}
)

// NewUserInfoModel returns a model for the mongo.
func NewUserInfoModel(url, db, collection string) UserInfoModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customUserInfoModel{
		defaultUserInfoModel: newDefaultUserInfoModel(conn),
	}
}

func (m *defaultUserInfoModel) FindByStudentID(ctx context.Context, studentID string) (*UserInfo, error) {
	var data UserInfo

	err := m.conn.FindOne(ctx, &data, bson.M{"student_id": studentID})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultUserInfoModel) FindByEmail(ctx context.Context, email string) (*UserInfo, error) {
	var data UserInfo

	err := m.conn.FindOne(ctx, &data, bson.M{"email": email})
	switch err {
	case nil:
		return &data, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customUserInfoModel) FindByUserIds(ctx context.Context, userIds []string) ([]*UserInfo, error) {
	oids := make([]primitive.ObjectID, 0, len(userIds))
	for _, id := range userIds {
		oid, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return nil, ErrInvalidObjectId
		}
		oids = append(oids, oid)
	}

	var userInfos []*UserInfo
	err := m.conn.Find(ctx, &userInfos, bson.M{"_id": bson.M{"$in": oids}})
	switch err {
	case nil:
		return userInfos, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultUserInfoModel) UpdateByEmail(ctx context.Context, data *UserInfo) (*mongo.UpdateResult, error) {
	data.UpdateAt = time.Now()

	res, err := m.conn.UpdateOne(ctx, bson.M{"email": data.Email}, bson.M{"$set": data})
	return res, err
}

func (m *defaultUserInfoModel) FindByUserType(ctx context.Context, userType string) ([]*UserInfo, error) {
	var userInfos []*UserInfo

	err := m.conn.Find(ctx, &userInfos, bson.M{"user_type": userType}, options.Find().SetSort(bson.D{{"nickname", 1}}))
	switch err {
	case nil:
		return userInfos, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
