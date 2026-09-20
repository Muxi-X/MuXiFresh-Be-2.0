package code

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/svc"
	"github.com/jinzhu/copier"
)

// EmailCodeExpired 邮箱验证码过期时间
var EmailCodeExpired int

// Conf captcha 配置
type Conf struct {
	Height          int
	Width           int
	Length          int
	Maxskew         float64
	Dotcount        int
	ExpireTime      int
	DebugExpireTime int
	TestingKey      string
}

var conf Conf

func Load(c config.Config, ctx *svc.ServiceContext) {
	copier.Copy(&conf, c.CaptchaConf)
	redisClient = ctx.RedisClient
	EmailCodeExpired = c.EmailCodeExpired
	if EmailCodeExpired <= 0 {
		// 缺失时验证码会永不过期（或不合法），必须 fail fast。
		panic("EmailCodeExpired must be positive")
	}
}
