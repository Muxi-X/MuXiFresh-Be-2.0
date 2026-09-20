package code

const (
	// 写入新码时同时清除旧的「已消费标记」：标记存在即代表这一代的码被用过，
	// 新一代码写入后旧标记必须失效。
	setEmailCodeScript = `
redis.call('setex', KEYS[1], ARGV[2], ARGV[1])
redis.call('del', KEYS[2])
return 1`

	// 校验通过即消费：值相等才删除，并留下「已消费标记」记录该码的剩余有效期
	// （PTTL），供业务失败时按相同的剩余 TTL 恢复。标记的 TTL 即原始截止时间。
	consumeEmailCodeScript = `
if redis.call('get', KEYS[1]) == ARGV[1] then
	local ttl = redis.call('pttl', KEYS[1])
	redis.call('del', KEYS[1])
	if ttl > 0 then
		redis.call('psetex', KEYS[2], ttl, ARGV[1])
	else
		redis.call('del', KEYS[2])
	end
	return 1
end
return 0`

	// 恢复本次消费掉的码：仅当「已消费标记」仍是本码、且当前没有更新的码时才写回，
	// 并把标记剩余的 TTL 原样应用给码，使其不晚于原始截止时间；剩余 TTL 已耗尽则
	// 视为失败，不复活过期码。
	restoreEmailCodeScript = `
if redis.call('get', KEYS[2]) == ARGV[1] then
	local ttl = redis.call('pttl', KEYS[2])
	if ttl <= 0 then
		redis.call('del', KEYS[2])
		return 0
	end
	if redis.call('exists', KEYS[1]) == 0 then
		redis.call('psetex', KEYS[1], ttl, ARGV[1])
		redis.call('del', KEYS[2])
		return 1
	end
end
return 0`

	// 发送失败回滚：仅当值仍为本次写入的码时才删除，避免误删重发覆盖的新码。
	deleteEmailCodeIfMatchScript = `
if redis.call('get', KEYS[1]) == ARGV[1] then
	redis.call('del', KEYS[1])
	return 1
end
return 0`

	// 已消费标记的 key 前缀，放在码 key 命名空间之外，避免邮箱中含后缀时撞 key。
	usedCodeKeyPrefix = "used:"
)

func emailCodeTTL() int {
	return EmailCodeExpired * 60
}

func usedCodeKey(prefix string, key string) string {
	return usedCodeKeyPrefix + prefix + key
}

func SetEmailCode(prefix string, key string, value string) error {
	_, err := redisClient.Eval(setEmailCodeScript,
		[]string{prefix + key, usedCodeKey(prefix, key)}, value, emailCodeTTL())
	return err
}

// DelEmailCodeIfMatch removes the code only when it still equals value, so a
// rollback cannot delete a newer code written by a concurrent resend.
func DelEmailCodeIfMatch(prefix string, key string, value string) error {
	_, err := redisClient.Eval(deleteEmailCodeIfMatchScript, []string{prefix + key}, value)
	return err
}

// RestoreEmailCode re-inserts a consumed code after the follow-up business step
// failed, so the user can retry without requesting a new one. It writes only
// when this exact code was the most recently consumed one, no newer code exists,
// and the original deadline has not passed, so a superseded or expired code is
// never revived. The restored code keeps the original remaining lifetime.
func RestoreEmailCode(prefix string, key string, value string) error {
	_, err := redisClient.Eval(restoreEmailCodeScript,
		[]string{prefix + key, usedCodeKey(prefix, key)}, value)
	return err
}

func VerifyEmailCode(prefix string, key string, value string) bool {
	if value == "" {
		return false
	}
	matched, err := redisClient.Eval(consumeEmailCodeScript,
		[]string{prefix + key, usedCodeKey(prefix, key)}, value)
	if err != nil {
		return false
	}
	result, ok := matched.(int64)
	return ok && result == 1
}
