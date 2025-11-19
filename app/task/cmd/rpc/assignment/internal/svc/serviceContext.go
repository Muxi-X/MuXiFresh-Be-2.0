package svc

import (
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/assignment/internal/config"
	"MuXiFresh-Be-2.0/app/task/model"
)

type ServiceContext struct {
	Config                config.Config
	AssignmentModelClient model.AssignmentModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:                c,
		AssignmentModelClient: model.NewAssignmentModel(c.MongoConf.URL, c.MongoConf.DB, "assignment"),
	}
}
