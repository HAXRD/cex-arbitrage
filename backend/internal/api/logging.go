package api

import (
	"bytes"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LoggingConfig 日志配置
type LoggingConfig struct {
	SkipPaths       []string      // 跳过的路径
	SkipHeaders     []string      // 跳过的请求头
	LogRequestBody  bool          // 是否记录请求体
	LogResponseBody bool          // 是否记录响应体
	MaxBodySize     int64         // 最大记录体大小
	Level           zapcore.Level // 日志级别
}

// DefaultLoggingConfig 默认日志配置
func DefaultLoggingConfig() LoggingConfig {
	return LoggingConfig{
		SkipPaths: []string{
			"/health",
			"/metrics",
			"/favicon.ico",
		},
		SkipHeaders: []string{
			"Authorization",
			"Cookie",
			"X-API-Key",
		},
		LogRequestBody:  false,
		LogResponseBody: false,
		MaxBodySize:     1024, // 1KB
		Level:           zapcore.InfoLevel,
	}
}

// RequestLogger 请求日志中间件
func RequestLogger(logger *zap.Logger, config LoggingConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否跳过此路径
		if shouldSkipPath(c.Request.URL.Path, config.SkipPaths) {
			c.Next()
			return
		}

		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// 记录请求信息
		requestID := c.GetString("request_id")
		clientIP := getRealIP(c)
		userAgent := c.Request.UserAgent()

		// 记录请求体（如果需要）
		var requestBody []byte
		if config.LogRequestBody && c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建响应写入器包装器
		responseWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = responseWriter

		// 处理请求
		c.Next()

		// 计算处理时间
		latency := time.Since(start)

		// 记录响应信息
		statusCode := c.Writer.Status()
		bodySize := c.Writer.Size()

		// 构建日志字段
		fields := []zap.Field{
			zap.String("request_id", requestID),
			zap.String("method", c.Request.Method),
			zap.String("path", path),
			zap.String("query", raw),
			zap.String("ip", clientIP),
			zap.String("user_agent", userAgent),
			zap.Int("status", statusCode),
			zap.Int("body_size", bodySize),
			zap.Duration("latency", latency),
		}

		// 添加请求头（过滤敏感信息）
		if headers := getFilteredHeaders(c.Request.Header, config.SkipHeaders); len(headers) > 0 {
			fields = append(fields, zap.Any("headers", headers))
		}

		// 添加请求体（如果配置了且大小合适）
		if config.LogRequestBody && len(requestBody) > 0 && int64(len(requestBody)) <= config.MaxBodySize {
			fields = append(fields, zap.String("request_body", string(requestBody)))
		}

		// 添加响应体（如果配置了且大小合适）
		if config.LogResponseBody && responseWriter.body.Len() > 0 && int64(responseWriter.body.Len()) <= config.MaxBodySize {
			fields = append(fields, zap.String("response_body", responseWriter.body.String()))
		}

		// 根据状态码选择日志级别
		var logFunc func(string, ...zap.Field)
		switch {
		case statusCode >= 500:
			logFunc = logger.Error
		case statusCode >= 400:
			logFunc = logger.Warn
		default:
			logFunc = logger.Info
		}

		// 记录日志
		logFunc("HTTP Request",
			fields...,
		)
	}
}

// AccessLogger 访问日志中间件（简化版）
func AccessLogger(logger *zap.Logger) gin.HandlerFunc {
	return gin.LoggerWithFormatter(func(param gin.LogFormatterParams) string {
		// 使用结构化日志而不是格式化字符串
		logger.Info("HTTP Access",
			zap.String("method", param.Method),
			zap.String("path", param.Path),
			zap.String("ip", param.ClientIP),
			zap.Int("status", param.StatusCode),
			zap.Duration("latency", param.Latency),
			zap.String("user_agent", param.Request.UserAgent()),
		)
		return ""
	})
}

// ErrorLogger 错误日志中间件
func ErrorLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 只记录错误状态码
		if c.Writer.Status() >= 400 {
			requestID := c.GetString("request_id")
			clientIP := getRealIP(c)

			logger.Error("HTTP Error",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", clientIP),
				zap.Int("status", c.Writer.Status()),
				zap.String("error", c.Errors.String()),
			)
		}
	}
}

// PerformanceLogger 性能日志中间件
func PerformanceLogger(logger *zap.Logger, slowThreshold time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()
		latency := time.Since(start)

		// 记录慢请求
		if latency > slowThreshold {
			requestID := c.GetString("request_id")
			clientIP := getRealIP(c)

			logger.Warn("Slow Request",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", clientIP),
				zap.Duration("latency", latency),
				zap.Duration("threshold", slowThreshold),
			)
		}
	}
}

// SecurityLogger 安全日志中间件
func SecurityLogger(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientIP := getRealIP(c)
		userAgent := c.Request.UserAgent()
		referer := c.Request.Referer()

		// 检查可疑请求
		if isSuspiciousRequest(c) {
			requestID := c.GetString("request_id")

			logger.Warn("Suspicious Request",
				zap.String("request_id", requestID),
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.String("ip", clientIP),
				zap.String("user_agent", userAgent),
				zap.String("referer", referer),
				zap.String("query", c.Request.URL.RawQuery),
			)
		}

		c.Next()
	}
}

// responseWriter 响应写入器包装器
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// 辅助函数

// shouldSkipPath 检查是否应该跳过路径
func shouldSkipPath(path string, skipPaths []string) bool {
	for _, skipPath := range skipPaths {
		if path == skipPath {
			return true
		}
	}
	return false
}

// getFilteredHeaders 获取过滤后的请求头
func getFilteredHeaders(headers http.Header, skipHeaders []string) map[string]string {
	filtered := make(map[string]string)

	for name, values := range headers {
		// 检查是否应该跳过此头
		shouldSkip := false
		for _, skipHeader := range skipHeaders {
			if name == skipHeader {
				shouldSkip = true
				break
			}
		}

		if !shouldSkip && len(values) > 0 {
			filtered[name] = values[0]
		}
	}

	return filtered
}

// isSuspiciousRequest 检查是否为可疑请求
func isSuspiciousRequest(c *gin.Context) bool {
	// 检查SQL注入模式
	query := c.Request.URL.RawQuery
	if containsSQLInjectionPattern(query) {
		return true
	}

	// 检查路径遍历模式
	path := c.Request.URL.Path
	if containsPathTraversalPattern(path) {
		return true
	}

	// 检查异常长的URL
	if len(c.Request.URL.String()) > 2048 {
		return true
	}

	// 检查异常多的查询参数
	if len(c.Request.URL.Query()) > 50 {
		return true
	}

	return false
}

// containsSQLInjectionPattern 检查是否包含SQL注入模式
func containsSQLInjectionPattern(input string) bool {
	patterns := []string{
		"' OR '1'='1",
		"'; DROP TABLE",
		"UNION SELECT",
		"INSERT INTO",
		"DELETE FROM",
		"UPDATE SET",
	}

	for _, pattern := range patterns {
		if containsIgnoreCase(input, pattern) {
			return true
		}
	}

	return false
}

// containsPathTraversalPattern 检查是否包含路径遍历模式
func containsPathTraversalPattern(input string) bool {
	patterns := []string{
		"../",
		"..\\",
		"/etc/passwd",
		"\\windows\\system32",
	}

	for _, pattern := range patterns {
		if containsIgnoreCase(input, pattern) {
			return true
		}
	}

	return false
}

// containsIgnoreCase 忽略大小写检查包含
func containsIgnoreCase(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexSubstringIgnoreCase(s, substr) >= 0))
}

// indexSubstringIgnoreCase 忽略大小写查找子字符串
func indexSubstringIgnoreCase(s, substr string) int {
	s = strings.ToLower(s)
	substr = strings.ToLower(substr)

	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
