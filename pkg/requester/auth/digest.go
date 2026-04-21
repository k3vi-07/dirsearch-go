package auth

import (
	"github.com/icholy/digest"
	"net/http"
)

// DigestAuth Digest认证
type DigestAuth struct {
	Transport digest.Transport
}

// NewDigestAuth 创建Digest认证
func NewDigestAuth(username, password string) *DigestAuth {
	return &DigestAuth{
		Transport: digest.Transport{
			Username: username,
			Password: password,
		},
	}
}

// Apply 应用认证
func (a *DigestAuth) Apply(req *http.Request) error {
	// Digest认证需要通过Transport处理
	// 这里只是标记，实际认证在Transport中
	return nil
}

// Client 返回使用Digest认证的HTTP客户端
func (a *DigestAuth) Client() *http.Client {
	return &http.Client{
		Transport: &a.Transport,
	}
}
