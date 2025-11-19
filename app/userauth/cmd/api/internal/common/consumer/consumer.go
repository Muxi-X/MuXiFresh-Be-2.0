package consumer

import (
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/email"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"
	"encoding/json"
	"github.com/zeromicro/go-queue/kq"
)

func Consume(c config.Config) func() {
	return func() {
		q := kq.MustNewQueue(c.KqConsumerConf, kq.WithHandle(func(k, v string) error {
			//处理消息
			var tmp map[string]string
			if err := json.Unmarshal([]byte(v), &tmp); err != nil {
				return err
			}
			if err := email.Send(tmp["email"], tmp["type"]); err != nil {
				return err
			}
			return nil
		}))
		defer q.Stop()
		q.Start()
	}
}
