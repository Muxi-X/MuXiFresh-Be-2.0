package main

import (
	"MuXiFresh-Be-2.0/common/nacos"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"flag"
	"fmt"

	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/config"
	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/server"
	"MuXiFresh-Be-2.0/app/schedule/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/schedule/rpc/pb"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/schedule.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoadService("schedule-rpc", &c, &c.Infra)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterScheduleClientServer(grpcServer, server.NewScheduleClientServer(ctx))
		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	if interceptor, err := rpcauth.UnaryServerInterceptor(c.Infra.RpcAuth.Token); err != nil {
		panic(err)
	} else {
		s.AddUnaryInterceptors(interceptor)
	}

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
