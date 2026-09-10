package svc

import (
	"MuXiFresh-Be-2.0/app/form/api/internal/config"
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config              config.Config
	FormClient          entryformclient.EntryFormClient
	UserInfoModelClient externalModel.UserInfoModel
	ScheduleModel       schedulemodel.ScheduleModel
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
		FormClient:          entryformclient.NewEntryFormClient(zrpc.MustNewClient(c.FormConf, rpcOpts...)),
		UserInfoModelClient: externalModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
		ScheduleModel:       schedulemodel.NewScheduleModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "schedule"),
	}
}
