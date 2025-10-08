package bitget

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"go.uber.org/zap"
	"golang.org/x/time/rate"
)

// BitgetClient BitGet API 客户端接口
type BitgetClient interface {
	// REST API 方法
	GetContractSymbols(ctx context.Context) ([]Symbol, error)
	GetKlines(ctx context.Context, req KlineRequest) ([]Kline, error)
	GetContractInfo(ctx context.Context, symbol string) (*Ticker, error)

	// 配置方法
	SetConfig(config BitgetConfig)
	GetConfig() BitgetConfig

	// 关闭客户端
	Close() error
}

// client BitGet API 客户端实现
type client struct {
	config     BitgetConfig
	httpClient *http.Client
	limiter    *rate.Limiter
	logger     *zap.Logger
}

// NewClient 创建新的 BitGet 客户端
func NewClient(config BitgetConfig, logger *zap.Logger) BitgetClient {
	// 创建 HTTP 客户端
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	// 创建速率限制器
	limiter := rate.NewLimiter(rate.Limit(config.RateLimit), 1)

	return &client{
		config:     config,
		httpClient: httpClient,
		limiter:    limiter,
		logger:     logger,
	}
}

// SetConfig 设置配置
func (c *client) SetConfig(config BitgetConfig) {
	c.config = config
	c.httpClient.Timeout = config.Timeout
	c.limiter = rate.NewLimiter(rate.Limit(config.RateLimit), 1)
}

// GetConfig 获取配置
func (c *client) GetConfig() BitgetConfig {
	return c.config
}

// Close 关闭客户端
func (c *client) Close() error {
	// HTTP 客户端会自动关闭连接
	return nil
}

// makeRequest 发送 HTTP 请求的通用方法
func (c *client) makeRequest(ctx context.Context, method, url string, body interface{}) (*http.Response, error) {
	// 等待速率限制
	if err := c.limiter.Wait(ctx); err != nil {
		return nil, fmt.Errorf("rate limiter wait failed: %w", err)
	}

	// 准备请求体
	var reqBody io.Reader
	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewBuffer(jsonData)
	}

	// 创建请求
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "BitGet-Go-Client/1.0")

	// 记录请求日志
	c.logger.Debug("sending request",
		zap.String("method", method),
		zap.String("url", url),
	)

	// 发送请求
	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.logger.Error("request failed",
			zap.String("method", method),
			zap.String("url", url),
			zap.Error(err),
		)
		return nil, fmt.Errorf("http request failed: %w", err)
	}

	// 记录响应日志
	c.logger.Debug("received response",
		zap.String("method", method),
		zap.String("url", url),
		zap.Int("status", resp.StatusCode),
	)

	return resp, nil
}

// parseResponse 解析 HTTP 响应的通用方法
func (c *client) parseResponse(resp *http.Response, result interface{}) error {
	defer resp.Body.Close()

	// 读取响应体
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.Error("failed to read response body",
			zap.String("method", resp.Request.Method),
			zap.String("url", resp.Request.URL.String()),
			zap.Int("status", resp.StatusCode),
			zap.Error(err),
		)
		return fmt.Errorf("read response body: %w", err)
	}

	// 记录响应体日志
	c.logger.Debug("response body",
		zap.String("method", resp.Request.Method),
		zap.String("url", resp.Request.URL.String()),
		zap.Int("status", resp.StatusCode),
		zap.String("body", string(body)),
	)

	// 解析 JSON
	if err := json.Unmarshal(body, result); err != nil {
		c.logger.Error("failed to unmarshal response",
			zap.String("method", resp.Request.Method),
			zap.String("url", resp.Request.URL.String()),
			zap.Int("status", resp.StatusCode),
			zap.String("body", string(body)),
			zap.Error(err),
		)
		return fmt.Errorf("unmarshal response: %w", err)
	}

	// 注意：不在这里检查 HTTP 状态码，让调用方处理 API 错误

	return nil
}

// buildURL 构建 API URL
func (c *client) buildURL(endpoint string) string {
	return fmt.Sprintf("%s%s", c.config.RestBaseURL, endpoint)
}
