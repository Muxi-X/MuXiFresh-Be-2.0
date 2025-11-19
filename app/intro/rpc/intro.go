package main

import (
	"flag"
	"fmt"

	"MuXiFresh-Be-2.0/app/intro/rpc/internal/config"
	"MuXiFresh-Be-2.0/app/intro/rpc/internal/server"
	"MuXiFresh-Be-2.0/app/intro/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/intro/rpc/pb"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/intro.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterIntroClientServer(grpcServer, server.NewIntroClientServer(ctx))

		if c.Mode == service.DevMode || c.Mode == service.TestMode {
			reflection.Register(grpcServer)
		}
	})
	defer s.Stop()

	fmt.Printf("Starting rpc server at %s...\n", c.ListenOn)
	s.Start()
}
