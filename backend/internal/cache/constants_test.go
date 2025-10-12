package cache

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCacheKeyBuilder(t *testing.T) {
	t.Run("构建简单键", func(t *testing.T) {
		key := NewCacheKeyBuilder(KeyTypePrice).Build()
		assert.Equal(t, "cryptosignal:price", key)
	})

	t.Run("构建带一个部分的键", func(t *testing.T) {
		key := NewCacheKeyBuilder(KeyTypePrice).
			WithPart("BTCUSDT").
			Build()
		assert.Equal(t, "cryptosignal:price:BTCUSDT", key)
	})

	t.Run("构建带多个部分的键", func(t *testing.T) {
		key := NewCacheKeyBuilder(KeyTypeKline).
			WithPart("BTCUSDT").
			WithPart("1m").
			WithPart("latest").
			Build()
		assert.Equal(t, "cryptosignal:kline:BTCUSDT:1m:latest", key)
	})

	t.Run("链式调用应该正常工作", func(t *testing.T) {
		builder := NewCacheKeyBuilder(KeyTypeMetrics)
		key := builder.
			WithPart("part1").
			WithPart("part2").
			WithPart("part3").
			Build()
		assert.Equal(t, "cryptosignal:metrics:part1:part2:part3", key)
	})
}

func TestBuildLatestPriceKey(t *testing.T) {
	tests := []struct {
		name     string
		symbol   string
		expected string
	}{
		{
			name:     "BTCUSDT",
			symbol:   "BTCUSDT",
			expected: "cryptosignal:latest_price:BTCUSDT",
		},
		{
			name:     "ETHUSDT",
			symbol:   "ETHUSDT",
			expected: "cryptosignal:latest_price:ETHUSDT",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			key := BuildLatestPriceKey(tt.symbol)
			assert.Equal(t, tt.expected, key)
		})
	}
}

func TestBuildPriceTickKey(t *testing.T) {
	key := BuildPriceTickKey("BTCUSDT")
	assert.Equal(t, "cryptosignal:price_tick:BTCUSDT", key)
}

func TestBuildPriceMultipleKey(t *testing.T) {
	key := BuildPriceMultipleKey()
	assert.Equal(t, "cryptosignal:price_multi", key)
}

func TestBuildMetricsKey(t *testing.T) {
	key := BuildMetricsKey("ETHUSDT")
	assert.Equal(t, "cryptosignal:metrics:ETHUSDT", key)
}

func TestBuildActiveSymbolsKey(t *testing.T) {
	key := BuildActiveSymbolsKey()
	assert.Equal(t, "cryptosignal:active_symbols", key)
}

func TestBuildSymbolKey(t *testing.T) {
	key := BuildSymbolKey("BNBUSDT")
	assert.Equal(t, "cryptosignal:symbol:BNBUSDT", key)
}

func TestBuildKlineLatestKey(t *testing.T) {
	key := BuildKlineLatestKey("BTCUSDT", "1m")
	assert.Equal(t, "cryptosignal:kline_latest:BTCUSDT:1m", key)

	key = BuildKlineLatestKey("ETHUSDT", "5m")
	assert.Equal(t, "cryptosignal:kline_latest:ETHUSDT:5m", key)
}

func TestBuildWSSessionKey(t *testing.T) {
	key := BuildWSSessionKey("session-123")
	assert.Equal(t, "cryptosignal:ws_session:session-123", key)
}

func TestBuildWSHeartbeatKey(t *testing.T) {
	key := BuildWSHeartbeatKey("session-456")
	assert.Equal(t, "cryptosignal:ws_heartbeat:session-456", key)
}

func TestBuildLockKey(t *testing.T) {
	key := BuildLockKey("resource-name")
	assert.Equal(t, "cryptosignal:lock:resource-name", key)
}

func TestKeyTypeConstants(t *testing.T) {
	// 验证所有键类型常量已定义
	assert.NotEmpty(t, KeyTypePrice)
	assert.NotEmpty(t, KeyTypePriceTick)
	assert.NotEmpty(t, KeyTypeLatestPrice)
	assert.NotEmpty(t, KeyTypePriceMultiple)
	assert.NotEmpty(t, KeyTypeMetrics)
	assert.NotEmpty(t, KeyTypeVolume)
	assert.NotEmpty(t, KeyTypePriceChange)
	assert.NotEmpty(t, KeyTypeVolatility)
	assert.NotEmpty(t, KeyTypeSymbol)
	assert.NotEmpty(t, KeyTypeSymbolList)
	assert.NotEmpty(t, KeyTypeActiveSymbols)
	assert.NotEmpty(t, KeyTypeKline)
	assert.NotEmpty(t, KeyTypeKlineLatest)
	assert.NotEmpty(t, KeyTypeWSSession)
	assert.NotEmpty(t, KeyTypeWSHeartbeat)
	assert.NotEmpty(t, KeyTypeHealth)
	assert.NotEmpty(t, KeyTypeLock)
}

func TestTTLConstants(t *testing.T) {
	// 验证 TTL 常量值合理
	assert.Equal(t, TTLRealTimePrice, 60*time.Second)
	assert.Equal(t, TTLRealTimeMetrics, 60*time.Second)
	assert.Equal(t, TTLSymbolList, 300*time.Second)
	assert.Equal(t, TTLWebSocketSession, 90*time.Second)
	assert.Equal(t, TTLKlineData, 300*time.Second)
}

func TestNamespaceConstant(t *testing.T) {
	assert.Equal(t, "cryptosignal", Namespace)
}

// Benchmark tests
func BenchmarkCacheKeyBuilder(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewCacheKeyBuilder(KeyTypePrice).
			WithPart("BTCUSDT").
			WithPart("latest").
			Build()
	}
}

func BenchmarkBuildLatestPriceKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildLatestPriceKey("BTCUSDT")
	}
}

func BenchmarkBuildKlineLatestKey(b *testing.B) {
	for i := 0; i < b.N; i++ {
		BuildKlineLatestKey("BTCUSDT", "1m")
	}
}

