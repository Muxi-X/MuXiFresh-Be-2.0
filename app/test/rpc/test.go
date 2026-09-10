package main

import (
	"MuXiFresh-Be-2.0/common/nacos"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"flag"
	"fmt"

	"MuXiFresh-Be-2.0/app/test/rpc/internal/config"
	"MuXiFresh-Be-2.0/app/test/rpc/internal/server"
	"MuXiFresh-Be-2.0/app/test/rpc/internal/svc"
	"MuXiFresh-Be-2.0/app/test/rpc/pb"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/test.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoadService("test-rpc", &c, &c.Infra)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterTestClientServer(grpcServer, server.NewTestClientServer(ctx))

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
