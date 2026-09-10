package main

import (
	"MuXiFresh-Be-2.0/common/nacos"
	"MuXiFresh-Be-2.0/common/rpcauth"
	"flag"
	"fmt"

	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/internal/config"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/internal/server"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/internal/svc"
	"MuXiFresh-Be-2.0/app/task/cmd/rpc/submission/pb"

	"github.com/zeromicro/go-zero/core/service"
	"github.com/zeromicro/go-zero/zrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var configFile = flag.String("f", "etc/submission.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoadService("submission", &c, &c.Infra)
	ctx := svc.NewServiceContext(c)

	s := zrpc.MustNewServer(c.RpcServerConf, func(grpcServer *grpc.Server) {
		pb.RegisterSubmissionClientServer(grpcServer, server.NewSubmissionClientServer(ctx))

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
