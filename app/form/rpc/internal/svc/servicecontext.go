package svc

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/internal/config"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
)

type ServiceContext struct {
	Config        config.Config
	FormClient    model.EntryFormModel
	UserInfoModel userauthModel.UserInfoModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:        c,
		FormClient:    model.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
		UserInfoModel: userauthModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
	}
}
