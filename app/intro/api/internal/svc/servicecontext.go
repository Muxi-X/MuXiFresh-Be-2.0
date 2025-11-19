package svc

import (
	"MuXiFresh-Be-2.0/app/intro/api/internal/config"
	"MuXiFresh-Be-2.0/app/intro/rpc/introclient"
	"github.com/zeromicro/go-zero/zrpc"
)

type ServiceContext struct {
	Config      config.Config
	IntroClient introclient.IntroClient
}

func NewServiceContext(c config.Config) *ServiceContext {
	return &ServiceContext{
		Config:      c,
		IntroClient: introclient.NewIntroClient(zrpc.MustNewClient(c.IntroConf)),
	}
}
