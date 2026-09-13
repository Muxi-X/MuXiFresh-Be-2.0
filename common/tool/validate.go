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

// ValidateAvatarURL 允许空串（未上传头像），否则要求 http/https 且不含拼接坏字面量。
// URL 由客户端提交、可被直连伪造，前端守卫不可信，故在服务端校验。
func ValidateAvatarURL(raw string) error {
	s := strings.TrimSpace(raw)
	if s == "" {
		return nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return ErrInvalidAvatarURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return ErrInvalidAvatarURL
	}
	if u.Host == "" {
		return ErrInvalidAvatarURL
	}

	if strings.Contains(strings.ToLower(s), "[object") {
		return ErrInvalidAvatarURL
	}
	for _, seg := range strings.Split(u.Path, "/") {
		if badURLSegments[strings.ToLower(seg)] {
			return ErrInvalidAvatarURL
		}
	}
	return nil
}
