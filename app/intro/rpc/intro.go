package main

import (
	"flag"
	"fmt"

	"MuXiFresh-Be-2.0/app/intro/rpc/internal/config"
	"MuXiFresh-Be-2.0/app/intro/rpc/internal/server"
	"MuXiFresh-Be-2.0/app/intro/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/intro/rpc/pb"
	"MuXiFresh-Be-2.0/common/nacos"
	"MuXiFresh-Be-2.0/common/rpcauth"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/intro.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoadService("intro-rpc", &c, &c.Infra)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterIntroClientServer(grpcServer, server.NewIntroClientServer(ctx))

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
