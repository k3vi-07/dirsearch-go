package requester

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/youruser/dirsearch-go/internal/dns"
	"github.com/youruser/dirsearch-go/internal/logger"
	"golang.org/x/time/rate"
)

// SyncRequester 同步HTTP请求器
type SyncRequester struct {
	client      *http.Client
	baseURL     string
	headers     sync.Map
	auth        AuthProvider
	rateLimiter *rate.Limiter
	config      *Config
	mu          sync.RWMutex
}

// NewSyncRequester 创建同步请求器
func NewSyncRequester(cfg *Config) (*SyncRequester, error) {
	if cfg == nil {
		cfg = &Config{
			ThreadCount:   30,
			Timeout:       10 * time.Second,
			MaxRetries:    1,
			VerifySSL:     false,
			MaxRedirects:  3,
			FollowRedirects: false,
		}
	}

	// 创建自定义Transport
	transport := &http.Transport{
		MaxIdleConns:        cfg.ThreadCount,
		MaxIdleConnsPerHost: cfg.ThreadCount,
		IdleConnTimeout:     90 * time.Second,
		// 自定义DNS解析器
		DialContext: (&dns.Resolver{
			Cache: dns.NewCache(),
		}).DialContext,
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !cfg.VerifySSL,
		},
		// 禁用HTTP/2（某些服务器不支持）
		ForceAttemptHTTP2: false,
	}

	// 设置网络接口
	if cfg.NetworkInterface != "" {
		transport.DialContext = (&dns.Resolver{
			Cache:        dns.NewCache(),
			Interface:    cfg.NetworkInterface,
		}).DialContext
	}

	// 创建HTTP客户端
	client := &http.Client{
		Transport: transport,
		Timeout:   cfg.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if !cfg.FollowRedirects {
				return http.ErrUseLastResponse
			}
			if len(via) >= cfg.MaxRedirects {
				return fmt.Errorf("stopped after %d redirects", cfg.MaxRedirects)
			}
			return nil
		},
	}

	// 创建速率限制器
	var rateLimiter *rate.Limiter
	if cfg.MaxRate > 0 {
		rateLimiter = rate.NewLimiter(rate.Limit(cfg.MaxRate), cfg.MaxRate)
	} else {
		rateLimiter = rate.NewLimiter(rate.Inf, 0)
	}

	// 设置代理
	if cfg.Proxy != "" {
		proxyURL, err := url.Parse(cfg.Proxy)
		if err != nil {
			return nil, fmt.Errorf("无效的代理URL: %w", err)
		}
		transport.Proxy = http.ProxyURL(proxyURL)
		logger.Debugf("使用代理: %s", cfg.Proxy)
	}

	r := &SyncRequester{
		client:      client,
		headers:     sync.Map{},
		rateLimiter: rateLimiter,
		config:      cfg,
	}

	// 设置默认User-Agent
	if cfg.UserAgent != "" {
		r.SetHeader("User-Agent", cfg.UserAgent)
	} else {
		r.SetHeader("User-Agent", "dirsearch-go")
	}

	return r, nil
}

// Request 发送HTTP请求
func (r *SyncRequester) Request(ctx context.Context, path string) (*Response, error) {
	// 等待速率限制
	if err := r.rateLimiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("速率限制等待失败: %w", err)
	}

	// 构建完整URL
	fullURL := r.baseURL + path

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, r.config.Headers["Method"], fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("创建请求失败: %w", err)
	}

	// 设置请求头
	r.headers.Range(func(key, value interface{}) bool {
		req.Header.Set(key.(string), value.(string))
		return true
	})

	// 应用认证
	if r.auth != nil {
		if err := r.auth.Apply(req); err != nil {
			return nil, fmt.Errorf("应用认证失败: %w", err)
		}
	}

	// 执行请求（带重试）
	var resp *http.Response
	var requestErr error

	for retry := 0; retry <= r.config.MaxRetries; retry++ {
		start := time.Now()
		resp, requestErr = r.client.Do(req)
		duration := time.Since(start)

		if requestErr != nil {
			if retry < r.config.MaxRetries {
				logger.Debugf("请求失败，重试 %d/%d: %v", retry+1, r.config.MaxRetries, requestErr)
				time.Sleep(time.Duration(retry+1) * time.Second)
				continue
			}
			return nil, fmt.Errorf("请求失败: %w", requestErr)
		}

		// 读取响应体
		body, err := r.readResponseBody(resp)
		if err != nil {
			return nil, fmt.Errorf("读取响应体失败: %w", err)
		}

		// 构建响应
		response := &Response{
			URL:         resp.Request.URL.String(),
			Status:      resp.StatusCode,
			Length:      int64(len(body)),
			ContentType: resp.Header.Get("Content-Type"),
			Body:        body,
			Redirect:    resp.Header.Get("Location"),
			Headers:     resp.Header.Clone(),
			Duration:    duration,
		}

		// 记录重定向历史
		// TODO: 实现重定向历史记录
		// 注意：标准库不直接提供ResponseHistory，需要手动跟踪
		response.History = []string{}

		// 解析路径
		if u, err := url.Parse(response.URL); err == nil {
			response.Path = u.Path
			if u.RawQuery != "" {
				response.Path += "?" + u.RawQuery
			}
		}

		// 尝试解码文本内容
		response.Content = decodeContent(body, resp.Header.Get("Content-Type"))

		return response, nil
	}

	return nil, requestErr
}

// readResponseBody 读取响应体（流式）
func (r *SyncRequester) readResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()

	// 检查Content-Length
	contentLength := resp.ContentLength
	if contentLength > MAX_RESPONSE_SIZE {
		logger.Debugf("响应过大 (%d bytes)，限制为 %d bytes", contentLength, MAX_RESPONSE_SIZE)
	}

	var body []byte
	var totalRead int64

	// 创建读取器
	reader := resp.Body
	buf := make([]byte, ITER_CHUNK_SIZE)

	for {
		// 检查大小限制
		if totalRead > MAX_RESPONSE_SIZE {
			logger.Debug("达到最大响应大小，停止读取")
			break
		}

		n, err := reader.Read(buf)
		if n > 0 {
			body = append(body, buf[:n]...)
			totalRead += int64(n)
		}

		if err != nil {
			if err == io.EOF {
				break
			}
			// 检查是否是二进制内容
			if isBinary(body) {
				logger.Debug("检测到二进制内容，停止读取")
				break
			}
			return nil, err
		}

		// 如果有Content-Length且已读取完，退出
		if contentLength > 0 && totalRead >= contentLength {
			break
		}
	}

	return body, nil
}

// SetURL 设置基础URL
func (r *SyncRequester) SetURL(url string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 确保URL以斜杠结尾
	if !strings.HasSuffix(url, "/") {
		url = url + "/"
	}

	r.baseURL = url
}

// SetHeader 设置请求头
func (r *SyncRequester) SetHeader(key, value string) {
	r.headers.Store(key, value)
}

// SetHeaders 批量设置请求头
func (r *SyncRequester) SetHeaders(headers map[string]string) {
	for k, v := range headers {
		r.SetHeader(k, v)
	}
}

// SetAuth 设置认证
func (r *SyncRequester) SetAuth(authProvider AuthProvider) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.auth = authProvider
}

// SetProxy 设置代理
func (r *SyncRequester) SetProxy(proxyURL string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	parsed, err := url.Parse(proxyURL)
	if err != nil {
		return err
	}

	if transport, ok := r.client.Transport.(*http.Transport); ok {
		transport.Proxy = http.ProxyURL(parsed)
		logger.Debugf("设置代理: %s", proxyURL)
		return nil
	}

	return fmt.Errorf("无法设置代理：无效的transport")
}

// Close 关闭请求器
func (r *SyncRequester) Close() error {
	r.client.CloseIdleConnections()
	return nil
}

// Rate 获取当前速率
func (r *SyncRequester) Rate() int {
	return r.config.MaxRate
}

// decodeContent 解码响应内容
func decodeContent(body []byte, contentType string) string {
	// 简单UTF-8解码
	// TODO: 添加更复杂的字符集检测和转换
	for i, b := range body {
		if b == 0 {
			// 发现空字节，可能是二进制
			return ""
		}
		if b > 127 {
			// 非ASCII字符，尝试UTF-8
			// 这里简化处理，实际应该用charset检测库
		}
		if i > 1024*100 { // 只检查前100KB
			break
		}
	}
	return string(body)
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
