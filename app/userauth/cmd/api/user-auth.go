package main

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/code"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/consumer"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/email"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/tube"
	"flag"
	"fmt"
	"github.com/zeromicro/go-zero/core/threading"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/handler"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"

	"github.com/zeromicro/go-zero/core/conf"
	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/user-auth.yaml", "the config file")

func main() {
	flag.Parse()

	var c config.Config
	conf.MustLoad(*configFile, &c)

	server := rest.MustNewServer(c.RestConf)
	defer server.Stop()

	ctx := svc.NewServiceContext(c)
	handler.RegisterHandlers(server, ctx)
	
	//加载captcha redis 和 配置
	code.Load(c, ctx)
	//加载邮箱配置
	email.Load(c)
	//加载图床配置
	tube.Load(c)
	//启一个消费信息处理
	threading.GoSafe(consumer.Consume(c))

	fmt.Printf("Starting server at %s:%d...\n", c.Host, c.Port)
	server.Start()
}
