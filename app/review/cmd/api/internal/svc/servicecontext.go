package svc

import (
	externalModel2 "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/review/cmd/api/internal/config"
	externalModel3 "MuXiFresh-Be-2.0/app/schedule/model"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	externalModel1 "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config         config.Config
	EntryFormModel externalModel2.EntryFormModel
	UserClient     userclient.UserClient
	ScheduleClient externalModel3.ScheduleModel
	UserInfoModel  externalModel1.UserInfoModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	clientInterceptor, err := rpcauth.UnaryClientInterceptor(c.Infra.RpcAuth.Token)
	if err != nil {
		panic(err)
	}
	rpcOpts := []zrpc.ClientOption{
		zrpc.WithDialOption(grpc.WithChainUnaryInterceptor(clientInterceptor)),
	}

	if err := EnsureReviewIndexes(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB); err != nil {
		logx.Errorf("EnsureReviewIndexes failed: %v", err)
	}

	return &ServiceContext{
		Config:         c,
		EntryFormModel: externalModel2.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
		UserClient:     userclient.NewUserClient(zrpc.MustNewClient(c.UserConf, rpcOpts...)),
		ScheduleClient: externalModel3.NewScheduleModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "schedule"),
		UserInfoModel:  externalModel1.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
	}
}
