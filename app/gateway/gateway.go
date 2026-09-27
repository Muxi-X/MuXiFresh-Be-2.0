package main

import (
	"flag"
	"fmt"
	"net/http"

	"MuXiFresh-Be-2.0/app/gateway/internal/config"
	"MuXiFresh-Be-2.0/app/gateway/internal/handler"
	"MuXiFresh-Be-2.0/app/gateway/internal/proxy"
	"MuXiFresh-Be-2.0/app/gateway/internal/route"
	"MuXiFresh-Be-2.0/app/gateway/internal/router"
	"MuXiFresh-Be-2.0/common/nacos"

	"github.com/zeromicro/go-zero/rest"
)

var configFile = flag.String("f", "etc/gateway.yaml", "the config file")

// dispatchMethods are every HTTP method the gateway accepts and forwards.
var dispatchMethods = []string{
	http.MethodGet,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodHead,
	http.MethodOptions,
}

func main() {
	flag.Parse()

	var c config.Config
	nacos.MustLoadService("gateway", &c, &c.Infra)

	server, err := buildServer(c)
	if err != nil {
		panic(err)
	}
	defer server.Stop()

	fmt.Printf("Starting gateway at %s:%d...\n", c.Host, c.Port)
	server.Start()
}

// buildServer wires the route table, the static upstreams and the catch-all
// router into a rest server. It fails fast when the route table references an
// upstream that is not configured.
func buildServer(c config.Config) (*rest.Server, error) {
	rules := route.Default()

	names := make([]string, 0, len(c.Upstreams))
	for _, upstream := range c.Upstreams {
		names = append(names, upstream.Name)
	}
	if err := route.Validate(rules, names); err != nil {
		return nil, err
	}

	proxies, err := proxy.NewSet(c.Upstreams)
	if err != nil {
		return nil, err
	}

	dispatcher := handler.NewDispatcher(rules, proxies)
	server := rest.MustNewServer(c.RestConf, rest.WithRouter(router.New()))

	routes := make([]rest.Route, 0, len(dispatchMethods))
	for _, method := range dispatchMethods {
		routes = append(routes, rest.Route{
			Method:  method,
			Path:    "/*",
			Handler: dispatcher.ServeHTTP,
		})
	}
	server.AddRoutes(routes)

	return server, nil
}
