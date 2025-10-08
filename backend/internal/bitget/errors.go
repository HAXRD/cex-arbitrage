package bitget

import (
	"fmt"
	"net/http"
	"time"
)

// BitgetError 表示 BitGet API 错误
type BitgetError struct {
	Code        string    `json:"code"`
	Message     string    `json:"msg"`
	Data        string    `json:"data,omitempty"`
	RequestTime int64     `json:"requestTime,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
	Retryable   bool      `json:"retryable"` // 是否可重试
}

func (e *BitgetError) Error() string {
	if e.Data != "" {
		return fmt.Sprintf("BitGet API Error [%s]: %s (data: %s)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("BitGet API Error [%s]: %s", e.Code, e.Message)
}

// IsRetryable 检查错误是否可重试
func (e *BitgetError) IsRetryable() bool {
	return e.Retryable
}

// NewBitgetError 创建新的 BitGet 错误
func NewBitgetError(code, message string) *BitgetError {
	return &BitgetError{
		Code:      code,
		Message:   message,
		Timestamp: time.Now(),
		Retryable: isRetryableError(code),
	}
}

// NewBitgetErrorWithData 创建带数据的 BitGet 错误
func NewBitgetErrorWithData(code, message, data string) *BitgetError {
	return &BitgetError{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: time.Now(),
		Retryable: isRetryableError(code),
	}
}

// isRetryableError 判断错误码是否可重试
func isRetryableError(code string) bool {
	switch code {
	case CodeRateLimit, CodeSystemError, CodeMaintenance:
		return true
	case CodeInvalidParam, CodeMissingParam, CodeInvalidAPIKey,
		CodeInvalidSignature, CodeInvalidTimestamp, CodeInvalidSymbol,
		CodeInvalidGranularity, CodeInvalidTimeRange:
		return false
	default:
		// 对于未知错误码，默认不可重试
		return false
	}
}

// 预定义错误类型
var (
	// 网络相关错误
	ErrNetworkTimeout     = fmt.Errorf("network timeout")
	ErrNetworkUnavailable = fmt.Errorf("network unavailable")
	ErrConnectionFailed   = fmt.Errorf("connection failed")

	// 速率限制错误
	ErrRateLimitExceeded = fmt.Errorf("rate limit exceeded")

	// API 相关错误
	ErrInvalidAPIKey      = fmt.Errorf("invalid API key")
	ErrInvalidSignature   = fmt.Errorf("invalid signature")
	ErrInvalidTimestamp   = fmt.Errorf("invalid timestamp")
	ErrInvalidSymbol      = fmt.Errorf("invalid symbol")
	ErrInvalidGranularity = fmt.Errorf("invalid granularity")
	ErrInvalidTimeRange   = fmt.Errorf("invalid time range")

	// WebSocket 相关错误
	ErrWebSocketClosed      = fmt.Errorf("websocket connection closed")
	ErrWebSocketTimeout     = fmt.Errorf("websocket timeout")
	ErrSubscriptionFailed   = fmt.Errorf("subscription failed")
	ErrUnsubscriptionFailed = fmt.Errorf("unsubscription failed")

	// 数据解析错误
	ErrInvalidJSON          = fmt.Errorf("invalid JSON response")
	ErrInvalidDataFormat    = fmt.Errorf("invalid data format")
	ErrMissingRequiredField = fmt.Errorf("missing required field")

	// 配置错误
	ErrInvalidConfig = fmt.Errorf("invalid configuration")
	ErrMissingConfig = fmt.Errorf("missing required configuration")

	// 重连相关错误
	ErrReconnectFailed      = fmt.Errorf("reconnect failed")
	ErrMaxReconnectAttempts = fmt.Errorf("max reconnect attempts reached")
	ErrReconnectDisabled    = fmt.Errorf("reconnect disabled")

	// 数据验证错误
	ErrInvalidPrice   = fmt.Errorf("invalid price")
	ErrInvalidVolume  = fmt.Errorf("invalid volume")
	ErrDataOutOfRange = fmt.Errorf("data out of range")

	// 业务逻辑错误
	ErrSymbolNotFound   = fmt.Errorf("symbol not found")
	ErrMarketClosed     = fmt.Errorf("market closed")
	ErrInsufficientData = fmt.Errorf("insufficient data")
	ErrDataExpired      = fmt.Errorf("data expired")
)

// 常见 BitGet API 错误码
const (
	// 成功
	CodeSuccess = "00000"

	// 参数错误
	CodeInvalidParam = "40001"
	CodeMissingParam = "40002"

	// 认证错误
	CodeInvalidAPIKey    = "40003"
	CodeInvalidSignature = "40004"
	CodeInvalidTimestamp = "40005"

	// 业务错误
	CodeInvalidSymbol      = "40006"
	CodeInvalidGranularity = "40007"
	CodeInvalidTimeRange   = "40008"

	// 系统错误
	CodeSystemError = "50000"
	CodeRateLimit   = "50001"
	CodeMaintenance = "50002"
)

// IsBitgetError 检查是否为 BitGet API 错误
func IsBitgetError(err error) bool {
	_, ok := err.(*BitgetError)
	return ok
}

// ParseHTTPError 解析 HTTP 错误
func ParseHTTPError(resp *http.Response) error {
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}

	switch resp.StatusCode {
	case 400:
		return fmt.Errorf("bad request: %d", resp.StatusCode)
	case 401:
		return ErrInvalidAPIKey
	case 403:
		return ErrInvalidSignature
	case 429:
		return ErrRateLimitExceeded
	case 500:
		return fmt.Errorf("internal server error: %d", resp.StatusCode)
	case 502:
		return fmt.Errorf("bad gateway: %d", resp.StatusCode)
	case 503:
		return fmt.Errorf("service unavailable: %d", resp.StatusCode)
	default:
		return fmt.Errorf("HTTP error: %d", resp.StatusCode)
	}
}

// ErrorContext 错误上下文
type ErrorContext struct {
	Operation string
	Symbol    string
	RequestID string
	Timestamp time.Time
	Details   map[string]interface{}
}

// NewErrorContext 创建错误上下文
func NewErrorContext(operation, symbol string) *ErrorContext {
	return &ErrorContext{
		Operation: operation,
		Symbol:    symbol,
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
	}
}

// WithRequestID 设置请求ID
func (ec *ErrorContext) WithRequestID(requestID string) *ErrorContext {
	ec.RequestID = requestID
	return ec
}

// WithDetail 添加详细信息
func (ec *ErrorContext) WithDetail(key string, value interface{}) *ErrorContext {
	ec.Details[key] = value
	return ec
}

// WrapError 包装错误并添加上下文
func WrapError(err error, context *ErrorContext) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s [%s] %s: %w",
		context.Operation,
		context.Symbol,
		context.Timestamp.Format(time.RFC3339),
		err)
}

// IsRetryableError 检查错误是否可重试
func IsRetryableError(err error) bool {
	if bitgetErr, ok := err.(*BitgetError); ok {
		return bitgetErr.IsRetryable()
	}

	// 检查预定义的可重试错误
	switch err {
	case ErrNetworkTimeout, ErrNetworkUnavailable, ErrConnectionFailed,
		ErrRateLimitExceeded, ErrWebSocketTimeout:
		return true
	default:
		return false
	}
}

// GetErrorCode 获取错误码
func GetErrorCode(err error) string {
	if bitgetErr, ok := err.(*BitgetError); ok {
		return bitgetErr.Code
	}
	return "UNKNOWN"
}

// GetErrorMessage 获取错误消息
func GetErrorMessage(err error) string {
	if bitgetErr, ok := err.(*BitgetError); ok {
		return bitgetErr.Message
	}
	return err.Error()
}
