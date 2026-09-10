package svc

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/producer"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/rpc/accountCenter/accountcenterclient"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config              config.Config
	KqPusher            *producer.Pusher
	RedisClient         *redis.Redis
	AccountCenterClient accountcenterclient.AccountCenterClient
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
		Config:              c,
		KqPusher:            producer.NewPusher(c.KqConf),
		RedisClient:         redis.MustNewRedis(c.Infra.Redis),
		AccountCenterClient: accountcenterclient.NewAccountCenterClient(zrpc.MustNewClient(c.AccountCenterConf, rpcOpts...)),
	}
}
