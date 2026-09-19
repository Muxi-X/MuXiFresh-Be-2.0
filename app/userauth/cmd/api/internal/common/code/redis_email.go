package code

const (
	// 值相等才删除并返回 1，用于「校验成功即消费」和「回滚仅删自己的值」两种场景。
	// 原子执行，避免校验/删除之间的并发窗口。
	deleteEmailCodeIfMatchScript = `
if redis.call('get', KEYS[1]) == ARGV[1] then
	redis.call('del', KEYS[1])
	return 1
end
return 0`
)

func SetEmailCode(prefix string, key string, value string) error {
	return redisClient.Setex(prefix+key, value, EmailCodeExpired*60)
}

// DelEmailCodeIfMatch removes the code only when it still equals value, so a
// rollback cannot delete a newer code written by a concurrent resend.
func DelEmailCodeIfMatch(prefix string, key string, value string) error {
	_, err := redisClient.Eval(deleteEmailCodeIfMatchScript, []string{prefix + key}, value)
	return err
}

func VerifyEmailCode(prefix string, key string, value string) bool {
	if value == "" {
		return false
	}
	matched, err := redisClient.Eval(deleteEmailCodeIfMatchScript, []string{prefix + key}, value)
	if err != nil {
		return false
	}
	result, ok := matched.(int64)
	return ok && result == 1
}
