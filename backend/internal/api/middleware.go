package api

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ErrorHandler 错误处理中间件
func ErrorHandler(logger *zap.Logger) gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			logger.Error("Panic recovered",
				zap.String("error", err),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
			)
			InternalErrorResponse(c, "服务器内部错误", map[string]interface{}{
				"error": err,
			})
		} else {
			logger.Error("Panic recovered",
				zap.Any("error", recovered),
				zap.String("path", c.Request.URL.Path),
				zap.String("method", c.Request.Method),
				zap.String("ip", c.ClientIP()),
			)
			InternalErrorResponse(c, "服务器内部错误", nil)
		}
		c.Abort()
	})
}

// ErrorLoggerMiddleware 错误日志中间件
func ErrorLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 只记录错误状态码
		if param.StatusCode >= 400 {
			logger.Error("HTTP Request Error",
				zap.String("method", param.Method),
				zap.String("path", param.Path),
				zap.Int("status", param.StatusCode),
				zap.String("ip", param.ClientIP),
				zap.String("user_agent", param.Request.UserAgent()),
				zap.Duration("latency", param.Latency),
				zap.String("error", param.ErrorMessage),
			)
		}
		return ""
	})
}

// NotFoundHandler 404处理中间件
func NotFoundHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		NotFoundResponse(c, "请求的资源不存在", map[string]interface{}{
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		})
	}
}

// MethodNotAllowedHandler 405处理中间件
func MethodNotAllowedHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		MethodNotAllowedResponse(c, "请求方法不允许", map[string]interface{}{
			"path":   c.Request.URL.Path,
			"method": c.Request.Method,
		})
	}
}

// TimeoutHandler 超时处理中间件
func TimeoutHandler(timeout time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置超时上下文
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// 创建超时通道
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		select {
		case <-done:
			// 请求正常完成
		case <-ctx.Done():
			// 请求超时
			c.Abort()
			ErrorResponse(c, http.StatusRequestTimeout, "REQUEST_TIMEOUT", "请求超时", nil)
		}
	}
}

// RateLimitHandler 限流处理中间件
func RateLimitHandler(maxRequests int, window time.Duration) gin.HandlerFunc {
	// 简单的内存限流器（生产环境建议使用Redis）
	requests := make(map[string][]time.Time)

	return func(c *gin.Context) {
		clientIP := c.ClientIP()
		now := time.Now()

		// 清理过期记录
		if clientRequests, exists := requests[clientIP]; exists {
			var validRequests []time.Time
			for _, reqTime := range clientRequests {
				if now.Sub(reqTime) < window {
					validRequests = append(validRequests, reqTime)
				}
			}
			requests[clientIP] = validRequests
		}

		// 检查是否超过限制
		if len(requests[clientIP]) >= maxRequests {
			TooManyRequestsResponse(c, "请求过于频繁，请稍后再试", map[string]interface{}{
				"limit":     maxRequests,
				"window":    window.String(),
				"remaining": 0,
			})
			c.Abort()
			return
		}

		// 记录当前请求
		requests[clientIP] = append(requests[clientIP], now)
		c.Next()
	}
}

// CORSHandler CORS处理中间件
func CORSHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// 设置CORS头
		c.Header("Access-Control-Allow-Origin", origin)
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization, X-Requested-With")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Max-Age", "86400")

		// 处理预检请求
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

// SecurityHeadersHandler 安全头处理中间件
func SecurityHeadersHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 设置安全头
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "DENY")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Header("Content-Security-Policy", "default-src 'self'")

		c.Next()
	}
}

// RequestIDHandler 请求ID处理中间件
func RequestIDHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}

		c.Header("X-Request-ID", requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}

// HealthCheckHandler 健康检查中间件
func HealthCheckHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.URL.Path == "/health" {
			SuccessResponse(c, "服务正常", map[string]interface{}{
				"status":    "healthy",
				"timestamp": time.Now().Unix(),
				"version":   "1.0.0",
			})
			c.Abort()
			return
		}
		c.Next()
	}
}

// 辅助函数

// generateRequestID 生成请求ID
func generateRequestID() string {
	return fmt.Sprintf("%d-%d", time.Now().UnixNano(), runtime.NumGoroutine())
}

// isPrivateIP 检查是否为私有IP
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}

	privateBlocks := []*net.IPNet{
		{IP: net.IPv4(10, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
		{IP: net.IPv4(172, 16, 0, 0), Mask: net.CIDRMask(12, 32)},
		{IP: net.IPv4(192, 168, 0, 0), Mask: net.CIDRMask(16, 32)},
		{IP: net.IPv4(127, 0, 0, 0), Mask: net.CIDRMask(8, 32)},
	}

	for _, block := range privateBlocks {
		if block.Contains(parsedIP) {
			return true
		}
	}

	return false
}

// getRealIP 获取真实IP
func getRealIP(c *gin.Context) string {
	// 检查X-Forwarded-For头
	if xff := c.GetHeader("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		for _, ip := range ips {
			ip = strings.TrimSpace(ip)
			if ip != "" && !isPrivateIP(ip) {
				return ip
			}
		}
	}

	// 检查X-Real-IP头
	if xri := c.GetHeader("X-Real-IP"); xri != "" {
		return xri
	}

	// 使用客户端IP
	return c.ClientIP()
}
