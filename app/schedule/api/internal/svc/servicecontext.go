package svc

import (
	"MuXiFresh-Be-2.0/app/schedule/api/internal/config"
	"MuXiFresh-Be-2.0/app/schedule/rpc/scheduleclient"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config         config.Config
	ScheduleClient scheduleclient.ScheduleClient
	UserInfoClient userauthModel.UserInfoModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	clientInterceptor, err := rpcauth.UnaryClientInterceptor(c.Infra.RpcAuth.Token)
	if err != nil {
		panic(err)
	}
	rpcOpts := []zrpc.ClientOption{
		zrpc.WithDialOption(grpc.WithChainUnaryInterceptor(clientInterceptor)),
	}

	return &ServiceContext{
		Config:         c,
		ScheduleClient: scheduleclient.NewScheduleClient(zrpc.MustNewClient(c.ScheduleConf, rpcOpts...)),
		UserInfoClient: userauthModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
	}
}
