package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRedisCache_BasicOperations(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("设置和获取字符串", func(t *testing.T) {
		key := "test:string"
		value := "hello world"
		ttl := 5 * time.Minute

		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})

	t.Run("设置和获取JSON数据", func(t *testing.T) {
		key := "test:json"
		value := map[string]interface{}{
			"name": "test",
			"age":  25,
		}
		ttl := 5 * time.Minute

		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.NotEmpty(t, result)
	})

	t.Run("检查键是否存在", func(t *testing.T) {
		key := "test:exists"
		value := "test value"
		ttl := 5 * time.Minute

		// 键不存在
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)

		// 设置键
		err = cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// 键存在
		exists, err = cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("删除键", func(t *testing.T) {
		key := "test:delete"
		value := "test value"
		ttl := 5 * time.Minute

		// 设置键
		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// 确认键存在
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		// 删除键
		err = cache.Delete(ctx, key)
		require.NoError(t, err)

		// 确认键不存在
		exists, err = cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestRedisCache_BatchOperations(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("批量设置", func(t *testing.T) {
		data := []CacheData{
			{
				Key:   "batch:1",
				Value: "value1",
				TTL:   5 * time.Minute,
			},
			{
				Key:   "batch:2",
				Value: "value2",
				TTL:   5 * time.Minute,
			},
			{
				Key:   "batch:3",
				Value: "value3",
				TTL:   5 * time.Minute,
			},
		}

		result, err := cache.SetBatch(ctx, data)
		require.NoError(t, err)
		assert.Equal(t, 3, result.TotalCount)
		assert.Equal(t, 3, result.SuccessCount)
		assert.Equal(t, 0, result.ErrorCount)

		// 验证数据
		for _, item := range data {
			value, err := cache.Get(ctx, item.Key)
			require.NoError(t, err)
			assert.Equal(t, item.Value, value)
		}
	})

	t.Run("批量获取", func(t *testing.T) {
		// 先设置一些数据
		keys := []string{"batch:1", "batch:2", "batch:3"}
		expectedValues := map[string]string{
			"batch:1": "value1",
			"batch:2": "value2",
			"batch:3": "value3",
		}

		values, err := cache.GetBatch(ctx, keys)
		require.NoError(t, err)
		assert.Len(t, values, 3)

		for key, expectedValue := range expectedValues {
			actualValue, exists := values[key]
			assert.True(t, exists)
			assert.Equal(t, expectedValue, actualValue)
		}
	})

	t.Run("批量删除", func(t *testing.T) {
		keys := []string{"batch:1", "batch:2", "batch:3"}

		err := cache.DeleteBatch(ctx, keys)
		require.NoError(t, err)

		// 验证键已删除
		for _, key := range keys {
			exists, err := cache.Exists(ctx, key)
			require.NoError(t, err)
			assert.False(t, exists)
		}
	})
}

func TestRedisCache_PriceDataOperations(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("设置和获取价格数据", func(t *testing.T) {
		price := &PriceData{
			Symbol:    "BTCUSDT",
			Price:     50000.0,
			BidPrice:  49999.0,
			AskPrice:  50001.0,
			Volume:    1000.0,
			Timestamp: time.Now(),
			Source:    "test",
			Latency:   10 * time.Millisecond,
		}

		err := cache.SetPrice(ctx, "BTCUSDT", price)
		require.NoError(t, err)

		result, err := cache.GetPrice(ctx, "BTCUSDT")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, price.Symbol, result.Symbol)
		assert.Equal(t, price.Price, result.Price)
		assert.Equal(t, price.BidPrice, result.BidPrice)
		assert.Equal(t, price.AskPrice, result.AskPrice)
		assert.Equal(t, price.Volume, result.Volume)
		assert.Equal(t, price.Source, result.Source)
	})

	t.Run("批量设置价格数据", func(t *testing.T) {
		prices := []*PriceData{
			{
				Symbol:    "BTCUSDT",
				Price:     50000.0,
				Timestamp: time.Now(),
				Source:    "test",
			},
			{
				Symbol:    "ETHUSDT",
				Price:     3000.0,
				Timestamp: time.Now(),
				Source:    "test",
			},
		}

		err := cache.SetPrices(ctx, prices)
		require.NoError(t, err)

		// 验证数据
		for _, price := range prices {
			result, err := cache.GetPrice(ctx, price.Symbol)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, price.Symbol, result.Symbol)
			assert.Equal(t, price.Price, result.Price)
		}
	})
}

func TestRedisCache_ChangeRateOperations(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("设置和获取变化率数据", func(t *testing.T) {
		rate := &ProcessedPriceChangeRate{
			Symbol:     "BTCUSDT",
			TimeWindow: "1m",
			ChangeRate: 2.5,
			StartPrice: 50000.0,
			EndPrice:   51250.0,
			Timestamp:  time.Now(),
			IsValid:    true,
			IsAnomaly:  false,
		}

		err := cache.SetChangeRate(ctx, "BTCUSDT", TimeWindow1m, rate)
		require.NoError(t, err)

		result, err := cache.GetChangeRate(ctx, "BTCUSDT", TimeWindow1m)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, rate.Symbol, result.Symbol)
		assert.Equal(t, rate.TimeWindow, result.TimeWindow)
		assert.Equal(t, rate.ChangeRate, result.ChangeRate)
		assert.Equal(t, rate.StartPrice, result.StartPrice)
		assert.Equal(t, rate.EndPrice, result.EndPrice)
		assert.Equal(t, rate.IsValid, result.IsValid)
		assert.Equal(t, rate.IsAnomaly, result.IsAnomaly)
	})

	t.Run("批量设置变化率数据", func(t *testing.T) {
		rates := map[TimeWindow]*ProcessedPriceChangeRate{
			TimeWindow1m: {
				Symbol:     "BTCUSDT",
				TimeWindow: "1m",
				ChangeRate: 1.0,
				IsValid:    true,
			},
			TimeWindow5m: {
				Symbol:     "BTCUSDT",
				TimeWindow: "5m",
				ChangeRate: 2.5,
				IsValid:    true,
			},
			TimeWindow15m: {
				Symbol:     "BTCUSDT",
				TimeWindow: "15m",
				ChangeRate: 5.0,
				IsValid:    true,
			},
		}

		err := cache.SetChangeRates(ctx, "BTCUSDT", rates)
		require.NoError(t, err)

		// 验证数据
		for window, expectedRate := range rates {
			result, err := cache.GetChangeRate(ctx, "BTCUSDT", window)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, expectedRate.Symbol, result.Symbol)
			assert.Equal(t, expectedRate.TimeWindow, result.TimeWindow)
			assert.Equal(t, expectedRate.ChangeRate, result.ChangeRate)
			assert.Equal(t, expectedRate.IsValid, result.IsValid)
		}
	})
}

func TestRedisCache_SymbolOperations(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("设置和获取交易对信息", func(t *testing.T) {
		symbol := &SymbolInfo{
			Symbol:     "BTCUSDT",
			Name:       "Bitcoin/USDT",
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		err := cache.SetSymbol(ctx, "BTCUSDT", symbol)
		require.NoError(t, err)

		result, err := cache.GetSymbol(ctx, "BTCUSDT")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, symbol.Symbol, result.Symbol)
		assert.Equal(t, symbol.Name, result.Name)
		assert.Equal(t, symbol.BaseAsset, result.BaseAsset)
		assert.Equal(t, symbol.QuoteAsset, result.QuoteAsset)
		assert.Equal(t, symbol.Status, result.Status)
	})

	t.Run("批量设置交易对信息", func(t *testing.T) {
		symbols := []*SymbolInfo{
			{
				Symbol:     "BTCUSDT",
				Name:       "Bitcoin/USDT",
				BaseAsset:  "BTC",
				QuoteAsset: "USDT",
				Status:     "active",
			},
			{
				Symbol:     "ETHUSDT",
				Name:       "Ethereum/USDT",
				BaseAsset:  "ETH",
				QuoteAsset: "USDT",
				Status:     "active",
			},
		}

		err := cache.SetSymbols(ctx, symbols)
		require.NoError(t, err)

		// 验证数据
		for _, symbol := range symbols {
			result, err := cache.GetSymbol(ctx, symbol.Symbol)
			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, symbol.Symbol, result.Symbol)
			assert.Equal(t, symbol.Name, result.Name)
			assert.Equal(t, symbol.BaseAsset, result.BaseAsset)
			assert.Equal(t, symbol.QuoteAsset, result.QuoteAsset)
			assert.Equal(t, symbol.Status, result.Status)
		}
	})
}

func TestRedisCache_KeyManagement(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("获取键列表", func(t *testing.T) {
		// 设置一些测试键
		testKeys := []string{
			"price:BTCUSDT",
			"price:ETHUSDT",
			"changerate:BTCUSDT:1m",
			"changerate:BTCUSDT:5m",
			"symbol:BTCUSDT",
		}

		for _, key := range testKeys {
			err := cache.Set(ctx, key, "test_value", 5*time.Minute)
			require.NoError(t, err)
		}

		// 获取所有键
		keys, err := cache.GetKeys(ctx, "*")
		require.NoError(t, err)
		assert.Len(t, keys, 5)

		// 获取价格键
		priceKeys, err := cache.GetKeys(ctx, "price:*")
		require.NoError(t, err)
		assert.Len(t, priceKeys, 2)

		// 获取变化率键
		changeRateKeys, err := cache.GetKeys(ctx, "changerate:*")
		require.NoError(t, err)
		assert.Len(t, changeRateKeys, 2)
	})

	t.Run("按类型获取键", func(t *testing.T) {
		// 清理之前的数据
		cache.DeleteKeys(ctx, "*")

		// 设置不同类型的键
		cache.Set(ctx, "price:BTCUSDT", "50000", 5*time.Minute)
		cache.Set(ctx, "changerate:BTCUSDT:1m", "2.5", 5*time.Minute)
		cache.Set(ctx, "symbol:BTCUSDT", "{}", 5*time.Minute)

		// 获取价格键
		priceKeys, err := cache.GetKeysByType(ctx, PriceKey)
		require.NoError(t, err)
		assert.Len(t, priceKeys, 1)
		assert.Contains(t, priceKeys, "price:BTCUSDT")

		// 获取变化率键
		changeRateKeys, err := cache.GetKeysByType(ctx, ChangeRateKey)
		require.NoError(t, err)
		assert.Len(t, changeRateKeys, 1)
		assert.Contains(t, changeRateKeys, "changerate:BTCUSDT:1m")
	})

	t.Run("删除键模式", func(t *testing.T) {
		// 设置一些测试键
		cache.Set(ctx, "price:BTCUSDT", "50000", 5*time.Minute)
		cache.Set(ctx, "price:ETHUSDT", "3000", 5*time.Minute)
		cache.Set(ctx, "symbol:BTCUSDT", "{}", 5*time.Minute)

		// 删除价格键
		err := cache.DeleteKeys(ctx, "price:*")
		require.NoError(t, err)

		// 验证价格键已删除
		exists, err := cache.Exists(ctx, "price:BTCUSDT")
		require.NoError(t, err)
		assert.False(t, exists)

		exists, err = cache.Exists(ctx, "price:ETHUSDT")
		require.NoError(t, err)
		assert.False(t, exists)

		// 验证其他键仍然存在
		exists, err = cache.Exists(ctx, "symbol:BTCUSDT")
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestRedisCache_TTLManagement(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("设置TTL", func(t *testing.T) {
		key := "test:ttl"
		value := "test value"
		ttl := 5 * time.Minute

		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// 获取TTL
		resultTTL, err := cache.GetTTL(ctx, key)
		require.NoError(t, err)
		assert.True(t, resultTTL > 0)
		assert.True(t, resultTTL <= ttl)
	})

	t.Run("更新TTL", func(t *testing.T) {
		key := "test:ttl_update"
		value := "test value"
		initialTTL := 1 * time.Minute
		newTTL := 10 * time.Minute

		// 设置初始TTL
		err := cache.Set(ctx, key, value, initialTTL)
		require.NoError(t, err)

		// 更新TTL
		err = cache.SetTTL(ctx, key, newTTL)
		require.NoError(t, err)

		// 验证TTL已更新
		resultTTL, err := cache.GetTTL(ctx, key)
		require.NoError(t, err)
		assert.True(t, resultTTL > initialTTL)
		assert.True(t, resultTTL <= newTTL)
	})

	t.Run("批量设置TTL", func(t *testing.T) {
		// 设置一些测试键
		keys := []string{"test:ttl1", "test:ttl2", "test:ttl3"}
		for _, key := range keys {
			cache.Set(ctx, key, "test value", 1*time.Minute)
		}

		// 批量设置TTL
		err := cache.ExpireKeys(ctx, "test:ttl*", 10*time.Minute)
		require.NoError(t, err)

		// 验证TTL已更新
		for _, key := range keys {
			ttl, err := cache.GetTTL(ctx, key)
			require.NoError(t, err)
			assert.True(t, ttl > 5*time.Minute)
		}
	})
}

func TestRedisCache_Stats(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("获取统计信息", func(t *testing.T) {
		// 执行一些操作
		cache.Set(ctx, "test:1", "value1", 5*time.Minute)
		cache.Set(ctx, "test:2", "value2", 5*time.Minute)
		cache.Get(ctx, "test:1")
		cache.Delete(ctx, "test:2")

		stats := cache.GetStats()
		require.NotNil(t, stats)
		assert.True(t, stats.WriteCount >= 2)
		assert.True(t, stats.ReadCount >= 1)
		assert.True(t, stats.DeleteCount >= 1)
	})

	t.Run("重置统计信息", func(t *testing.T) {
		// 执行一些操作
		cache.Set(ctx, "test:reset", "value", 5*time.Minute)

		// 重置统计
		cache.ResetStats()

		stats := cache.GetStats()
		require.NotNil(t, stats)
		assert.Equal(t, int64(0), stats.WriteCount)
		assert.Equal(t, int64(0), stats.ReadCount)
		assert.Equal(t, int64(0), stats.DeleteCount)
	})
}

func TestRedisCache_HealthCheck(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("健康检查", func(t *testing.T) {
		err := cache.HealthCheck(ctx)
		require.NoError(t, err)
	})
}
