package auth

import (
	"github.com/Azure/go-ntlmssp"
	"net/http"
)

// NTLMAuth NTLM认证
type NTLMAuth struct {
	Username string
	Password string
}

// NewNTLMAuth 创建NTLM认证
func NewNTLMAuth(username, password string) *NTLMAuth {
	return &NTLMAuth{
		Username: username,
		Password: password,
	}
}

// Apply 应用认证
func (a *NTLMAuth) Apply(req *http.Request) error {
	// NTLM认证需要通过Transport处理
	// 这里只是标记，实际认证在Transport中
	return nil
}

// GetTransport 获取NTLM Transport
func (a *NTLMAuth) GetTransport() http.RoundTripper {
	return &ntlmssp.Negotiator{
		RoundTripper: &http.Transport{},
	}
}

// Client 返回使用NTLM认证的HTTP客户端
func (a *NTLMAuth) Client() *http.Client {
	return &http.Client{
		Transport: a.GetTransport(),
	}
}
