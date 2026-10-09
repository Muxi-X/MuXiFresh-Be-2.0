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
// 空、无 @、含显示名（如 `Name <a@b.com>`）、域名非主机名形式均判为非法。
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
	if at < 0 || !validDomainLabels(s[at+1:]) {
		return "", ErrInvalidEmail
	}
	return s, nil
}

// validDomainLabels 校验域名为主机名形式：至少一个点，每个 label 非空、长度
// 不超过 63、仅含字母/数字/连字符且不以连字符开头或结尾。
// mail.ParseAddress 会放过 `a@bad!.com` 这类 SMTP 不可用的域名，故单独校验。
func validDomainLabels(domain string) bool {
	if !strings.Contains(domain, ".") {
		return false
	}
	for _, label := range strings.Split(domain, ".") {
		if label == "" || len(label) > 63 {
			return false
		}
		if label[0] == '-' || label[len(label)-1] == '-' {
			return false
		}
		for i := 0; i < len(label); i++ {
			c := label[i]
			isAlnum := (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9')
			if !isAlnum && c != '-' {
				return false
			}
		}
	}
	return true
}
