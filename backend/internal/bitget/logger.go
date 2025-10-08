package bitget

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// LogLevel 日志级别
type LogLevel int

const (
	LogLevelDebug LogLevel = iota
	LogLevelInfo
	LogLevelWarn
	LogLevelError
	LogLevelFatal
)

// String 返回日志级别的字符串表示
func (l LogLevel) String() string {
	switch l {
	case LogLevelDebug:
		return "debug"
	case LogLevelInfo:
		return "info"
	case LogLevelWarn:
		return "warn"
	case LogLevelError:
		return "error"
	case LogLevelFatal:
		return "fatal"
	default:
		return "unknown"
	}
}

// Logger 封装 zap.Logger 提供更友好的接口
type Logger struct {
	*zap.Logger
	context map[string]interface{}
}

// NewLogger 创建新的日志记录器
func NewLogger(level LogLevel, development bool) *Logger {
	var config zap.Config
	if development {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}

	// 设置日志级别
	config.Level = zap.NewAtomicLevelAt(zapcore.Level(level))

	logger, err := config.Build()
	if err != nil {
		panic("failed to create logger: " + err.Error())
	}

	return &Logger{
		Logger:  logger,
		context: make(map[string]interface{}),
	}
}

// WithContext 添加上下文信息
func (l *Logger) WithContext(ctx context.Context) *Logger {
	newLogger := l.Logger.With(
		zap.String("trace_id", getTraceID(ctx)),
		zap.String("span_id", getSpanID(ctx)),
	)

	return &Logger{
		Logger:  newLogger,
		context: l.context,
	}
}

// WithField 添加字段
func (l *Logger) WithField(key string, value interface{}) *Logger {
	newLogger := l.Logger.With(zap.Any(key, value))
	return &Logger{
		Logger:  newLogger,
		context: l.context,
	}
}

// WithFields 添加多个字段
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	zapFields := make([]zap.Field, 0, len(fields))
	for k, v := range fields {
		zapFields = append(zapFields, zap.Any(k, v))
	}

	newLogger := l.Logger.With(zapFields...)
	return &Logger{
		Logger:  newLogger,
		context: l.context,
	}
}

// WithError 添加错误信息
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}

	newLogger := l.Logger.With(
		zap.Error(err),
		zap.String("error_code", GetErrorCode(err)),
		zap.String("error_message", GetErrorMessage(err)),
		zap.Bool("retryable", IsRetryableError(err)),
	)

	return &Logger{
		Logger:  newLogger,
		context: l.context,
	}
}

// APIRequest 记录 API 请求
func (l *Logger) APIRequest(method, url string, duration time.Duration, statusCode int) {
	l.Info("API request completed",
		zap.String("method", method),
		zap.String("url", url),
		zap.Duration("duration", duration),
		zap.Int("status_code", statusCode),
	)
}

// APIError 记录 API 错误
func (l *Logger) APIError(method, url string, err error, duration time.Duration) {
	l.Error("API request failed",
		zap.String("method", method),
		zap.String("url", url),
		zap.Duration("duration", duration),
		zap.Error(err),
		zap.String("error_code", GetErrorCode(err)),
		zap.Bool("retryable", IsRetryableError(err)),
	)
}

// WebSocketEvent 记录 WebSocket 事件
func (l *Logger) WebSocketEvent(event string, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.Time("timestamp", time.Now()),
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	l.Info("WebSocket event", fields...)
}

// WebSocketError 记录 WebSocket 错误
func (l *Logger) WebSocketError(event string, err error, details map[string]interface{}) {
	fields := []zap.Field{
		zap.String("event", event),
		zap.Error(err),
		zap.String("error_code", GetErrorCode(err)),
		zap.Bool("retryable", IsRetryableError(err)),
		zap.Time("timestamp", time.Now()),
	}

	for k, v := range details {
		fields = append(fields, zap.Any(k, v))
	}

	l.Error("WebSocket error", fields...)
}

// ReconnectEvent 记录重连事件
func (l *Logger) ReconnectEvent(attempt int, maxAttempts int, delay time.Duration, success bool) {
	l.Info("reconnect event",
		zap.Int("attempt", attempt),
		zap.Int("max_attempts", maxAttempts),
		zap.Duration("delay", delay),
		zap.Bool("success", success),
		zap.Time("timestamp", time.Now()),
	)
}

// DataReceived 记录数据接收
func (l *Logger) DataReceived(dataType string, count int, duration time.Duration) {
	l.Debug("data received",
		zap.String("type", dataType),
		zap.Int("count", count),
		zap.Duration("duration", duration),
		zap.Time("timestamp", time.Now()),
	)
}

// PerformanceMetric 记录性能指标
func (l *Logger) PerformanceMetric(metric string, value float64, unit string) {
	l.Info("performance metric",
		zap.String("metric", metric),
		zap.Float64("value", value),
		zap.String("unit", unit),
		zap.Time("timestamp", time.Now()),
	)
}

// getTraceID 从上下文中获取追踪ID
func getTraceID(ctx context.Context) string {
	if traceID := ctx.Value("trace_id"); traceID != nil {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// getSpanID 从上下文中获取跨度ID
func getSpanID(ctx context.Context) string {
	if spanID := ctx.Value("span_id"); spanID != nil {
		if id, ok := spanID.(string); ok {
			return id
		}
	}
	return ""
}
