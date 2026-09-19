package consumer

import (
	"encoding/json"
	"time"

	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/common/email"
	"MuXiFresh-Be-2.0/app/userauth/cmd/api/internal/config"

	"github.com/zeromicro/go-queue/kq"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// 一次初始投递，失败后再最多重试 maxSendRetries 次。
	maxSendRetries  = 3
	maxSendAttempts = maxSendRetries + 1
	sendRetryDelay  = 500 * time.Millisecond
)

type message struct {
	Email    string `json:"email"`
	Type     string `json:"type"`
	RandCode string `json:"rand_code"`
}

// sendEmail is overridable so the retry logic stays testable without a live
// SMTP server.
var sendEmail = email.Send

var sleep = time.Sleep

func Consume(c config.Config) func() {
	return func() {
		q := kq.MustNewQueue(c.KqConsumerConf, kq.WithHandle(func(k, v string) error {
			handle(v)
			return nil
		}))
		defer q.Stop()
		q.Start()
	}
}

// handle delivers one message. Delivery is best effort: the code is already
// persisted by the API, so every outcome is logged and committed. Retrying a
// bad payload would only stall the offset, and retrying forever after SMTP
// outages would redeliver the whole backlog.
func handle(v string) {
	var msg message
	if err := json.Unmarshal([]byte(v), &msg); err != nil {
		logx.Errorf("drop malformed email message: %v", err)
		return
	}

	var lastErr error
	for attempt := 0; attempt < maxSendAttempts; attempt++ {
		if attempt > 0 {
			sleep(time.Duration(attempt) * sendRetryDelay)
		}

		lastErr = sendEmail(msg.Email, msg.Type, msg.RandCode)
		if lastErr == nil {
			return
		}
		if email.IsPermanent(lastErr) {
			logx.Errorf("drop undeliverable email to %s: %v", msg.Email, lastErr)
			return
		}
	}

	logx.Errorf("send email to %s failed after %d attempts: %v", msg.Email, maxSendAttempts, lastErr)
}
