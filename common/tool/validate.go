package tool

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidAvatarURL = errors.New("非法的头像地址")

// badURLSegments 是前端模板串拼接失败时可能落入 URL 的坏字面量
// （例如 `https://host/${response.key}` 在 key 缺失时得到 `https://host/undefined`）。
var badURLSegments = map[string]bool{
	"undefined": true,
	"null":      true,
	"nan":       true,
}

// ValidateAvatarURL 校验并规范化头像地址：返回去空白后的值（纯空白一律归一为空串），
// 非空时要求 http/https 且带主机名、不含拼接坏字面量。
// 调用方应持久化返回值，而非原始入参。
// URL 由客户端提交、可被直连伪造，前端守卫不可信，故在服务端校验。
func ValidateAvatarURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", ErrInvalidAvatarURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", ErrInvalidAvatarURL
	}
	if u.Hostname() == "" {
		return "", ErrInvalidAvatarURL
	}

	lower := strings.ToLower(s)
	if strings.Contains(lower, "[object") {
		return "", ErrInvalidAvatarURL
	}
	// 按 URL 分隔符切分后逐段精确匹配，覆盖 path、query、fragment
	for _, seg := range strings.FieldsFunc(lower, func(r rune) bool {
		switch r {
		case '/', '?', '&', '=', '#', ';':
			return true
		}
		return false
	}) {
		if badURLSegments[seg] {
			return "", ErrInvalidAvatarURL
		}
	}
	return s, nil
}
