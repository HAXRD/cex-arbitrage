package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// PriceData 价格数据结构
type PriceData struct {
	Symbol      string    `json:"symbol"`
	LastPrice   float64   `json:"last_price"`
	AskPrice    *float64  `json:"ask_price,omitempty"`
	BidPrice    *float64  `json:"bid_price,omitempty"`
	High24h     *float64  `json:"high_24h,omitempty"`
	Low24h      *float64  `json:"low_24h,omitempty"`
	Change24h   *float64  `json:"change_24h,omitempty"`
	BaseVolume  *float64  `json:"base_volume,omitempty"`
	QuoteVolume *float64  `json:"quote_volume,omitempty"`
	Timestamp   time.Time `json:"timestamp"`
}

// MetricsData 指标数据结构
type MetricsData struct {
	Symbol         string    `json:"symbol"`
	PriceChange    float64   `json:"price_change"`     // 价格变化百分比
	Volume24h      float64   `json:"volume_24h"`       // 24小时成交量
	Volatility     float64   `json:"volatility"`       // 波动率
	Turnover       float64   `json:"turnover"`         // 成交额
	UpdateTime     time.Time `json:"update_time"`      // 更新时间
}

// PriceCache 价格缓存接口
type PriceCache interface {
	// SetPrice 设置单个交易对的价格
	SetPrice(ctx context.Context, data *PriceData) error

	// GetPrice 获取单个交易对的价格
	GetPrice(ctx context.Context, symbol string) (*PriceData, error)

	// GetMultiplePrices 批量获取多个交易对的价格
	GetMultiplePrices(ctx context.Context, symbols []string) (map[string]*PriceData, error)

	// DeletePrice 删除单个交易对的价格
	DeletePrice(ctx context.Context, symbol string) error

	// SetMetrics 设置交易对的指标数据
	SetMetrics(ctx context.Context, data *MetricsData) error

	// GetMetrics 获取交易对的指标数据
	GetMetrics(ctx context.Context, symbol string) (*MetricsData, error)

	// SetActiveSymbols 设置活跃交易对列表
	SetActiveSymbols(ctx context.Context, symbols []string) error

	// GetActiveSymbols 获取活跃交易对列表
	GetActiveSymbols(ctx context.Context) ([]string, error)
}

// priceCacheImpl PriceCache 实现
type priceCacheImpl struct {
	client *redis.Client
}

// NewPriceCache 创建 PriceCache 实例
func NewPriceCache(client *Client) PriceCache {
	return &priceCacheImpl{
		client: client.GetClient(),
	}
}

// SetPrice 设置单个交易对的价格
func (c *priceCacheImpl) SetPrice(ctx context.Context, data *PriceData) error {
	if data == nil || data.Symbol == "" {
		return fmt.Errorf("invalid price data")
	}

	// 将价格数据序列化为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal price data: %w", err)
	}

	// 构建缓存键
	key := BuildLatestPriceKey(data.Symbol)

	// 设置缓存，TTL 为 60 秒
	err = c.client.Set(ctx, key, jsonData, TTLRealTimePrice).Err()
	if err != nil {
		return fmt.Errorf("failed to set price cache: %w", err)
	}

	return nil
}

// GetPrice 获取单个交易对的价格
func (c *priceCacheImpl) GetPrice(ctx context.Context, symbol string) (*PriceData, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	key := BuildLatestPriceKey(symbol)

	// 获取缓存数据
	jsonData, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("price not found for symbol: %s", symbol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get price cache: %w", err)
	}

	// 反序列化
	var data PriceData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal price data: %w", err)
	}

	return &data, nil
}

// GetMultiplePrices 批量获取多个交易对的价格
func (c *priceCacheImpl) GetMultiplePrices(ctx context.Context, symbols []string) (map[string]*PriceData, error) {
	if len(symbols) == 0 {
		return make(map[string]*PriceData), nil
	}

	// 使用 Pipeline 批量获取
	pipe := c.client.Pipeline()

	// 构建所有键的获取命令
	keys := make([]string, len(symbols))
	cmds := make([]*redis.StringCmd, len(symbols))
	for i, symbol := range symbols {
		keys[i] = BuildLatestPriceKey(symbol)
		cmds[i] = pipe.Get(ctx, keys[i])
	}

	// 执行 Pipeline
	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("failed to execute pipeline: %w", err)
	}

	// 解析结果
	result := make(map[string]*PriceData)
	for i, cmd := range cmds {
		jsonData, err := cmd.Result()
		if err == redis.Nil {
			// 缓存未命中，跳过
			continue
		}
		if err != nil {
			// 其他错误，跳过这个键
			continue
		}

		var data PriceData
		if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
			// 解析错误，跳过
			continue
		}

		result[symbols[i]] = &data
	}

	return result, nil
}

// DeletePrice 删除单个交易对的价格
func (c *priceCacheImpl) DeletePrice(ctx context.Context, symbol string) error {
	if symbol == "" {
		return fmt.Errorf("symbol cannot be empty")
	}

	key := BuildLatestPriceKey(symbol)
	err := c.client.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete price cache: %w", err)
	}

	return nil
}

// SetMetrics 设置交易对的指标数据
func (c *priceCacheImpl) SetMetrics(ctx context.Context, data *MetricsData) error {
	if data == nil || data.Symbol == "" {
		return fmt.Errorf("invalid metrics data")
	}

	// 将指标数据序列化为 JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal metrics data: %w", err)
	}

	// 构建缓存键
	key := BuildMetricsKey(data.Symbol)

	// 设置缓存，TTL 为 60 秒
	err = c.client.Set(ctx, key, jsonData, TTLRealTimeMetrics).Err()
	if err != nil {
		return fmt.Errorf("failed to set metrics cache: %w", err)
	}

	return nil
}

// GetMetrics 获取交易对的指标数据
func (c *priceCacheImpl) GetMetrics(ctx context.Context, symbol string) (*MetricsData, error) {
	if symbol == "" {
		return nil, fmt.Errorf("symbol cannot be empty")
	}

	key := BuildMetricsKey(symbol)

	// 获取缓存数据
	jsonData, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return nil, fmt.Errorf("metrics not found for symbol: %s", symbol)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get metrics cache: %w", err)
	}

	// 反序列化
	var data MetricsData
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal metrics data: %w", err)
	}

	return &data, nil
}

// SetActiveSymbols 设置活跃交易对列表
func (c *priceCacheImpl) SetActiveSymbols(ctx context.Context, symbols []string) error {
	if len(symbols) == 0 {
		return fmt.Errorf("symbols list cannot be empty")
	}

	key := BuildActiveSymbolsKey()

	// 先删除旧的集合
	if err := c.client.Del(ctx, key).Err(); err != nil {
		return fmt.Errorf("failed to delete old active symbols: %w", err)
	}

	// 使用 SADD 添加所有交易对到集合
	members := make([]interface{}, len(symbols))
	for i, symbol := range symbols {
		members[i] = symbol
	}

	if err := c.client.SAdd(ctx, key, members...).Err(); err != nil {
		return fmt.Errorf("failed to set active symbols: %w", err)
	}

	// 设置过期时间
	if err := c.client.Expire(ctx, key, TTLSymbolList).Err(); err != nil {
		return fmt.Errorf("failed to set TTL for active symbols: %w", err)
	}

	return nil
}

// GetActiveSymbols 获取活跃交易对列表
func (c *priceCacheImpl) GetActiveSymbols(ctx context.Context) ([]string, error) {
	key := BuildActiveSymbolsKey()

	// 使用 SMEMBERS 获取所有成员
	symbols, err := c.client.SMembers(ctx, key).Result()
	if err == redis.Nil {
		return []string{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get active symbols: %w", err)
	}

	return symbols, nil
}

