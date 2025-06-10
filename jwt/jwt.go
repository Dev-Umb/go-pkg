/**
* @Author: chenhao29
* @Date: 2024/6/10
* @QQ: 1149558764
* @Email: i@umb.ink
 */
package jwt

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/Dev-Umb/go-pkg/errno"
	"github.com/Dev-Umb/go-pkg/logger"
	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt/v4"
)

type UserInfo struct {
	UnionId   string `json:"union_id"`   // 用户在通用账户平台的唯一身份标识
	OpenId    string `json:"open_id"`    // 用户在特定业务平台的身份标识
	UserId    string `json:"user_id"`    // 用户ID
	UserName  string `json:"user_name"`  // 用户名
	AvatarURL string `json:"avatar_url"` //
}

type CustomClaims struct {
	UserInfo
	jwt.RegisteredClaims
}

var jwtSecret = "JWT_SECRET"

func InitJwtSecret(targetJwtSecret string) {
	jwtSecret = targetJwtSecret
}

// GenerateToken 生成token
func GenerateToken(user UserInfo) (string, error) {
	// 创建声明
	claims := CustomClaims{
		UserInfo: user,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 30)), // 24小时过期
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
		},
	}

	// 创建token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 获取密钥
	if jwtSecret == "" {
		logger.Errorf(context.Background(), "JWT密钥未配置")
		return "", errors.New("JWT密钥未配置")
	}

	// 签名token
	return token.SignedString([]byte(jwtSecret))
}

// ParseToken 解析token
func ParseToken(tokenString string) (*CustomClaims, error) {
	// 获取密钥
	if jwtSecret == "" {
		logger.Errorf(context.Background(), "JWT密钥未配置")
		return nil, errors.New("JWT密钥未配置")
	}

	// 解析token
	token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证token
	if claims, ok := token.Claims.(*CustomClaims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("无效的token")
}

// RefreshToken 刷新token
func RefreshToken(tokenString string) (string, error) {
	// 解析token
	claims, err := ParseToken(tokenString)
	if err != nil {
		return "", err
	}

	// 更新过期时间
	claims.ExpiresAt = jwt.NewNumericDate(time.Now().Add(time.Hour * 24))

	// 创建新token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名token
	return token.SignedString([]byte(jwtSecret))
}

// IsJwtTokenValid 判断jwt token是否有效
func IsJwtTokenValid(tokenString string) (bool, error) {
	// Parse the token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(jwtSecret), nil
	})
	if err != nil {
		logger.Errorf(context.Background(), "parse jwt token error: %v", err)
		return false, err
	}

	// Check if the token is valid
	if _, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return true, nil
	} else {
		return false, errno.InvalidTokenError
	}
}

// ExtractBearerToken 从Authorization头中提取Bearer token
func ExtractBearerToken(authHeader string) string {
	if authHeader == "" {
		logger.Errorf(context.Background(), "Authorization header is empty")
		return ""
	}

	// 处理Bearer token
	token := authHeader
	if len(authHeader) > 7 && strings.ToUpper(authHeader[0:7]) == "BEARER " {
		token = authHeader[7:]
	} else if len(authHeader) > 6 && strings.ToUpper(authHeader[0:6]) == "BEARER" {
		token = authHeader[6:]
	}

	return strings.TrimSpace(token)
}

// GetJwtToken 从请求中获取JWT token
func GetJwtToken(c *gin.Context) string {
	// 首先尝试从上下文中获取已处理的token
	tokenInterface, exists := c.Get("token")
	if exists {
		return tokenInterface.(string)
	}

	// 否则，从头部获取并处理
	authHeader := c.GetHeader("Authorization")
	return ExtractBearerToken(authHeader)
}
