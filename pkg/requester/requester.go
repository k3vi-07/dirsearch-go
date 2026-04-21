package requester

import (
	"context"
	"io"
	"net/http"
	"time"
)

// Requester HTTP请求器接口
type Requester interface {
	// Request 发送HTTP请求
	Request(ctx context.Context, path string) (*Response, error)

	// SetURL 设置基础URL
	SetURL(url string)

	// SetHeader 设置请求头
	SetHeader(key, value string)

	// SetHeaders 批量设置请求头
	SetHeaders(headers map[string]string)

	// SetAuth 设置认证
	SetAuth(authProvider AuthProvider)

	// SetProxy 设置代理
	SetProxy(proxyURL string) error

	// Close 关闭请求器
	Close() error

	// Rate 获取当前速率
	Rate() int
}

// AuthProvider 认证提供者接口
type AuthProvider interface {
	// Apply 应用认证到请求
	Apply(req *http.Request) error
}

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

	// Content 文本内容（UTF-8解码）
	Content string

	// Body 原始响应体
	Body []byte

	// Redirect 重定向URL
	Redirect string

	// Headers 响应头
	Headers http.Header

	// Duration 请求耗时
	Duration time.Duration

	// History 重定向历史
	History []string
}

// Config 请求器配置
type Config struct {
	// ThreadCount 线程数
	ThreadCount int

	// Timeout 超时时间
	Timeout time.Duration

	// MaxRate 最大请求速率
	MaxRate int

	// UserAgent User-Agent
	UserAgent string

	// Headers 自定义请求头
	Headers map[string]string

	// Proxy 代理URL
	Proxy string

	// ProxyAuth 代理认证
	ProxyAuth string

	// VerifySSL 验证SSL证书
	VerifySSL bool

	// FollowRedirects 跟随重定向
	FollowRedirects bool

	// MaxRedirects 最大重定向次数
	MaxRedirects int

	// MaxRetries 最大重试次数
	MaxRetries int

	// NetworkInterface 网络接口
	NetworkInterface string
}

// StreamReader 流式读取器
type StreamReader interface {
	Read(p []byte) (n int, err error)
	Close() error
}

// bodyReader 响应体读取器
type bodyReader struct {
	reader io.ReadCloser
	chunk  []byte
	pos    int
}

// newBodyReader 创建响应体读取器
func newBodyReader(reader io.ReadCloser, chunkSize int) *bodyReader {
	return &bodyReader{
		reader: reader,
		chunk:  make([]byte, 0, chunkSize),
	}
}

// Read 读取数据
func (r *bodyReader) Read(p []byte) (n int, err error) {
	if r.pos < len(r.chunk) {
		n = copy(p, r.chunk[r.pos:])
		r.pos += n
		return n, nil
	}

	// 读取下一块
	r.chunk = r.chunk[:cap(r.chunk)]
	n, err = io.ReadFull(r.reader, r.chunk)
	if err != nil {
		return n, err
	}
	r.chunk = r.chunk[:n]
	r.pos = 0

	n = copy(p, r.chunk)
	r.pos += n
	return n, nil
}

// Close 关闭读取器
func (r *bodyReader) Close() error {
	return r.reader.Close()
}

const (
	// ITER_CHUNK_SIZE 迭代块大小
	ITER_CHUNK_SIZE = 512

	// MAX_RESPONSE_SIZE 最大响应大小
	MAX_RESPONSE_SIZE = 10 * 1024 * 1024 // 10MB
)
