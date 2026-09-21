package svc

import (
	formmodel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/config"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config          config.Config
	ScheduleClient  model.ScheduleModel
	EntryFormClient formmodel.EntryFormModel
	UserInfoClient  userauthModel.UserInfoModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// schedule 的索引迁移由 schedule 服务负责（数据归属方），失败即启动失败
	if err := EnsureScheduleIndexes(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB); err != nil {
		logx.Errorf("EnsureScheduleIndexes failed: %v", err)
		panic(err)
	}

	return &ServiceContext{
		Config:          c,
		ScheduleClient:  model.NewScheduleModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "schedule"),
		EntryFormClient: formmodel.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
		UserInfoClient:  userauthModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
	}
}
