package svc

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/test/rpc/internal/config"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
)

type ServiceContext struct {
	Config          config.Config
	UserInfoClient  userauthModel.UserInfoModel
	EntryFormClient model.EntryFormModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:          c,
		UserInfoClient:  userauthModel.NewUserInfoModel(c.MongoDBConf.URL, c.MongoDBConf.DB, "userinfo"),
		EntryFormClient: model.NewEntryFormModel(c.MongoDBConf.URL, c.MongoDBConf.DB, "entry_form"),
	}
}
