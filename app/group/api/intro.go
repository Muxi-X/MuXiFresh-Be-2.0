package main

import (
	"MuxiFresh2.0/MuXiFresh-Be-2.0/app/group/api/internal/config"
	"MuxiFresh2.0/MuXiFresh-Be-2.0/app/group/api/internal/handler"
	"MuxiFresh2.0/MuXiFresh-Be-2.0/app/group/api/internal/svc"
	"flag"
	"fmt"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/intro-api.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}