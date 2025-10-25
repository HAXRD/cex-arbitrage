package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// CORSConfig CORS配置
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig 默认CORS配置
func DefaultCORSConfig() *CORSConfig {
	return &CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders: []string{
			"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With",
			"X-CSRF-Token", "Cache-Control", "Pragma", "Date", "If-Modified-Since",
		},
		ExposeHeaders:    []string{"Content-Length", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           86400, // 24小时
	}
}

// CORS 跨域资源共享中间件
func CORS() gin.HandlerFunc {
	return CORSWithConfig(DefaultCORSConfig())
}

// CORSWithConfig 使用自定义配置的CORS中间件
func CORSWithConfig(config *CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		
		// 检查是否允许该源
		if isOriginAllowed(origin, config.AllowOrigins) {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
		} else if len(config.AllowOrigins) == 1 && config.AllowOrigins[0] == "*" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		}

		// 设置其他CORS头
		if config.AllowCredentials {
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		if len(config.AllowMethods) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
		}

		if len(config.AllowHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
		}

		if len(config.ExposeHeaders) > 0 {
			c.Writer.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
		}

		if config.MaxAge > 0 {
			c.Writer.Header().Set("Access-Control-Max-Age", string(rune(config.MaxAge)))
		}

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

// isOriginAllowed 检查源是否被允许
func isOriginAllowed(origin string, allowedOrigins []string) bool {
	if origin == "" {
		return false
	}

	for _, allowedOrigin := range allowedOrigins {
		if allowedOrigin == "*" {
			return true
		}
		if allowedOrigin == origin {
			return true
		}
		// 支持通配符匹配
		if strings.Contains(allowedOrigin, "*") {
			pattern := strings.Replace(allowedOrigin, "*", ".*", -1)
			if matchPattern(origin, pattern) {
				return true
			}
		}
	}

	return false
}

// matchPattern 简单的模式匹配
func matchPattern(str, pattern string) bool {
	// 这里可以实现更复杂的模式匹配逻辑
	// 为了简化，只支持基本的通配符匹配
	return strings.Contains(str, strings.Replace(pattern, ".*", "", -1))
}

