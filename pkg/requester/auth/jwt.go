package auth

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt"
)

// JWTAuth JWT认证
type JWTAuth struct {
	Token string
}

// NewJWTAuth 创建JWT认证
func NewJWTAuth(token string) *JWTAuth {
	return &JWTAuth{
		Token: token,
	}
}

// Apply 应用认证
func (a *JWTAuth) Apply(req *http.Request) error {
	req.Header.Set("Authorization", "Bearer "+a.Token)
	return nil
}

// ValidateToken 验证JWT令牌（可选）
func (a *JWTAuth) ValidateToken(secret string) (*jwt.Token, error) {
	token, err := jwt.Parse(a.Token, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("意外的签名方法: %v", token.Header["alg"])
		}
		return []byte(secret), nil
	})

	if err != nil {
		return nil, err
	}

	return token, nil
}

// ParseCredentials 解析认证凭据并创建认证提供者
func ParseCredentials(authType, credential string) (interface{}, error) {
	authType = strings.ToLower(authType)

	switch authType {
	case "basic":
		username, password := ParseCredential(credential)
		return NewBasicAuth(username, password), nil

	case "digest":
		username, password := ParseCredential(credential)
		return NewDigestAuth(username, password), nil

	case "ntlm":
		username, password := ParseCredential(credential)
		return NewNTLMAuth(username, password), nil

	case "bearer":
		return NewBearerAuth(credential), nil

	case "jwt":
		return NewJWTAuth(credential), nil

	default:
		return nil, fmt.Errorf("不支持的认证类型: %s", authType)
	}
}
