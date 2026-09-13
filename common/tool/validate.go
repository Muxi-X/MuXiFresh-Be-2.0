package tool

import (
	"errors"
	"net/url"
	"strings"
)

// ErrInvalidAvatarURL 头像地址非法。
var ErrInvalidAvatarURL = errors.New("非法的头像地址")

// badURLSegments 是前端模板串拼接失败时可能落入 URL 的坏字面量
// （例如 `https://host/${response.key}` 在 key 缺失时得到 `https://host/undefined`）。
var badURLSegments = map[string]bool{
	"undefined": true,
	"null":      true,
	"nan":       true,
}

// ValidateAvatarURL 校验头像地址：
//   - 允许空串（表示未上传头像）
//   - 否则必须是 http/https 且带主机名
//   - 各路径段与整体不得包含前端拼接失败产生的坏字面量（undefined/null/NaN/[object Object]）
//
// 该 URL 由客户端提交、可被直连伪造，前端守卫不可信，故在服务端校验。
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

	// 坏字面量整体兜底（如 "[object Object]" 被编码或原样拼入）
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
