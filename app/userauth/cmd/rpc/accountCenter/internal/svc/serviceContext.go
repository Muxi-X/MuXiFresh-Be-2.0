package svc

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/model"
	ccnusvc "MuXiFresh-Be-2.0/common/ccnu"
)

type ServiceContext struct {
	Config         config.Config
	UserInfoClient model.UserInfoModel
	UserAuthClient model.UserAuthModel
	CCNUSvc        ccnusvc.CCNUService
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		UserInfoClient: model.NewUserInfoModel(c.MongoConf.URL, c.MongoConf.DB, "userinfo"),
		UserAuthClient: model.NewUserAuthModel(c.MongoConf.URL, c.MongoConf.DB, "userauth"),
		CCNUSvc:        ccnusvc.NewCCNUService(),
	}
}
