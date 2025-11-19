package model

import "github.com/zeromicro/go-zero/core/stores/mon"

var _ IntroModel = (*customIntroModel)(nil)

type (
	// IntroModel is an interface to be customized, add more methods here,
	// and implement the added methods in customIntroModel.
	IntroModel interface {
		introModel
	}

	customIntroModel struct {
		*defaultIntroModel
	}
)

// NewIntroModel returns a model for the mongo.
func NewIntroModel(url, db, collection string) IntroModel {
	conn := mon.MustNewModel(url, db, collection)
	return &customIntroModel{
		defaultIntroModel: newDefaultIntroModel(conn),
	}
}
