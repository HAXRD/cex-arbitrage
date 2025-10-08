package bitget

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"go.uber.org/zap"
)

// ContractSymbolsResponse 合约列表响应
type ContractSymbolsResponse struct {
	Code        string   `json:"code"`
	Msg         string   `json:"msg"`
	RequestTime int64    `json:"requestTime"`
	Data        []Symbol `json:"data"`
}

// KlinesResponse K线数据响应
type KlinesResponse struct {
	Code        string       `json:"code"`
	Msg         string       `json:"msg"`
	RequestTime int64        `json:"requestTime"`
	Data        []KlineArray `json:"data"`
}

// ContractInfoResponse 合约信息响应
type ContractInfoResponse struct {
	Code        string   `json:"code"`
	Msg         string   `json:"msg"`
	RequestTime int64    `json:"requestTime"`
	Data        []Ticker `json:"data"`
}

// GetContractSymbols 获取合约列表
func (c *client) GetContractSymbols(ctx context.Context) ([]Symbol, error) {
	endpoint := "/api/v2/mix/market/contracts"
	apiURL := c.buildURL(endpoint)

	// 构建查询参数
	params := url.Values{}
	params.Set("productType", ProductTypeUSDTFutures)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 发送请求
	resp, err := c.makeRequest(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get contract symbols: %w", err)
	}

	// 解析响应
	var result ContractSymbolsResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("parse contract symbols response: %w", err)
	}

	// 检查 API 错误
	if result.Code != CodeSuccess {
		return nil, &BitgetError{
			Code:    result.Code,
			Message: result.Msg,
		}
	}

	c.logger.Info("successfully fetched contract symbols",
		zap.Int("count", len(result.Data)),
		zap.Int64("requestTime", result.RequestTime),
	)

	return result.Data, nil
}

// GetKlines 获取 K线数据
func (c *client) GetKlines(ctx context.Context, req KlineRequest) ([]Kline, error) {
	endpoint := "/api/v2/mix/market/candles"
	apiURL := c.buildURL(endpoint)

	// 构建查询参数
	params := url.Values{}
	params.Set("symbol", req.Symbol)
	params.Set("granularity", req.Granularity)
	params.Set("productType", ProductTypeUSDTFutures)

	if req.StartTime != nil {
		params.Set("startTime", strconv.FormatInt(*req.StartTime, 10))
	}
	if req.EndTime != nil {
		params.Set("endTime", strconv.FormatInt(*req.EndTime, 10))
	}
	if req.Limit > 0 {
		params.Set("limit", strconv.Itoa(req.Limit))
	}

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 发送请求
	resp, err := c.makeRequest(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get klines: %w", err)
	}

	// 解析响应
	var result KlinesResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("parse klines response: %w", err)
	}

	// 检查 API 错误
	if result.Code != CodeSuccess {
		return nil, &BitgetError{
			Code:    result.Code,
			Message: result.Msg,
		}
	}

	// 转换数组格式为结构体格式
	klines := make([]Kline, 0, len(result.Data))
	for _, klineArray := range result.Data {
		kline := ParseKlineArray(klineArray)
		klines = append(klines, kline)
	}

	c.logger.Info("successfully fetched klines",
		zap.String("symbol", req.Symbol),
		zap.String("granularity", req.Granularity),
		zap.Int("count", len(klines)),
		zap.Int64("requestTime", result.RequestTime),
	)

	return klines, nil
}

// GetContractInfo 获取合约行情
func (c *client) GetContractInfo(ctx context.Context, symbol string) (*Ticker, error) {
	endpoint := "/api/v2/mix/market/ticker"
	apiURL := c.buildURL(endpoint)

	// 构建查询参数
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("productType", ProductTypeUSDTFutures)

	fullURL := fmt.Sprintf("%s?%s", apiURL, params.Encode())

	// 发送请求
	resp, err := c.makeRequest(ctx, http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, fmt.Errorf("get contract info: %w", err)
	}

	// 解析响应
	var result ContractInfoResponse
	if err := c.parseResponse(resp, &result); err != nil {
		return nil, fmt.Errorf("parse contract info response: %w", err)
	}

	// 检查 API 错误
	if result.Code != CodeSuccess {
		return nil, &BitgetError{
			Code:    result.Code,
			Message: result.Msg,
		}
	}

	// 检查是否有数据
	if len(result.Data) == 0 {
		return nil, &BitgetError{
			Code:    "404",
			Message: "No data found for symbol: " + symbol,
		}
	}

	// 返回第一个匹配的交易对信息
	ticker := result.Data[0]

	c.logger.Info("successfully fetched contract info",
		zap.String("symbol", symbol),
		zap.String("lastPrice", ticker.LastPr),
		zap.Int64("requestTime", result.RequestTime),
	)

	return &ticker, nil
}

// ValidateKlineRequest 验证 K线请求参数
func ValidateKlineRequest(req KlineRequest) error {
	// 验证交易对
	if req.Symbol == "" {
		return ErrInvalidSymbol
	}

	// 验证时间周期
	validGranularities := map[string]bool{
		Granularity1m:  true,
		Granularity5m:  true,
		Granularity15m: true,
		Granularity30m: true,
		Granularity1H:  true,
		Granularity4H:  true,
		Granularity6H:  true,
		Granularity12H: true,
		Granularity1D:  true,
		Granularity1W:  true,
	}

	if !validGranularities[req.Granularity] {
		return ErrInvalidGranularity
	}

	// 验证时间范围
	if req.StartTime != nil && req.EndTime != nil {
		if *req.StartTime >= *req.EndTime {
			return ErrInvalidTimeRange
		}

		// 检查时间范围是否过大（超过30天）
		timeDiff := time.Duration(*req.EndTime-*req.StartTime) * time.Millisecond
		if timeDiff > 30*24*time.Hour {
			return ErrInvalidTimeRange
		}
	}

	// 验证限制条数
	if req.Limit > 0 && req.Limit > MaxKlineLimit {
		req.Limit = MaxKlineLimit
	}

	return nil
}

// GetCurrentTimestamp 获取当前时间戳（毫秒）
func GetCurrentTimestamp() int64 {
	return time.Now().UnixNano() / int64(time.Millisecond)
}

// FormatTimestamp 格式化时间戳为字符串
func FormatTimestamp(timestamp int64) string {
	return strconv.FormatInt(timestamp, 10)
}
