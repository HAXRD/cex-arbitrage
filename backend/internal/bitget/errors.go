package bitget

import (
	"fmt"
	"net/http"
)

// BitgetError 表示 BitGet API 错误
type BitgetError struct {
	Code    string `json:"code"`
	Message string `json:"msg"`
	Data    string `json:"data,omitempty"`
}

func (e *BitgetError) Error() string {
	if e.Data != "" {
		return fmt.Sprintf("BitGet API Error [%s]: %s (data: %s)", e.Code, e.Message, e.Data)
	}
	return fmt.Sprintf("BitGet API Error [%s]: %s", e.Code, e.Message)
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
	ErrInvalidAPIKey     = fmt.Errorf("invalid API key")
	ErrInvalidSignature  = fmt.Errorf("invalid signature")
	ErrInvalidTimestamp  = fmt.Errorf("invalid timestamp")
	ErrInvalidSymbol     = fmt.Errorf("invalid symbol")
	ErrInvalidGranularity = fmt.Errorf("invalid granularity")
	ErrInvalidTimeRange  = fmt.Errorf("invalid time range")
	
	// WebSocket 相关错误
	ErrWebSocketClosed    = fmt.Errorf("websocket connection closed")
	ErrWebSocketTimeout   = fmt.Errorf("websocket timeout")
	ErrSubscriptionFailed = fmt.Errorf("subscription failed")
	ErrUnsubscriptionFailed = fmt.Errorf("unsubscription failed")
	
	// 数据解析错误
	ErrInvalidJSON        = fmt.Errorf("invalid JSON response")
	ErrInvalidDataFormat  = fmt.Errorf("invalid data format")
	ErrMissingRequiredField = fmt.Errorf("missing required field")
	
	// 配置错误
	ErrInvalidConfig      = fmt.Errorf("invalid configuration")
	ErrMissingConfig      = fmt.Errorf("missing required configuration")
)

// 常见 BitGet API 错误码
const (
	// 成功
	CodeSuccess = "00000"
	
	// 参数错误
	CodeInvalidParam = "40001"
	CodeMissingParam  = "40002"
	
	// 认证错误
	CodeInvalidAPIKey    = "40003"
	CodeInvalidSignature = "40004"
	CodeInvalidTimestamp = "40005"
	
	// 业务错误
	CodeInvalidSymbol     = "40006"
	CodeInvalidGranularity = "40007"
	CodeInvalidTimeRange  = "40008"
	
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
