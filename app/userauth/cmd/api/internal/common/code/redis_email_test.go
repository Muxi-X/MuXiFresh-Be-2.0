package code

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/zeromicro/go-zero/core/stores/redis"
)

func withMiniredis(t *testing.T) *miniredis.Miniredis {
	t.Helper()

	server := miniredis.RunT(t)
	previous := redisClient
	redisClient = redis.MustNewRedis(redis.RedisConf{Host: server.Addr(), Type: "node"})
	t.Cleanup(func() { redisClient = previous })

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
