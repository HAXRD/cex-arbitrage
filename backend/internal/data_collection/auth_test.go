package data_collection

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAuthManager_BasicOperations(t *testing.T) {
	t.Run("创建认证管理器", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true

		authManager := NewAuthManager(config, zap.NewNop())
		require.NotNil(t, authManager)
	})

	t.Run("生成和验证Token", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 生成Token
		permissions := []string{"read", "write"}
		token, err := authManager.GenerateToken("user123", permissions)
		require.NoError(t, err)
		assert.NotEmpty(t, token)

		// 验证Token
		claims, err := authManager.ValidateToken(token)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, permissions, claims.Permissions)
		assert.Equal(t, config.Issuer, claims.Issuer)
	})

	t.Run("权限检查", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 生成Token
		permissions := []string{"read", "write"}
		token, err := authManager.GenerateToken("user123", permissions)
		require.NoError(t, err)

		// 检查有效权限
		hasPermission, err := authManager.CheckPermission(token, "read")
		require.NoError(t, err)
		assert.True(t, hasPermission)

		hasPermission, err = authManager.CheckPermission(token, "write")
		require.NoError(t, err)
		assert.True(t, hasPermission)

		// 检查无效权限
		hasPermission, err = authManager.CheckPermission(token, "admin")
		require.NoError(t, err)
		assert.False(t, hasPermission)
	})

	t.Run("管理员权限", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 生成管理员Token
		permissions := []string{"admin"}
		token, err := authManager.GenerateToken("admin", permissions)
		require.NoError(t, err)

		// 管理员应该有所有权限
		hasPermission, err := authManager.CheckPermission(token, "read")
		require.NoError(t, err)
		assert.True(t, hasPermission)

		hasPermission, err = authManager.CheckPermission(token, "write")
		require.NoError(t, err)
		assert.True(t, hasPermission)

		hasPermission, err = authManager.CheckPermission(token, "admin")
		require.NoError(t, err)
		assert.True(t, hasPermission)
	})

	t.Run("Token刷新", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 生成Token
		permissions := []string{"read", "write"}
		originalToken, err := authManager.GenerateToken("user123", permissions)
		require.NoError(t, err)

		// 刷新Token
		newToken, err := authManager.RefreshToken(originalToken)
		require.NoError(t, err)
		assert.NotEqual(t, originalToken, newToken)

		// 验证新Token
		claims, err := authManager.ValidateToken(newToken)
		require.NoError(t, err)
		assert.Equal(t, "user123", claims.UserID)
		assert.Equal(t, permissions, claims.Permissions)
	})
}

func TestAuthManager_ErrorCases(t *testing.T) {
	t.Run("认证未启用", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = false

		authManager := NewAuthManager(config, zap.NewNop())

		// 生成Token应该失败
		_, err := authManager.GenerateToken("user123", []string{"read"})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "认证未启用")
	})

	t.Run("无效Token", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 验证无效Token
		_, err := authManager.ValidateToken("invalid-token")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "Token解码失败")
	})

	t.Run("权限不足", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 生成只有读权限的Token
		token, err := authManager.GenerateToken("user123", []string{"read"})
		require.NoError(t, err)

		// 检查写权限应该失败
		hasPermission, err := authManager.CheckPermission(token, "write")
		require.NoError(t, err)
		assert.False(t, hasPermission)
	})
}

func TestAuthMiddleware(t *testing.T) {
	t.Run("有效Token", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 生成Token
		token, err := authManager.GenerateToken("user123", []string{"read"})
		require.NoError(t, err)

		// 创建测试处理器
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)
			permissions := GetPermissions(r)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("user:%s,permissions:%v", userID, permissions)))
		})

		// 应用认证中间件
		middleware := AuthMiddleware(authManager, "read")
		wrappedHandler := middleware(handler)

		// 创建请求
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// 执行请求
		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user:user123")
		assert.Contains(t, w.Body.String(), "permissions:[read]")
	})

	t.Run("无效Token", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 创建测试处理器
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// 应用认证中间件
		middleware := AuthMiddleware(authManager, "")
		wrappedHandler := middleware(handler)

		// 创建请求（无效Token）
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		w := httptest.NewRecorder()

		// 执行请求
		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("缺少认证头", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 创建测试处理器
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// 应用认证中间件
		middleware := AuthMiddleware(authManager, "")
		wrappedHandler := middleware(handler)

		// 创建请求（无认证头）
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// 执行请求
		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("权限不足", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = true
		config.SecretKey = "test-secret-key"

		authManager := NewAuthManager(config, zap.NewNop())

		// 生成只有读权限的Token
		token, err := authManager.GenerateToken("user123", []string{"read"})
		require.NoError(t, err)

		// 创建测试处理器
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
		})

		// 应用认证中间件（需要写权限）
		middleware := AuthMiddleware(authManager, "write")
		wrappedHandler := middleware(handler)

		// 创建请求
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		w := httptest.NewRecorder()

		// 执行请求
		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})

	t.Run("认证未启用", func(t *testing.T) {
		config := DefaultAuthConfig()
		config.Enabled = false

		authManager := NewAuthManager(config, zap.NewNop())

		// 创建测试处理器
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userID := GetUserID(r)
			permissions := GetPermissions(r)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(fmt.Sprintf("user:%s,permissions:%v", userID, permissions)))
		})

		// 应用认证中间件
		middleware := AuthMiddleware(authManager, "")
		wrappedHandler := middleware(handler)

		// 创建请求（无认证头）
		req := httptest.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()

		// 执行请求
		wrappedHandler.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "user:anonymous")
		assert.Contains(t, w.Body.String(), "permissions:[read]")
	})
}
