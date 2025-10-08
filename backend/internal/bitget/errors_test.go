package bitget

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// TestBitgetError 测试 BitGet 错误
func TestBitgetError(t *testing.T) {
	// 测试基本错误
	err := NewBitgetError("40001", "Invalid parameter")
	if err.Code != "40001" {
		t.Errorf("期望错误码 = 40001, 实际 = %s", err.Code)
	}
	if err.Message != "Invalid parameter" {
		t.Errorf("期望错误消息 = Invalid parameter, 实际 = %s", err.Message)
	}
	if err.Retryable {
		t.Error("参数错误不应该可重试")
	}

	// 测试带数据的错误
	errWithData := NewBitgetErrorWithData("50001", "Rate limit exceeded", "retry after 60s")
	if errWithData.Data != "retry after 60s" {
		t.Errorf("期望错误数据 = retry after 60s, 实际 = %s", errWithData.Data)
	}
	if !errWithData.Retryable {
		t.Error("速率限制错误应该可重试")
	}

	// 测试错误字符串
	expectedError := "BitGet API Error [40001]: Invalid parameter"
	if err.Error() != expectedError {
		t.Errorf("期望错误字符串 = %s, 实际 = %s", expectedError, err.Error())
	}
}

// TestIsRetryableError 测试错误重试性
func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "可重试的 BitGet 错误",
			err:      NewBitgetError(CodeRateLimit, "Rate limit exceeded"),
			expected: true,
		},
		{
			name:     "不可重试的 BitGet 错误",
			err:      NewBitgetError(CodeInvalidParam, "Invalid parameter"),
			expected: false,
		},
		{
			name:     "网络超时错误",
			err:      ErrNetworkTimeout,
			expected: true,
		},
		{
			name:     "连接失败错误",
			err:      ErrConnectionFailed,
			expected: true,
		},
		{
			name:     "速率限制错误",
			err:      ErrRateLimitExceeded,
			expected: true,
		},
		{
			name:     "WebSocket 超时错误",
			err:      ErrWebSocketTimeout,
			expected: true,
		},
		{
			name:     "普通错误",
			err:      errors.New("some error"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError(%v) = %v, 期望 %v", tt.err, result, tt.expected)
			}
		})
	}
}

// TestErrorContext 测试错误上下文
func TestErrorContext(t *testing.T) {
	// 创建错误上下文
	context := NewErrorContext("GetContractSymbols", "BTCUSDT").
		WithRequestID("req-123").
		WithDetail("url", "https://api.bitget.com/api/v2/mix/market/contracts").
		WithDetail("status_code", 200)

	if context.Operation != "GetContractSymbols" {
		t.Errorf("期望操作 = GetContractSymbols, 实际 = %s", context.Operation)
	}
	if context.Symbol != "BTCUSDT" {
		t.Errorf("期望交易对 = BTCUSDT, 实际 = %s", context.Symbol)
	}
	if context.RequestID != "req-123" {
		t.Errorf("期望请求ID = req-123, 实际 = %s", context.RequestID)
	}
	if context.Details["url"] != "https://api.bitget.com/api/v2/mix/market/contracts" {
		t.Errorf("期望URL不匹配")
	}
	if context.Details["status_code"] != 200 {
		t.Errorf("期望状态码 = 200, 实际 = %v", context.Details["status_code"])
	}
}

// TestWrapError 测试错误包装
func TestWrapError(t *testing.T) {
	originalErr := errors.New("original error")
	context := NewErrorContext("TestOperation", "BTCUSDT")

	wrappedErr := WrapError(originalErr, context)
	if wrappedErr == nil {
		t.Error("包装后的错误不应该为 nil")
	}

	// 检查错误消息格式
	errorMsg := wrappedErr.Error()
	if errorMsg == "" {
		t.Error("错误消息不应该为空")
	}
}

// TestGetErrorCode 测试获取错误码
func TestGetErrorCode(t *testing.T) {
	// 测试 BitGet 错误
	bitgetErr := NewBitgetError("40001", "Invalid parameter")
	code := GetErrorCode(bitgetErr)
	if code != "40001" {
		t.Errorf("期望错误码 = 40001, 实际 = %s", code)
	}

	// 测试普通错误
	normalErr := errors.New("normal error")
	code = GetErrorCode(normalErr)
	if code != "UNKNOWN" {
		t.Errorf("期望错误码 = UNKNOWN, 实际 = %s", code)
	}
}

// TestGetErrorMessage 测试获取错误消息
func TestGetErrorMessage(t *testing.T) {
	// 测试 BitGet 错误
	bitgetErr := NewBitgetError("40001", "Invalid parameter")
	message := GetErrorMessage(bitgetErr)
	if message != "Invalid parameter" {
		t.Errorf("期望错误消息 = Invalid parameter, 实际 = %s", message)
	}

	// 测试普通错误
	normalErr := errors.New("normal error")
	message = GetErrorMessage(normalErr)
	if message != "normal error" {
		t.Errorf("期望错误消息 = normal error, 实际 = %s", message)
	}
}

// TestParseHTTPError 测试 HTTP 错误解析
func TestParseHTTPError(t *testing.T) {
	tests := []struct {
		name            string
		statusCode      int
		shouldHaveError bool
		expectedError   error
	}{
		{
			name:            "成功状态码",
			statusCode:      200,
			shouldHaveError: false,
			expectedError:   nil,
		},
		{
			name:            "400 错误",
			statusCode:      400,
			shouldHaveError: true,
			expectedError:   nil,
		},
		{
			name:            "401 错误",
			statusCode:      401,
			shouldHaveError: true,
			expectedError:   ErrInvalidAPIKey,
		},
		{
			name:            "403 错误",
			statusCode:      403,
			shouldHaveError: true,
			expectedError:   ErrInvalidSignature,
		},
		{
			name:            "429 错误",
			statusCode:      429,
			shouldHaveError: true,
			expectedError:   ErrRateLimitExceeded,
		},
		{
			name:            "500 错误",
			statusCode:      500,
			shouldHaveError: true,
			expectedError:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建模拟响应
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			// 创建请求
			req, _ := http.NewRequest("GET", server.URL, nil)
			resp, _ := http.DefaultClient.Do(req)
			defer resp.Body.Close()

			// 解析错误
			err := ParseHTTPError(resp)

			if !tt.shouldHaveError {
				if err != nil {
					t.Errorf("期望无错误，实际 = %v", err)
				}
			} else {
				if err == nil {
					t.Error("期望有错误，实际为 nil")
				} else if tt.expectedError != nil && err != tt.expectedError {
					t.Errorf("期望错误 = %v, 实际 = %v", tt.expectedError, err)
				}
			}
		})
	}
}

// TestIsBitgetError 测试 BitGet 错误检查
func TestIsBitgetError(t *testing.T) {
	// 测试 BitGet 错误
	bitgetErr := NewBitgetError("40001", "Invalid parameter")
	if !IsBitgetError(bitgetErr) {
		t.Error("BitGet 错误应该被识别")
	}

	// 测试普通错误
	normalErr := errors.New("normal error")
	if IsBitgetError(normalErr) {
		t.Error("普通错误不应该被识别为 BitGet 错误")
	}
}

// TestErrorTimestamp 测试错误时间戳
func TestErrorTimestamp(t *testing.T) {
	err := NewBitgetError("40001", "Invalid parameter")

	// 检查时间戳是否在合理范围内
	now := time.Now()
	diff := now.Sub(err.Timestamp)
	if diff < 0 || diff > time.Second {
		t.Errorf("错误时间戳不在合理范围内: %v", err.Timestamp)
	}
}

// TestErrorRetryableLogic 测试错误重试逻辑
func TestErrorRetryableLogic(t *testing.T) {
	tests := []struct {
		code     string
		expected bool
	}{
		{CodeSuccess, false},
		{CodeInvalidParam, false},
		{CodeMissingParam, false},
		{CodeInvalidAPIKey, false},
		{CodeInvalidSignature, false},
		{CodeInvalidTimestamp, false},
		{CodeInvalidSymbol, false},
		{CodeInvalidGranularity, false},
		{CodeInvalidTimeRange, false},
		{CodeSystemError, true},
		{CodeRateLimit, true},
		{CodeMaintenance, true},
		{"UNKNOWN", false},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			err := NewBitgetError(tt.code, "test message")
			if err.Retryable != tt.expected {
				t.Errorf("错误码 %s 的重试性 = %v, 期望 %v", tt.code, err.Retryable, tt.expected)
			}
		})
	}
}
