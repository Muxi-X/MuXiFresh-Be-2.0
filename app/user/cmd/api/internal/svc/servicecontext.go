package svc

import (
	"MuXiFresh-Be-2.0/app/user/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config     config.Config
	UserClient userclient.UserClient
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
		Config:     c,
		UserClient: userclient.NewUserClient(zrpc.MustNewClient(c.UserConf, rpcOpts...)),
	}
}
