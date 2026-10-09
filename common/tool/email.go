package tool

import "strings"

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
