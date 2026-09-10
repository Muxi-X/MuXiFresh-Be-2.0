package svc

import (
	"MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/test/api/internal/config"
	"MuXiFresh-Be-2.0/app/test/rpc/testclient"
	"MuXiFresh-Be-2.0/app/user/cmd/rpc/user/userclient"
	userauthModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config         config.Config
	TestClient     testclient.TestClient
	UserClient     userclient.UserClient
	FormClient     model.EntryFormModel
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
		TestClient:     testclient.NewTestClient(zrpc.MustNewClient(c.TestConf, rpcOpts...)),
		UserClient:     userclient.NewUserClient(zrpc.MustNewClient(c.UserConf, rpcOpts...)),
		FormClient:     model.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
		UserInfoClient: userauthModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
	}
}
