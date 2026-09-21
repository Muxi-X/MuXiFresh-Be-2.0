package svc

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/internal/config"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config        config.Config
	FormClient    model.EntryFormModel
	UserInfoModel userauthModel.UserInfoModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// entry_form 的索引迁移由 form 服务负责（数据归属方），失败即启动失败：
	// 静默失败会让重新报名继续撞旧索引，且无任何外部迹象。
	if err := EnsureEntryFormIndexes(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB); err != nil {
		logx.Errorf("EnsureEntryFormIndexes failed: %v", err)
		panic(err)
	}

	return &ServiceContext{
		Config:        c,
		FormClient:    model.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
		UserInfoModel: userauthModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
	}
}
