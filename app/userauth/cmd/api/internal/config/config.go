package config

import (
	"github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/rest"
	"github.com/zeromicro/go-zero/zrpc"
)

type CaptchaConf struct {
	Height          int
	Width           int
	Length          int
	Maxskew         float64
	Dotcount        int
	ExpireTime      int
	DebugExpireTime int
	TestingKey      string
}

type Config struct {
	rest.RestConf
	AccountCenterConf zrpc.RpcClientConf
	JwtAuth           struct {
		AccessSecret string
		AccessExpire int64
	}
	EmailConf *struct {
		Host     string
		Port     string
		UserName string
		Password string
	}
	Oss *struct {
		AccessKey  string
		SecretKey  string
		BucketName string
		DomainName string
	}
	CaptchaConf      *CaptchaConf
	EmailCodeExpired int
	RedisConf        redis.RedisConf
	JwtAuthChPass    struct {
		AccessSecret string
		AccessExpire int64
	}
}
