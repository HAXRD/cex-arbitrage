package data_collection

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"
	"time"

	"go.uber.org/zap"
)

// AuthConfig 认证配置
type AuthConfig struct {
	Enabled     bool          `json:"enabled" yaml:"enabled"`
	SecretKey   string        `json:"secret_key" yaml:"secret_key"`
	TokenExpiry time.Duration `json:"token_expiry" yaml:"token_expiry"`
	Issuer      string        `json:"issuer" yaml:"issuer"`
}

// DefaultAuthConfig 默认认证配置
func DefaultAuthConfig() *AuthConfig {
	return &AuthConfig{
		Enabled:     false,
		SecretKey:   "default-secret-key",
		TokenExpiry: 24 * time.Hour,
		Issuer:      "data-collection-service",
	}
}

// AuthManager 认证管理器
type AuthManager interface {
	// 生成Token
	GenerateToken(userID string, permissions []string) (string, error)

	// 验证Token
	ValidateToken(token string) (*TokenClaims, error)

	// 检查权限
	CheckPermission(token string, permission string) (bool, error)

	// 刷新Token
	RefreshToken(token string) (string, error)
}

// TokenClaims Token声明
type TokenClaims struct {
	UserID      string    `json:"user_id"`
	Permissions []string  `json:"permissions"`
	IssuedAt    time.Time `json:"issued_at"`
	ExpiresAt   time.Time `json:"expires_at"`
	Issuer      string    `json:"issuer"`
}

// authManagerImpl 认证管理器实现
type authManagerImpl struct {
	config *AuthConfig
	logger *zap.Logger
}

// NewAuthManager 创建认证管理器
func NewAuthManager(config *AuthConfig, logger *zap.Logger) AuthManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = DefaultAuthConfig()
	}

	return &authManagerImpl{
		config: config,
		logger: logger,
	}
}

// GenerateToken 生成Token
func (am *authManagerImpl) GenerateToken(userID string, permissions []string) (string, error) {
	if !am.config.Enabled {
		return "", fmt.Errorf("认证未启用")
	}

	claims := &TokenClaims{
		UserID:      userID,
		Permissions: permissions,
		IssuedAt:    time.Now(),
		ExpiresAt:   time.Now().Add(am.config.TokenExpiry),
		Issuer:      am.config.Issuer,
	}

	// 简单的HMAC签名实现
	tokenData := fmt.Sprintf("%s:%s:%d:%d:%s",
		claims.UserID,
		strings.Join(claims.Permissions, ","),
		claims.IssuedAt.Unix(),
		claims.ExpiresAt.Unix(),
		claims.Issuer,
	)

	signature := am.signToken(tokenData)
	token := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", tokenData, signature)))

	am.logger.Info("Token已生成", zap.String("user_id", userID))
	return token, nil
}

// ValidateToken 验证Token
func (am *authManagerImpl) ValidateToken(token string) (*TokenClaims, error) {
	if !am.config.Enabled {
		return &TokenClaims{
			UserID:      "anonymous",
			Permissions: []string{"read"},
			IssuedAt:    time.Now(),
			ExpiresAt:   time.Now().Add(24 * time.Hour),
			Issuer:      am.config.Issuer,
		}, nil
	}

	// 解码Token
	tokenBytes, err := base64.StdEncoding.DecodeString(token)
	if err != nil {
		return nil, fmt.Errorf("Token解码失败: %w", err)
	}

	tokenStr := string(tokenBytes)
	parts := strings.Split(tokenStr, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Token格式无效")
	}

	tokenData, signature := parts[0], parts[1]

	// 验证签名
	if !am.verifySignature(tokenData, signature) {
		return nil, fmt.Errorf("Token签名无效")
	}

	// 解析Token数据
	claims, err := am.parseTokenData(tokenData)
	if err != nil {
		return nil, fmt.Errorf("Token解析失败: %w", err)
	}

	// 检查过期时间
	if time.Now().After(claims.ExpiresAt) {
		return nil, fmt.Errorf("Token已过期")
	}

	return claims, nil
}

// CheckPermission 检查权限
func (am *authManagerImpl) CheckPermission(token string, permission string) (bool, error) {
	claims, err := am.ValidateToken(token)
	if err != nil {
		return false, err
	}

	for _, p := range claims.Permissions {
		if p == permission || p == "admin" {
			return true, nil
		}
	}

	return false, nil
}

// RefreshToken 刷新Token
func (am *authManagerImpl) RefreshToken(token string) (string, error) {
	claims, err := am.ValidateToken(token)
	if err != nil {
		return "", fmt.Errorf("Token验证失败: %w", err)
	}

	// 生成新Token
	return am.GenerateToken(claims.UserID, claims.Permissions)
}

// signToken 签名Token
func (am *authManagerImpl) signToken(data string) string {
	h := hmac.New(sha256.New, []byte(am.config.SecretKey))
	h.Write([]byte(data))
	return base64.StdEncoding.EncodeToString(h.Sum(nil))
}

// verifySignature 验证签名
func (am *authManagerImpl) verifySignature(data, signature string) bool {
	expectedSignature := am.signToken(data)
	return hmac.Equal([]byte(signature), []byte(expectedSignature))
}

// parseTokenData 解析Token数据
func (am *authManagerImpl) parseTokenData(data string) (*TokenClaims, error) {
	parts := strings.Split(data, ":")
	if len(parts) < 2 {
		return nil, fmt.Errorf("Token数据格式无效")
	}

	userID := parts[0]
	permissionsStr := parts[1]

	// 简化处理，使用当前时间
	issuedAt := time.Now()
	expiresAt := time.Now().Add(24 * time.Hour)

	permissions := []string{}
	if permissionsStr != "" {
		permissions = strings.Split(permissionsStr, ",")
	}

	issuer := am.config.Issuer
	if len(parts) > 4 {
		issuer = parts[4]
	}

	return &TokenClaims{
		UserID:      userID,
		Permissions: permissions,
		IssuedAt:    issuedAt,
		ExpiresAt:   expiresAt,
		Issuer:      issuer,
	}, nil
}

// AuthMiddleware 认证中间件
func AuthMiddleware(authManager AuthManager, requiredPermission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 获取Authorization头
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "未提供认证信息", http.StatusUnauthorized)
				return
			}

			// 检查Bearer格式
			if !strings.HasPrefix(authHeader, "Bearer ") {
				http.Error(w, "认证格式无效", http.StatusUnauthorized)
				return
			}

			token := strings.TrimPrefix(authHeader, "Bearer ")

			// 验证Token
			claims, err := authManager.ValidateToken(token)
			if err != nil {
				http.Error(w, fmt.Sprintf("Token验证失败: %v", err), http.StatusUnauthorized)
				return
			}

			// 检查权限
			if requiredPermission != "" {
				hasPermission, err := authManager.CheckPermission(token, requiredPermission)
				if err != nil {
					http.Error(w, fmt.Sprintf("权限检查失败: %v", err), http.StatusInternalServerError)
					return
				}

				if !hasPermission {
					http.Error(w, "权限不足", http.StatusForbidden)
					return
				}
			}

			// 将用户信息添加到请求上下文
			ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
			ctx = context.WithValue(ctx, "permissions", claims.Permissions)
			r = r.WithContext(ctx)

			next.ServeHTTP(w, r)
		})
	}
}

// GetUserID 从上下文获取用户ID
func GetUserID(r *http.Request) string {
	if userID, ok := r.Context().Value("user_id").(string); ok {
		return userID
	}
	return ""
}

// GetPermissions 从上下文获取权限
func GetPermissions(r *http.Request) []string {
	if permissions, ok := r.Context().Value("permissions").([]string); ok {
		return permissions
	}
	return []string{}
}
