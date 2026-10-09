package tool

import (
	"errors"
	"net/mail"
	"strings"
)

// ErrInvalidEmail 表示邮箱为空或格式非法。
var ErrInvalidEmail = errors.New("非法的邮箱地址")

// NormalizeEmail 规范化邮箱：去首尾空白，并把 @ 之后的域名部分转为小写。
//
// 域名大小写不敏感，故统一小写，避免同一邮箱因域名大小写不同被当作不同账号；
// 本地部分（@ 之前）大小写敏感，原样保留。无 @ 时仅去空白，非法值交由校验层处理。
func NormalizeEmail(raw string) string {
	s := strings.TrimSpace(raw)
	at := strings.LastIndex(s, "@")
	if at < 0 {
		return s
	}
	return s[:at] + "@" + strings.ToLower(s[at+1:])
}

// ValidateEmail 校验并规范化邮箱：返回 NormalizeEmail 后的值。
// 空、无 @、含显示名（如 `Name <a@b.com>`）、域名无点号均判为非法。
// 仅用于有副作用的入口（发验证码/注册/改绑）；查询类入口应只做规范化。
func ValidateEmail(raw string) (string, error) {
	s := NormalizeEmail(raw)
	if s == "" {
		return "", ErrInvalidEmail
	}
	addr, err := mail.ParseAddress(s)
	if err != nil {
		return "", ErrInvalidEmail
	}
	// 拒绝带显示名的形式：解析结果应与原串一致。
	if addr.Address != s {
		return "", ErrInvalidEmail
	}
	at := strings.LastIndex(s, "@")
	if at < 0 || !strings.Contains(s[at+1:], ".") {
		return "", ErrInvalidEmail
	}
	return s, nil
}
