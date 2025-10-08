package bitget

import (
	"context"
	"errors"
	"testing"
	"time"
)

// TestLogger 测试日志记录器
func TestLogger(t *testing.T) {
	// 创建日志记录器
	logger := NewLogger(LogLevelDebug, true)
	if logger == nil {
		t.Fatal("日志记录器不应该为 nil")
	}

	// 测试基本日志记录
	logger.Info("test info message")
	logger.Debug("test debug message")
	logger.Warn("test warning message")
	logger.Error("test error message")
}

// TestLoggerWithContext 测试带上下文的日志记录
func TestLoggerWithContext(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 创建上下文
	ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
	ctx = context.WithValue(ctx, "span_id", "span-456")

	// 测试带上下文的日志记录器
	contextLogger := logger.WithContext(ctx)
	if contextLogger == nil {
		t.Fatal("带上下文的日志记录器不应该为 nil")
	}

	contextLogger.Info("test message with context")
}

// TestLoggerWithField 测试带字段的日志记录
func TestLoggerWithField(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试单个字段
	fieldLogger := logger.WithField("symbol", "BTCUSDT")
	if fieldLogger == nil {
		t.Fatal("带字段的日志记录器不应该为 nil")
	}

	fieldLogger.Info("test message with field")
}

// TestLoggerWithFields 测试带多个字段的日志记录
func TestLoggerWithFields(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试多个字段
	fields := map[string]interface{}{
		"symbol":    "BTCUSDT",
		"operation": "GetContractSymbols",
		"duration":  time.Millisecond * 100,
	}

	fieldsLogger := logger.WithFields(fields)
	if fieldsLogger == nil {
		t.Fatal("带多个字段的日志记录器不应该为 nil")
	}

	fieldsLogger.Info("test message with multiple fields")
}

// TestLoggerWithError 测试带错误的日志记录
func TestLoggerWithError(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试 BitGet 错误
	bitgetErr := NewBitgetError("40001", "Invalid parameter")
	errorLogger := logger.WithError(bitgetErr)
	if errorLogger == nil {
		t.Fatal("带错误的日志记录器不应该为 nil")
	}

	errorLogger.Error("test message with BitGet error")

	// 测试普通错误
	normalErr := errors.New("normal error")
	normalErrorLogger := logger.WithError(normalErr)
	if normalErrorLogger == nil {
		t.Fatal("带普通错误的日志记录器不应该为 nil")
	}

	normalErrorLogger.Error("test message with normal error")

	// 测试 nil 错误
	nilErrorLogger := logger.WithError(nil)
	if nilErrorLogger == nil {
		t.Fatal("带 nil 错误的日志记录器不应该为 nil")
	}

	nilErrorLogger.Info("test message with nil error")
}

// TestLoggerAPIRequest 测试 API 请求日志
func TestLoggerAPIRequest(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试 API 请求日志
	logger.APIRequest("GET", "https://api.bitget.com/api/v2/mix/market/contracts",
		time.Millisecond*150, 200)
}

// TestLoggerAPIError 测试 API 错误日志
func TestLoggerAPIError(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试 API 错误日志
	err := NewBitgetError("40001", "Invalid parameter")
	logger.APIError("GET", "https://api.bitget.com/api/v2/mix/market/contracts",
		err, time.Millisecond*150)
}

// TestLoggerWebSocketEvent 测试 WebSocket 事件日志
func TestLoggerWebSocketEvent(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试 WebSocket 事件日志
	details := map[string]interface{}{
		"symbol": "BTCUSDT",
		"type":   "ticker",
		"count":  1,
	}

	logger.WebSocketEvent("message_received", details)
}

// TestLoggerWebSocketError 测试 WebSocket 错误日志
func TestLoggerWebSocketError(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试 WebSocket 错误日志
	err := ErrWebSocketTimeout
	details := map[string]interface{}{
		"symbol": "BTCUSDT",
		"type":   "ticker",
		"retry":  true,
	}

	logger.WebSocketError("connection_lost", err, details)
}

// TestLoggerReconnectEvent 测试重连事件日志
func TestLoggerReconnectEvent(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试重连事件日志
	logger.ReconnectEvent(1, 3, time.Second*2, false)
	logger.ReconnectEvent(2, 3, time.Second*4, true)
}

// TestLoggerDataReceived 测试数据接收日志
func TestLoggerDataReceived(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试数据接收日志
	logger.DataReceived("ticker", 10, time.Millisecond*50)
	logger.DataReceived("kline", 100, time.Millisecond*200)
}

// TestLoggerPerformanceMetric 测试性能指标日志
func TestLoggerPerformanceMetric(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试性能指标日志
	logger.PerformanceMetric("response_time", 150.5, "ms")
	logger.PerformanceMetric("throughput", 1000.0, "requests/s")
	logger.PerformanceMetric("error_rate", 0.01, "%")
}

// TestLogLevel 测试日志级别
func TestLogLevel(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{LogLevelDebug, "debug"},
		{LogLevelInfo, "info"},
		{LogLevelWarn, "warn"},
		{LogLevelError, "error"},
		{LogLevelFatal, "fatal"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			if tt.level.String() != tt.expected {
				t.Errorf("期望日志级别 = %s, 实际 = %s", tt.expected, tt.level.String())
			}
		})
	}
}

// TestLoggerChaining 测试日志记录器链式调用
func TestLoggerChaining(t *testing.T) {
	logger := NewLogger(LogLevelDebug, true)

	// 测试链式调用
	chainedLogger := logger.
		WithField("symbol", "BTCUSDT").
		WithField("operation", "GetContractSymbols").
		WithError(NewBitgetError("40001", "Invalid parameter"))

	if chainedLogger == nil {
		t.Fatal("链式调用的日志记录器不应该为 nil")
	}

	chainedLogger.Error("test chained logger")
}

// TestLoggerContextValues 测试上下文值提取
func TestLoggerContextValues(t *testing.T) {
	// 测试有追踪ID的上下文
	ctx := context.WithValue(context.Background(), "trace_id", "trace-123")
	traceID := getTraceID(ctx)
	if traceID != "trace-123" {
		t.Errorf("期望追踪ID = trace-123, 实际 = %s", traceID)
	}

	// 测试有跨度ID的上下文
	ctx = context.WithValue(ctx, "span_id", "span-456")
	spanID := getSpanID(ctx)
	if spanID != "span-456" {
		t.Errorf("期望跨度ID = span-456, 实际 = %s", spanID)
	}

	// 测试没有追踪ID的上下文
	emptyCtx := context.Background()
	emptyTraceID := getTraceID(emptyCtx)
	if emptyTraceID != "" {
		t.Errorf("期望空追踪ID = \"\", 实际 = %s", emptyTraceID)
	}

	// 测试类型不匹配的上下文
	wrongTypeCtx := context.WithValue(context.Background(), "trace_id", 123)
	wrongTypeTraceID := getTraceID(wrongTypeCtx)
	if wrongTypeTraceID != "" {
		t.Errorf("期望类型不匹配的追踪ID = \"\", 实际 = %s", wrongTypeTraceID)
	}
}
