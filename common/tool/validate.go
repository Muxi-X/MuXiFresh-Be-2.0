package tool

import (
	"errors"
	"net/url"
	"strings"
)

var ErrInvalidAvatarURL = errors.New("非法的头像地址")

var ErrInvalidResourceURL = errors.New("非法的资源地址")

// errInvalidURL 是内部校验失败的统一原因，由导出函数映射为各自语义的错误。
var errInvalidURL = errors.New("invalid url")

// badURLSegments 是前端模板串拼接失败时可能落入 URL 的坏字面量
// （例如 `https://host/${response.key}` 在 key 缺失时得到 `https://host/undefined`）。
var badURLSegments = map[string]bool{
	"undefined": true,
	"null":      true,
	"nan":       true,
}

// validateHTTPURL 校验并规范化单个 http/https 地址：返回去空白后的值
// （纯空白一律归一为空串），非空时要求 http/https 且带主机名、不含拼接坏字面量。
// URL 由客户端提交、可被直连伪造，前端守卫不可信，故在服务端校验。
func validateHTTPURL(raw string) (string, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return "", nil
	}

	u, err := url.Parse(s)
	if err != nil {
		return "", errInvalidURL
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return "", errInvalidURL
	}
	if u.Hostname() == "" {
		return "", errInvalidURL
	}

	lower := strings.ToLower(s)
	if strings.Contains(lower, "[object") {
		return "", errInvalidURL
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
			return "", errInvalidURL
		}
	}
	return s, nil
}

// ValidateAvatarURL 校验并规范化头像地址：返回去空白后的值（纯空白一律归一为空串），
// 非空时要求 http/https 且带主机名、不含拼接坏字面量。
// 调用方应持久化返回值，而非原始入参。
func ValidateAvatarURL(raw string) (string, error) {
	s, err := validateHTTPURL(raw)
	if err != nil {
		return "", ErrInvalidAvatarURL
	}
	return s, nil
}

// ValidateResourceURLs 校验并规范化资源地址列表：逐项去空白、丢弃空项，
// 任一项非法则整体报错（fail closed），返回清洗后的切片。调用方应持久化返回值。
func ValidateResourceURLs(raws []string) ([]string, error) {
	cleaned := make([]string, 0, len(raws))
	for _, raw := range raws {
		s, err := validateHTTPURL(raw)
		if err != nil {
			return nil, ErrInvalidResourceURL
		}
		if s != "" {
			cleaned = append(cleaned, s)
		}
	}
	if len(cleaned) == 0 {
		return nil, nil
	}
	return cleaned, nil
}
