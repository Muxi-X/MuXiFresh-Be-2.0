package svc

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/model"

	"github.com/zeromicro/go-zero/core/logx"
)

type ServiceContext struct {
	Config         config.Config
	UserInfoClient model.UserInfoModel
	UserAuthClient model.UserAuthModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	// 邮箱唯一约束由账户数据归属方 accountCenter 负责，创建失败即启动失败：
	// 静默失败会让同邮箱多账号继续产生，且无任何外部迹象。
	if err := EnsureAccountIndexes(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB); err != nil {
		logx.Errorf("EnsureAccountIndexes failed: %v", err)
		panic(err)
	}

	return &ServiceContext{
		Config:         c,
		UserInfoClient: model.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
		UserAuthClient: model.NewUserAuthModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userauth"),
	}
}
