package svc

import (
	"MuXiFresh-Be-2.0/app/form/api/internal/config"
	formModel "MuXiFresh-Be-2.0/app/form/model"
	"MuXiFresh-Be-2.0/app/form/rpc/entryformclient"
	schedulemodel "MuXiFresh-Be-2.0/app/schedule/model"
	externalModel "MuXiFresh-Be-2.0/app/userauth/model"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
)

type ServiceContext struct {
	Config              config.Config
	FormClient          entryformclient.EntryFormClient
	EntryFormModel      formModel.EntryFormModel
	UserInfoModelClient externalModel.UserInfoModel
	ScheduleModel       schedulemodel.ScheduleModel
	RecruitSettingModel formModel.RecruitSettingModel
}

func NewServiceContext(c config.Config) *ServiceContext {
	clientInterceptor, err := rpcauth.UnaryClientInterceptor(c.Infra.RpcAuth.Token)
	if err != nil {
		panic(err)
	}
	rpcOpts := []zrpc.ClientOption{
		zrpc.WithDialOption(grpc.WithChainUnaryInterceptor(clientInterceptor)),
	}

	// recruit_setting 的 cycle 唯一约束是乐观锁首写的兜底，建不起来即启动失败，
	// 否则并发首写可能插入重复届次。
	if err := EnsureRecruitSettingIndexes(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB); err != nil {
		logx.Errorf("EnsureRecruitSettingIndexes failed: %v", err)
		panic(err)
	}

	return &ServiceContext{
		Config:              c,
		FormClient:          entryformclient.NewEntryFormClient(zrpc.MustNewClient(c.FormConf, rpcOpts...)),
		EntryFormModel:      formModel.NewEntryFormModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "entry_form"),
		UserInfoModelClient: externalModel.NewUserInfoModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "userinfo"),
		ScheduleModel:       schedulemodel.NewScheduleModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "schedule"),
		RecruitSettingModel: formModel.NewRecruitSettingModel(c.Infra.MongoDB.URL, c.Infra.MongoDB.DB, "recruit_setting"),
	}
}
