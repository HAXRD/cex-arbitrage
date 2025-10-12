package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis 创建测试用的 Redis 客户端（使用 miniredis）
func setupTestRedis(t *testing.T) (*Client, *miniredis.Miniredis) {
	// 创建 miniredis 服务器
	mr, err := miniredis.Run()
	require.NoError(t, err)

	// 创建 Redis 客户端
	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	client := &Client{
		client: rdb,
		logger: nil,
	}

	return client, mr
}

// createTestPriceData 创建测试用的价格数据
func createTestPriceData(symbol string) *PriceData {
	askPrice := 50001.0
	bidPrice := 49999.0
	high24h := 51000.0
	low24h := 49000.0
	change24h := 2.5
	baseVolume := 1000.5
	quoteVolume := 50000000.0

	return &PriceData{
		Symbol:      symbol,
		LastPrice:   50000.0,
		AskPrice:    &askPrice,
		BidPrice:    &bidPrice,
		High24h:     &high24h,
		Low24h:      &low24h,
		Change24h:   &change24h,
		BaseVolume:  &baseVolume,
		QuoteVolume: &quoteVolume,
		Timestamp:   time.Now().UTC(),
	}
}

// createTestMetricsData 创建测试用的指标数据
func createTestMetricsData(symbol string) *MetricsData {
	return &MetricsData{
		Symbol:      symbol,
		PriceChange: 2.5,
		Volume24h:   1000000.0,
		Volatility:  0.15,
		Turnover:    50000000.0,
		UpdateTime:  time.Now().UTC(),
	}
}

func TestPriceCache_SetPrice(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("成功设置价格", func(t *testing.T) {
		data := createTestPriceData("BTCUSDT")
		err := cache.SetPrice(ctx, data)
		require.NoError(t, err)

		// 验证缓存已设置
		key := BuildLatestPriceKey("BTCUSDT")
		exists, err := client.GetClient().Exists(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)

		// 验证 TTL
		ttl, err := client.GetClient().TTL(ctx, key).Result()
		require.NoError(t, err)
		assert.Greater(t, ttl.Seconds(), 0.0)
		assert.LessOrEqual(t, ttl.Seconds(), TTLRealTimePrice.Seconds())
	})

	t.Run("设置 nil 数据应返回错误", func(t *testing.T) {
		err := cache.SetPrice(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("设置空 symbol 应返回错误", func(t *testing.T) {
		data := createTestPriceData("")
		err := cache.SetPrice(ctx, data)
		assert.Error(t, err)
	})
}

func TestPriceCache_GetPrice(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("成功获取价格", func(t *testing.T) {
		// 先设置
		originalData := createTestPriceData("BTCUSDT")
		err := cache.SetPrice(ctx, originalData)
		require.NoError(t, err)

		// 再获取
		data, err := cache.GetPrice(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "BTCUSDT", data.Symbol)
		assert.Equal(t, 50000.0, data.LastPrice)
		assert.NotNil(t, data.AskPrice)
		assert.Equal(t, 50001.0, *data.AskPrice)
	})

	t.Run("获取不存在的价格应返回错误", func(t *testing.T) {
		_, err := cache.GetPrice(ctx, "NONEXISTENT")
		assert.Error(t, err)
	})

	t.Run("获取空 symbol 应返回错误", func(t *testing.T) {
		_, err := cache.GetPrice(ctx, "")
		assert.Error(t, err)
	})
}

func TestPriceCache_GetMultiplePrices(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	// 准备测试数据
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
	for _, symbol := range symbols {
		data := createTestPriceData(symbol)
		err := cache.SetPrice(ctx, data)
		require.NoError(t, err)
	}

	t.Run("成功批量获取价格", func(t *testing.T) {
		result, err := cache.GetMultiplePrices(ctx, symbols)
		require.NoError(t, err)
		assert.Len(t, result, 3)

		for _, symbol := range symbols {
			data, exists := result[symbol]
			assert.True(t, exists, "symbol %s should exist", symbol)
			assert.NotNil(t, data)
			assert.Equal(t, symbol, data.Symbol)
		}
	})

	t.Run("部分交易对不存在时只返回存在的", func(t *testing.T) {
		mixedSymbols := []string{"BTCUSDT", "NONEXISTENT", "ETHUSDT"}
		result, err := cache.GetMultiplePrices(ctx, mixedSymbols)
		require.NoError(t, err)
		assert.Len(t, result, 2)

		assert.Contains(t, result, "BTCUSDT")
		assert.Contains(t, result, "ETHUSDT")
		assert.NotContains(t, result, "NONEXISTENT")
	})

	t.Run("空列表应返回空 map", func(t *testing.T) {
		result, err := cache.GetMultiplePrices(ctx, []string{})
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

func TestPriceCache_DeletePrice(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("成功删除价格", func(t *testing.T) {
		// 先设置
		data := createTestPriceData("BTCUSDT")
		err := cache.SetPrice(ctx, data)
		require.NoError(t, err)

		// 再删除
		err = cache.DeletePrice(ctx, "BTCUSDT")
		require.NoError(t, err)

		// 验证已删除
		key := BuildLatestPriceKey("BTCUSDT")
		exists, err := client.GetClient().Exists(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(0), exists)
	})

	t.Run("删除不存在的价格不应返回错误", func(t *testing.T) {
		err := cache.DeletePrice(ctx, "NONEXISTENT")
		assert.NoError(t, err)
	})

	t.Run("删除空 symbol 应返回错误", func(t *testing.T) {
		err := cache.DeletePrice(ctx, "")
		assert.Error(t, err)
	})
}

func TestPriceCache_SetMetrics(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("成功设置指标", func(t *testing.T) {
		data := createTestMetricsData("BTCUSDT")
		err := cache.SetMetrics(ctx, data)
		require.NoError(t, err)

		// 验证缓存已设置
		key := BuildMetricsKey("BTCUSDT")
		exists, err := client.GetClient().Exists(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(1), exists)
	})

	t.Run("设置 nil 数据应返回错误", func(t *testing.T) {
		err := cache.SetMetrics(ctx, nil)
		assert.Error(t, err)
	})

	t.Run("设置空 symbol 应返回错误", func(t *testing.T) {
		data := createTestMetricsData("")
		err := cache.SetMetrics(ctx, data)
		assert.Error(t, err)
	})
}

func TestPriceCache_GetMetrics(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("成功获取指标", func(t *testing.T) {
		// 先设置
		originalData := createTestMetricsData("BTCUSDT")
		err := cache.SetMetrics(ctx, originalData)
		require.NoError(t, err)

		// 再获取
		data, err := cache.GetMetrics(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.NotNil(t, data)
		assert.Equal(t, "BTCUSDT", data.Symbol)
		assert.Equal(t, 2.5, data.PriceChange)
		assert.Equal(t, 1000000.0, data.Volume24h)
	})

	t.Run("获取不存在的指标应返回错误", func(t *testing.T) {
		_, err := cache.GetMetrics(ctx, "NONEXISTENT")
		assert.Error(t, err)
	})

	t.Run("获取空 symbol 应返回错误", func(t *testing.T) {
		_, err := cache.GetMetrics(ctx, "")
		assert.Error(t, err)
	})
}

func TestPriceCache_SetActiveSymbols(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("成功设置活跃交易对列表", func(t *testing.T) {
		symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
		err := cache.SetActiveSymbols(ctx, symbols)
		require.NoError(t, err)

		// 验证缓存已设置
		key := BuildActiveSymbolsKey()
		count, err := client.GetClient().SCard(ctx, key).Result()
		require.NoError(t, err)
		assert.Equal(t, int64(3), count)

		// 验证 TTL
		ttl, err := client.GetClient().TTL(ctx, key).Result()
		require.NoError(t, err)
		assert.Greater(t, ttl.Seconds(), 0.0)
	})

	t.Run("更新活跃交易对列表应替换旧数据", func(t *testing.T) {
		// 第一次设置
		symbols1 := []string{"BTCUSDT", "ETHUSDT"}
		err := cache.SetActiveSymbols(ctx, symbols1)
		require.NoError(t, err)

		// 第二次设置（不同的列表）
		symbols2 := []string{"BNBUSDT", "ADAUSDT", "DOGEUSDT"}
		err = cache.SetActiveSymbols(ctx, symbols2)
		require.NoError(t, err)

		// 验证只有新列表的数据
		result, err := cache.GetActiveSymbols(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 3)
		assert.Contains(t, result, "BNBUSDT")
		assert.Contains(t, result, "ADAUSDT")
		assert.Contains(t, result, "DOGEUSDT")
		assert.NotContains(t, result, "BTCUSDT")
	})

	t.Run("设置空列表应返回错误", func(t *testing.T) {
		err := cache.SetActiveSymbols(ctx, []string{})
		assert.Error(t, err)
	})
}

func TestPriceCache_GetActiveSymbols(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("成功获取活跃交易对列表", func(t *testing.T) {
		// 先设置
		symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT"}
		err := cache.SetActiveSymbols(ctx, symbols)
		require.NoError(t, err)

		// 再获取
		result, err := cache.GetActiveSymbols(ctx)
		require.NoError(t, err)
		assert.Len(t, result, 3)

		// 验证所有交易对都在列表中
		for _, symbol := range symbols {
			assert.Contains(t, result, symbol)
		}
	})

	t.Run("获取不存在的列表应返回空列表", func(t *testing.T) {
		// 确保缓存为空
		key := BuildActiveSymbolsKey()
		client.GetClient().Del(ctx, key)

		result, err := cache.GetActiveSymbols(ctx)
		require.NoError(t, err)
		assert.Empty(t, result)
	})
}

// Benchmark tests
func BenchmarkPriceCache_SetPrice(b *testing.B) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	client := &Client{client: rdb}
	cache := NewPriceCache(client)
	ctx := context.Background()
	data := createTestPriceData("BTCUSDT")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.SetPrice(ctx, data)
	}
}

func BenchmarkPriceCache_GetPrice(b *testing.B) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	client := &Client{client: rdb}
	cache := NewPriceCache(client)
	ctx := context.Background()

	// 预先设置数据
	data := createTestPriceData("BTCUSDT")
	cache.SetPrice(ctx, data)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetPrice(ctx, "BTCUSDT")
	}
}

func BenchmarkPriceCache_GetMultiplePrices(b *testing.B) {
	mr, _ := miniredis.Run()
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	client := &Client{client: rdb}
	cache := NewPriceCache(client)
	ctx := context.Background()

	// 预先设置数据
	symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "DOGEUSDT"}
	for _, symbol := range symbols {
		data := createTestPriceData(symbol)
		cache.SetPrice(ctx, data)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		cache.GetMultiplePrices(ctx, symbols)
	}
}

