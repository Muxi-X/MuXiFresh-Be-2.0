package model

import (
	"context"

	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var _ SubmissionModel = (*customSubmissionModel)(nil)

type (
	// SubmissionModel is an interface to be customized, add more methods here,
	// and implement the added methods in customSubmissionModel.
	SubmissionModel interface {
		submissionModel
		FindByUserIdAndAssignmentID(ctx context.Context, userId, assignmentID string) ([]*Submission, error)
		FindByAssignmentID(ctx context.Context, assignmentID string) ([]*SubmissionStats, error)
		CountByUserAndAssignment(ctx context.Context, userId, assignmentId string) (int64, error)
	}

	customSubmissionModel struct {
		*defaultSubmissionModel
	}
)

// NewSubmissionModel returns a model for the mongo.
func NewSubmissionModel(url, db, collection string) SubmissionModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customSubmissionModel{
		defaultSubmissionModel: newDefaultSubmissionModel(conn),
	}
}

func (m *customSubmissionModel) FindByUserIdAndAssignmentID(ctx context.Context, userId, assignmentID string) ([]*Submission, error) {
	uid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return nil, err
	}
	aid, err := primitive.ObjectIDFromHex(assignmentID)
	if err != nil {
		return nil, err
	}
	var submissions []*Submission
	err = m.conn.Find(ctx, &submissions, bson.M{"user_id": uid, "assignment_id": aid}, options.Find().SetSort(bson.D{{"version", 1}}))
	switch err {
	case nil:
		return submissions, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *customSubmissionModel) FindByAssignmentID(ctx context.Context, assignmentID string) ([]*SubmissionStats, error) {
	var submissionStats []*SubmissionStats
	aid, err := primitive.ObjectIDFromHex(assignmentID)
	if err != nil {
		return nil, err
	}
	pipeline := mongo.Pipeline{
		{{"$match", bson.M{"assignment_id": aid}}},
		{{"$group", bson.M{
			"_id":         "$user_id",
			"status":      bson.M{"$last": "$status"},
			"version_num": bson.M{"$sum": 1},
		}}},
		{{"$sort", bson.D{{"_id", 1}}}},
	}
	err = m.conn.Aggregate(ctx, &submissionStats, pipeline)
	switch err {
	case nil:
		return submissionStats, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *customSubmissionModel) CountByUserAndAssignment(ctx context.Context, userId, assignmentId string) (int64, error) {
	userid, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return 0, ErrInvalidObjectId
	}
	assignmentID, err := primitive.ObjectIDFromHex(assignmentId)
	if err != nil {
		return 0, ErrInvalidObjectId
	}
	filter := bson.M{
		"user_id":       userid,
		"assignment_id": assignmentID,
	}
	return m.conn.CountDocuments(ctx, filter)
}
