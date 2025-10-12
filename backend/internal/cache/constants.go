package cache

import (
	"fmt"
	"time"
)

// 缓存键命名规范：namespace:type:identifier
// 例如：cryptosignal:price:BTCUSDT

const (
	// Namespace 缓存命名空间前缀
	Namespace = "cryptosignal"

	// TTL 定义
	TTLRealTimePrice    = 60 * time.Second  // 实时价格：60秒
	TTLRealTimeMetrics  = 60 * time.Second  // 实时指标：60秒
	TTLSymbolList       = 300 * time.Second // 交易对列表：5分钟
	TTLWebSocketSession = 90 * time.Second  // WebSocket 会话：90秒
	TTLKlineData        = 300 * time.Second // K线数据：5分钟
)

// CacheKeyType 缓存键类型
type CacheKeyType string

const (
	// 价格相关
	KeyTypePrice         CacheKeyType = "price"         // 实时价格
	KeyTypePriceTick     CacheKeyType = "price_tick"    // 价格 Tick 数据
	KeyTypeLatestPrice   CacheKeyType = "latest_price"  // 最新价格
	KeyTypePriceMultiple CacheKeyType = "price_multi"   // 批量价格

	// 指标相关
	KeyTypeMetrics       CacheKeyType = "metrics"        // 实时指标
	KeyTypeVolume        CacheKeyType = "volume"         // 交易量
	KeyTypePriceChange   CacheKeyType = "price_change"   // 价格变化
	KeyTypeVolatility    CacheKeyType = "volatility"     // 波动率

	// 交易对相关
	KeyTypeSymbol        CacheKeyType = "symbol"         // 交易对信息
	KeyTypeSymbolList    CacheKeyType = "symbol_list"    // 交易对列表
	KeyTypeActiveSymbols CacheKeyType = "active_symbols" // 活跃交易对

	// K线数据
	KeyTypeKline         CacheKeyType = "kline"          // K线数据
	KeyTypeKlineLatest   CacheKeyType = "kline_latest"   // 最新K线

	// WebSocket 相关
	KeyTypeWSSession     CacheKeyType = "ws_session"     // WebSocket 会话
	KeyTypeWSHeartbeat   CacheKeyType = "ws_heartbeat"   // WebSocket 心跳

	// 系统相关
	KeyTypeHealth        CacheKeyType = "health"         // 健康检查
	KeyTypeLock          CacheKeyType = "lock"           // 分布式锁
)

// CacheKeyBuilder 缓存键构建器
type CacheKeyBuilder struct {
	namespace string
	keyType   CacheKeyType
	parts     []string
}

// NewCacheKeyBuilder 创建缓存键构建器
func NewCacheKeyBuilder(keyType CacheKeyType) *CacheKeyBuilder {
	return &CacheKeyBuilder{
		namespace: Namespace,
		keyType:   keyType,
		parts:     make([]string, 0),
	}
}

// WithPart 添加键部分
func (b *CacheKeyBuilder) WithPart(part string) *CacheKeyBuilder {
	b.parts = append(b.parts, part)
	return b
}

// Build 构建缓存键
func (b *CacheKeyBuilder) Build() string {
	if len(b.parts) == 0 {
		return fmt.Sprintf("%s:%s", b.namespace, b.keyType)
	}

	key := fmt.Sprintf("%s:%s", b.namespace, b.keyType)
	for _, part := range b.parts {
		key = fmt.Sprintf("%s:%s", key, part)
	}
	return key
}

// 便捷方法：构建常用的缓存键

// BuildLatestPriceKey 构建最新价格缓存键
// 格式：cryptosignal:latest_price:BTCUSDT
func BuildLatestPriceKey(symbol string) string {
	return NewCacheKeyBuilder(KeyTypeLatestPrice).
		WithPart(symbol).
		Build()
}

// BuildPriceTickKey 构建价格 Tick 缓存键
// 格式：cryptosignal:price_tick:BTCUSDT
func BuildPriceTickKey(symbol string) string {
	return NewCacheKeyBuilder(KeyTypePriceTick).
		WithPart(symbol).
		Build()
}

// BuildPriceMultipleKey 构建批量价格缓存键
// 格式：cryptosignal:price_multi
func BuildPriceMultipleKey() string {
	return NewCacheKeyBuilder(KeyTypePriceMultiple).Build()
}

// BuildMetricsKey 构建指标缓存键
// 格式：cryptosignal:metrics:BTCUSDT
func BuildMetricsKey(symbol string) string {
	return NewCacheKeyBuilder(KeyTypeMetrics).
		WithPart(symbol).
		Build()
}

// BuildActiveSymbolsKey 构建活跃交易对列表缓存键
// 格式：cryptosignal:active_symbols
func BuildActiveSymbolsKey() string {
	return NewCacheKeyBuilder(KeyTypeActiveSymbols).Build()
}

// BuildSymbolKey 构建交易对信息缓存键
// 格式：cryptosignal:symbol:BTCUSDT
func BuildSymbolKey(symbol string) string {
	return NewCacheKeyBuilder(KeyTypeSymbol).
		WithPart(symbol).
		Build()
}

// BuildKlineLatestKey 构建最新K线缓存键
// 格式：cryptosignal:kline_latest:BTCUSDT:1m
func BuildKlineLatestKey(symbol, granularity string) string {
	return NewCacheKeyBuilder(KeyTypeKlineLatest).
		WithPart(symbol).
		WithPart(granularity).
		Build()
}

// BuildWSSessionKey 构建 WebSocket 会话缓存键
// 格式：cryptosignal:ws_session:session_id
func BuildWSSessionKey(sessionID string) string {
	return NewCacheKeyBuilder(KeyTypeWSSession).
		WithPart(sessionID).
		Build()
}

// BuildWSHeartbeatKey 构建 WebSocket 心跳缓存键
// 格式：cryptosignal:ws_heartbeat:session_id
func BuildWSHeartbeatKey(sessionID string) string {
	return NewCacheKeyBuilder(KeyTypeWSHeartbeat).
		WithPart(sessionID).
		Build()
}

// BuildLockKey 构建分布式锁缓存键
// 格式：cryptosignal:lock:resource_name
func BuildLockKey(resourceName string) string {
	return NewCacheKeyBuilder(KeyTypeLock).
		WithPart(resourceName).
		Build()
}

