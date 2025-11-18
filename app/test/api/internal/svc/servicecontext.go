package svc

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/test/api/internal/config"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config         config.Config
	UserClient     userclient.UserClient
	FormClient     model.EntryFormModel
	UserInfoClient userauthModel.UserInfoModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:         c,
		UserClient:     userclient.NewUserClient(zrpc.MustNewClient(c.UserConf)),
		FormClient:     model.NewEntryFormModel(c.MongoDBConf.URL, c.MongoDBConf.DB, "entry_form"),
		UserInfoClient: userauthModel.NewUserInfoModel(c.MongoDBConf.URL, c.MongoDBConf.DB, "userinfo"),
	}
}
