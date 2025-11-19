package svc

import (
	"MuXiFresh-Be-2.0/app/intro/model"
	"MuXiFresh-Be-2.0/app/intro/rpc/internal/config"
)

type ServiceContext struct {
	Config      config.Config
	IntroClient model.IntroModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		IntroClient: model.NewIntroModel(c.MongoDBConf.URL, c.MongoDBConf.DB, "intro"),
	}
}
