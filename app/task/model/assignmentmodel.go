package model

import (
	"context"
	"github.com/zeromicro/go-zero/core/stores/mon"
	"go.mongodb.org/mongo-driver/bson"
)

var _ AssignmentModel = (*customAssignmentModel)(nil)

type (
	// AssignmentModel is an interface to be customized, add more methods here,
	// and implement the added methods in customAssignmentModel.
	AssignmentModel interface {
		assignmentModel
		FindByGroup(ctx context.Context, group string) ([]*Assignment, error)
		FindByGroupYearandSemester(ctx context.Context, group string, year int64, semester string) ([]*Assignment, error)
	}

	customAssignmentModel struct {
		*defaultAssignmentModel
	}
)

// NewAssignmentModel returns a model for the mongo.
func NewAssignmentModel(url, db, collection string) AssignmentModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customAssignmentModel{
		defaultAssignmentModel: newDefaultAssignmentModel(conn),
	}
}

func (m *customAssignmentModel) FindByGroup(ctx context.Context, group string) ([]*Assignment, error) {
	var assignment []*Assignment
	err := m.conn.Find(ctx, &assignment, bson.M{"group": group})
	switch err {
	case nil:
		return assignment, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
func (m *customAssignmentModel) FindByGroupYearandSemester(ctx context.Context, group string, year int64, semester string) ([]*Assignment, error) {
	var assignment []*Assignment
	err := m.conn.Find(ctx, &assignment, bson.M{"group": group, "year": year, "semester": semester})
	switch err {
	case nil:
		return assignment, nil
	case mon.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}
