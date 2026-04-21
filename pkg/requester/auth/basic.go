package auth

import (
	"encoding/base64"
	"net/http"
)

// BasicAuth HTTP基本认证
type BasicAuth struct {
	Username string
	Password string
}

// NewBasicAuth 创建基本认证
func NewBasicAuth(username, password string) *BasicAuth {
	return &BasicAuth{
		Username: username,
		Password: password,
	}
}

// Apply 应用认证
func (a *BasicAuth) Apply(req *http.Request) error {
	req.SetBasicAuth(a.Username, a.Password)
	return nil
}

// BearerAuth Bearer令牌认证
type BearerAuth struct {
	Token string
}

// NewBearerAuth 创建Bearer认证
func NewBearerAuth(token string) *BearerAuth {
	return &BearerAuth{
		Token: token,
	}
}

// Apply 应用认证
func (a *BearerAuth) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return nil
}

// CustomHeaderAuth 自定义头认证
type CustomHeaderAuth struct {
	Header string
	Value  string
}

// NewCustomHeaderAuth 创建自定义头认证
func NewCustomHeaderAuth(header, value string) *CustomHeaderAuth {
	return &CustomHeaderAuth{
		Header: header,
		Value:  value,
	}
}

// Apply 应用认证
func (a *CustomHeaderAuth) Apply(req *http.Request) error {
	req.Header.Set(a.Header, a.Value)
	return nil
}

// ProxyAuth 代理认证
type ProxyAuth struct {
	Username string
	Password string
}

// NewProxyAuth 创建代理认证
func NewProxyAuth(username, password string) *ProxyAuth {
	return &ProxyAuth{
		Username: username,
		Password: password,
	}
}

// GetProxyAuthorization 获取代理授权头
func (a *ProxyAuth) GetProxyAuthorization() string {
	auth := a.Username + ":" + a.Password
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// ParseCredential 解析认证凭据
func ParseCredential(credential string) (username, password string) {
	if idx := findIndex(credential, ':'); idx != -1 {
		username = credential[:idx]
		password = credential[idx+1:]
	} else {
		username = credential
	}
	return
}

// findIndex 查找字符位置
func findIndex(s string, ch rune) int {
	for i, c := range s {
		if c == ch {
			return i
		}
	}
	return -1
}
