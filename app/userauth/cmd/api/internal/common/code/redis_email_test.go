package code

import (
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func withMiniredis(t *testing.T) *miniredis.Miniredis {
	t.Helper()

	server := miniredis.RunT(t)
	previousClient := redisClient
	previousExpired := EmailCodeExpired
	redisClient = redis.MustNewRedis(redis.RedisConf{Host: server.Addr(), Type: "node"})
	// 生产由 code.Load 从配置注入，测试里显式设置以模拟真实 TTL。
	EmailCodeExpired = 10
	t.Cleanup(func() {
		redisClient = previousClient
		EmailCodeExpired = previousExpired
	})

	return server
}

func TestVerifyEmailCodeConsumesOnce(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "set_password"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("set code: %v", err)
	}

	if !VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("correct code must verify")
	}
	if VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("code must not be reusable after a successful verify")
	}
}

func TestVerifyEmailCodeConcurrentlyConsumedOnce(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "set_password"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("set code: %v", err)
	}

	const goroutines = 8
	start := make(chan struct{})
	results := make(chan bool, goroutines)
	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			results <- VerifyEmailCode(prefix, email, "ABCDEF")
		}()
	}

	close(start)
	wg.Wait()
	close(results)

	succeeded := 0
	for ok := range results {
		if ok {
			succeeded++
		}
	}
	if succeeded != 1 {
		t.Fatalf("exactly one concurrent verify may succeed, got %d", succeeded)
	}
}

func TestVerifyEmailCodeRejectsWrongValueWithoutConsuming(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "register"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("set code: %v", err)
	}

	if VerifyEmailCode(prefix, email, "WRONG1") {
		t.Fatal("wrong code must not verify")
	}
	if !VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("a wrong attempt must not consume the stored code")
	}
}

func TestVerifyEmailCodeRejectsEmptyValue(t *testing.T) {
	withMiniredis(t)

	if err := SetEmailCode("register", "user@example.com", "ABCDEF"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if VerifyEmailCode("register", "user@example.com", "") {
		t.Fatal("empty submitted value must never verify")
	}
}

func TestDelEmailCodeIfMatchKeepsNewerCode(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "set_password"
		email  = "user@example.com"
	)
	// 模拟并发：后一次请求已覆盖写入新码，前一次失败请求的回滚不得删除它。
	if err := SetEmailCode(prefix, email, "NEW123"); err != nil {
		t.Fatalf("set code: %v", err)
	}

	if err := DelEmailCodeIfMatch(prefix, email, "OLD123"); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "NEW123") {
		t.Fatal("rollback of a stale code must not delete the newer code")
	}
}

func TestDelEmailCodeIfMatchRemovesOwnCode(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "set_password"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "OWN123"); err != nil {
		t.Fatalf("set code: %v", err)
	}

	if err := DelEmailCodeIfMatch(prefix, email, "OWN123"); err != nil {
		t.Fatalf("rollback: %v", err)
	}
	if VerifyEmailCode(prefix, email, "OWN123") {
		t.Fatal("rollback must delete the code it wrote")
	}
}

func TestRestoreEmailCodeEnablesRetry(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "register"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	// 校验会消费掉验证码，模拟随后的业务步骤失败。
	if !VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("setup: code should verify once")
	}
	if VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("setup: code should already be consumed")
	}

	if err := RestoreEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("restored code must verify, otherwise retry is impossible")
	}
}

func TestRestoreEmailCodeKeepsNewerCode(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "register"
		email  = "user@example.com"
	)
	// 用户消费掉旧码后重新获取了新码，恢复旧码不得覆盖它。
	if err := SetEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("setup: old code should verify once")
	}
	if err := SetEmailCode(prefix, email, "NEW123"); err != nil {
		t.Fatalf("set code: %v", err)
	}

	if err := RestoreEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "NEW123") {
		t.Fatal("restore must not clobber a newer code")
	}
	if VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("restore must not revive a superseded code")
	}
}

func TestRestoreEmailCodeDoesNotReviveConsumedNewerCode(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "register"
		email  = "user@example.com"
	)
	// 关键交错：旧码 OLD 消费后业务卡住；用户重新取 NEW 并成功消费；
	// 此时 OLD 的恢复不得把它写回。
	if err := SetEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("setup: old code should verify once")
	}
	if err := SetEmailCode(prefix, email, "NEW123"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "NEW123") {
		t.Fatal("setup: new code should verify once")
	}

	if err := RestoreEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("a superseded code must never be revived")
	}
}

func TestRestoreEmailCodeDoesNotReviveAfterMarkerGone(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "register"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("setup: old code should verify once")
	}

	// 已消费标记过期/被清理后，旧码不再允许恢复。
	if _, err := redisClient.Del(usedCodeKey(prefix, email)); err != nil {
		t.Fatalf("clear used marker: %v", err)
	}

	if err := RestoreEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("restore must not revive when the used marker is gone")
	}
}

func TestRestoreEmailCodeRestoresTheNewestCode(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "register"
		email  = "user@example.com"
	)
	// OLD 消费后用户重取 NEW 并消费，此时 NEW 的业务步骤同样失败，应恢复 NEW。
	if err := SetEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("setup: old code should verify once")
	}
	if err := SetEmailCode(prefix, email, "NEW123"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "NEW123") {
		t.Fatal("setup: new code should verify once")
	}

	if err := RestoreEmailCode(prefix, email, "NEW123"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "NEW123") {
		t.Fatal("restoring the newest consumed code must enable retry")
	}
}

func TestSetEmailCodeClearsUsedMarker(t *testing.T) {
	withMiniredis(t)

	const (
		prefix = "register"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("setup: old code should verify once")
	}

	// 重新发码必须使上一代的已消费标记失效。
	if err := SetEmailCode(prefix, email, "NEW123"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if err := RestoreEmailCode(prefix, email, "OLD123"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if VerifyEmailCode(prefix, email, "OLD123") {
		t.Fatal("a new code must invalidate the previous used marker")
	}
}

func TestRestoreEmailCodeKeepsOriginalDeadline(t *testing.T) {
	server := withMiniredis(t)

	const (
		prefix = "set_password"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("set code: %v", err)
	}

	// 让码临近过期（总时长 10 分钟，推进 9 分钟，仅剩 1 分钟）。
	server.FastForward(9 * time.Minute)
	if !VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("setup: code should still verify before expiry")
	}
	if err := RestoreEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("restored code must verify once before the original deadline")
	}

	// 越过原始截止时间后，恢复出来的码必须已过期，不能续期存活。
	server.FastForward(2 * time.Minute)
	if err := RestoreEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("restored code must not outlive its original deadline")
	}
}

func TestRestoreEmailCodeFailsAfterOriginalDeadline(t *testing.T) {
	server := withMiniredis(t)

	const (
		prefix = "set_password"
		email  = "user@example.com"
	)
	if err := SetEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("set code: %v", err)
	}
	if !VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("setup: code should verify once")
	}

	// 消费后已过原始截止时间，恢复不得复活该码。
	server.FastForward(11 * time.Minute)
	if err := RestoreEmailCode(prefix, email, "ABCDEF"); err != nil {
		t.Fatalf("restore: %v", err)
	}
	if VerifyEmailCode(prefix, email, "ABCDEF") {
		t.Fatal("an expired code must never be revived")
	}
}
