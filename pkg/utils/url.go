package utils

import (
	"net/url"
	"regexp"
	"strings"
)

// CleanPath 清理路径
func CleanPath(path string, keepQuery, keepFragment bool) string {
	if !keepFragment {
		if idx := strings.Index(path, "#"); idx != -1 {
			path = path[:idx]
		}
	}

	if !keepQuery {
		if idx := strings.Index(path, "?"); idx != -1 {
			path = path[:idx]
		}
	}

	return path
}

// ParsePath 解析URL路径
func ParsePath(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return u.Path
}

// ReplacePath 替换路径
func ReplacePath(content, path, replaceWith string) string {
	// URL编码
	encodedPath := url.QueryEscape(path)
	doubleEncodedPath := url.QueryEscape(encodedPath)

	decodedPath, _ := url.QueryUnescape(path)
	doubleDecodedPath, _ := url.QueryUnescape(decodedPath)

	// 尝试多种替换
	replacements := []struct {
		from string
		to   string
	}{
		{encodedPath, replaceWith},
		{doubleEncodedPath, replaceWith},
		{decodedPath, replaceWith},
		{doubleDecodedPath, replaceWith},
		{path, replaceWith},
	}

	result := content
	for _, repl := range replacements {
		// 使用正则替换（支持部分匹配）
		re := regexp.MustCompile(regexp.QuoteMeta(repl.from) + "(?=[^\\w]|$)")
		result = re.ReplaceAllString(result, repl.to)
	}

	return result
}

// GetReadableSize 获取人类可读的大小
func GetReadableSize(size int64) string {
	const unit = 1024
	if size < unit {
		return "0B"
	}

	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return string(rune('K'+exp)) + "B"
}
