package response

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/unicode"
	"golang.org/x/text/transform"
)

// Response HTTP响应
type Response struct {
	// URL 完整URL
	URL string

	// Path 路径部分
	Path string

	// Status HTTP状态码
	Status int

	// Length 响应体长度
	Length int64

	// ContentType 内容类型
	ContentType string

	// Content 文本内容
	Content string

	// Body 原始响应体
	Body []byte

	// Redirect 重定向URL
	Redirect string

	// Headers 响应头
	Headers http.Header

	// Cookies 响应cookies
	Cookies []*http.Cookie
}

// NewResponse 从HTTP响应创建Response
func NewResponse(resp *http.Response, body []byte) (*Response, error) {
	if resp == nil {
		return nil, fmt.Errorf("nil response")
	}

	r := &Response{
		URL:         resp.Request.URL.String(),
		Status:      resp.StatusCode,
		Length:      int64(len(body)),
		ContentType: resp.Header.Get("Content-Type"),
		Body:        body,
		Redirect:    resp.Header.Get("Location"),
		Headers:     resp.Header.Clone(),
		Cookies:     resp.Cookies(),
	}

	// 解析路径
	if u, err := url.Parse(r.URL); err == nil {
		r.Path = u.Path
		if u.RawQuery != "" {
			r.Path += "?" + u.RawQuery
		}
	}

	// 解码内容
	r.Content = decodeBody(body, resp.Header)

	return r, nil
}

// decodeBody 解码响应体
func decodeBody(body []byte, headers http.Header) string {
	// 检查是否是二进制
	if isBinary(body) {
		return ""
	}

	// 尝试UTF-8解码
	if utf8.Valid(body) {
		return string(body)
	}

	// 尝试其他编码
	contentType := headers.Get("Content-Type")
	if encoding := detectEncoding(contentType); encoding != nil {
		decoder := encoding.NewDecoder()
		transformed, _, err := transform.Bytes(decoder, body)
		if err == nil {
			return string(transformed)
		}
	}

	// 尝试UTF-16
	decoder := unicode.UTF16(unicode.LittleEndian, unicode.UseBOM).NewDecoder()
	transformed, _, err := transform.Bytes(decoder, body)
	if err == nil {
		return string(transformed)
	}

	// 回退到原始字符串（可能包含乱码）
	return string(body)
}

// detectEncoding 检测编码
func detectEncoding(contentType string) encoding.Encoding {
	// 从Content-Type提取charset
	if idx := strings.Index(contentType, "charset="); idx != -1 {
		charset := strings.ToLower(contentType[idx+8:])
		charset = strings.TrimSpace(charset)

		switch charset {
		case "utf-8":
			return nil
		case "iso-8859-1", "latin1":
			return encoding.Nop
		// 可以添加更多编码
		}
	}

	return nil
}

// isBinary 检查是否是二进制内容
func isBinary(data []byte) bool {
	if len(data) == 0 {
		return false
	}

	// 检查前512字节
	checkLen := 512
	if len(data) < checkLen {
		checkLen = len(data)
	}

	nullCount := 0
	for i := 0; i < checkLen; i++ {
		if data[i] == 0 {
			nullCount++
		}
	}

	// 如果空字节超过1%，认为是二进制
	return nullCount > checkLen/100
}

// GetSize 获取人类可读的大小
func GetSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

// CleanPath 清理路径
func CleanPath(path string, keepQuery bool, keepFragment bool) string {
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

// StreamResponse 流式读取响应
func StreamResponse(resp *http.Response, maxBytes int64) ([]byte, error) {
	defer resp.Body.Close()

	// 检查Content-Length
	contentLength := resp.ContentLength
	if contentLength > 0 && contentLength < maxBytes {
		maxBytes = contentLength
	}

	// 使用LimitReader限制读取大小
	limitedReader := io.LimitReader(resp.Body, maxBytes)

	// 读取响应
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(limitedReader); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// ReadLines 按行读取响应
func ReadLines(resp *http.Response) ([]string, error) {
	body, err := StreamResponse(resp, 10*1024*1024) // 10MB
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(bytes.NewReader(body))
	var lines []string

	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

// GetStatusCodeText 获取状态码文本
func GetStatusCodeText(status int) string {
	switch {
	case status >= 100 && status < 200:
		return "1xx Informational"
	case status >= 200 && status < 300:
		return "2xx Success"
	case status >= 300 && status < 400:
		return "3xx Redirection"
	case status >= 400 && status < 500:
		return "4xx Client Error"
	case status >= 500 && status < 600:
		return "5xx Server Error"
	default:
		return "Unknown"
	}
}

// IsSuccess 是否是成功状态码
func IsSuccess(status int) bool {
	return status >= 200 && status < 300
}

// IsRedirect 是否是重定向状态码
func IsRedirect(status int) bool {
	return status >= 300 && status < 400
}

// IsClientError 是否是客户端错误
func IsClientError(status int) bool {
	return status >= 400 && status < 500
}

// IsServerError 是否是服务器错误
func IsServerError(status int) bool {
	return status >= 500 && status < 600
}

// ExtractLinks 从HTML中提取链接
func ExtractLinks(content string) []string {
	// 简单的链接提取正则
	linkRegex := regexp.MustCompile(`<a\s+(?:[^>]*?\s+)?href="([^"]*)"`)
	matches := linkRegex.FindAllStringSubmatch(content, -1)

	links := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			links = append(links, match[1])
		}
	}

	return links
}

// ExtractForms 从HTML中提取表单
func ExtractForms(content string) []string {
	// 简单的表单提取正则
	formRegex := regexp.MustCompile(`<form\s+(?:[^>]*?\s+)?action="([^"]*)"`)
	matches := formRegex.FindAllStringSubmatch(content, -1)

	forms := make([]string, 0, len(matches))
	for _, match := range matches {
		if len(match) > 1 {
			forms = append(forms, match[1])
		}
	}

	return forms
}

// ComparePaths 比较两个路径是否相同
func ComparePaths(path1, path2 string) bool {
	// 标准化路径
	path1 = CleanPath(path1, false, false)
	path2 = CleanPath(path2, false, false)

	return path1 == path2
}

// Hash 响应哈希（用于过滤重复响应）
func (r *Response) Hash() string {
	// 简单哈希：基于状态码和响应体大小
	return fmt.Sprintf("%d:%d", r.Status, r.Length)
}

// String 字符串表示
func (r *Response) String() string {
	return fmt.Sprintf("[%d] %s (%s)", r.Status, r.Path, GetSize(r.Length))
}
