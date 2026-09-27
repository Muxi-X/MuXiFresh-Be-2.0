// Package config holds the gateway configuration, loaded from the Nacos data
// ID "gateway" plus the shared "infra" data ID.
package config

import (
	"MuXiFresh-Be-2.0/common/infra"

	"github.com/zeromicro/go-zero/rest"
)

// Upstream is a static backend the gateway can proxy to. Names are referenced
// by the route table.
type Upstream struct {
	Name string
	Host string
	Port int
}

// Config is the gateway configuration.
type Config struct {
	rest.RestConf
	Infra     infra.Config
	Upstreams []Upstream
}

// ApplyInfra inherits the shared middleware switches from infra and then pins
// the two that a transparent reverse proxy must keep disabled.
//
// The pin is necessary because infra.ApplyMiddlewares only ever enables
// switches; a false written in the gateway data ID would be flipped back to
// true whenever infra enables the same switch. These two are transparency
// invariants rather than deployment knobs:
//   - Timeout buffers the whole response body in memory, so a streaming export
//     such as review's Excel download would be held back and inflate gateway
//     memory.
//   - Gunzip replaces the request body with a decompressed reader while
//     leaving Content-Encoding in place, so an upstream receives a plain body
//     advertised as gzip.
//
// It also ties the entry size filter (Middlewares.MaxBytes) to MaxBytes being
// positive: M2 loses the json default, so an unset limit would be 0 and
// enabling the switch would reject every request that carries a body.
func (c *Config) ApplyInfra() {
	c.Infra.ApplyMiddlewares(&c.RestConf)

	c.RestConf.Middlewares.Timeout = false
	c.RestConf.Middlewares.Gunzip = false
	c.RestConf.Middlewares.MaxBytes = c.RestConf.MaxBytes > 0
}
