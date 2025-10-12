package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// setupTestCache 设置测试缓存
func setupTestCache(t *testing.T) (PriceCache, *Client) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})

	logger := zaptest.NewLogger(t)

	// 创建缓存客户端
	cfg := DefaultConfig()
	cfg.Host = "localhost"
	cfg.Port = 6379
	cacheClient, err := NewClient(cfg, logger)
	require.NoError(t, err)

	// 创建缓存
	cache := NewPriceCache(cacheClient)

	return cache, cacheClient
}

func TestWriteThroughStrategy_Consistency(t *testing.T) {
	cache, cacheClient := setupTestCache(t)
	defer cacheClient.Close()

	logger := zaptest.NewLogger(t)

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建写穿透策略
	strategy := NewWriteThroughStrategy(manager)

	// 测试数据
	priceData := &PriceData{
		Symbol:    "BTCUSDT",
		LastPrice: 50000.0,
		Timestamp: time.Now(),
	}

	// 写入数据
	err := strategy.WritePrice(context.Background(), priceData)
	require.NoError(t, err)

	// 验证缓存中的数据
	cachedData, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", cachedData.Symbol)
	assert.Equal(t, 50000.0, cachedData.LastPrice)
}

func TestWriteBehindStrategy_Consistency(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建写回策略
	strategy := NewWriteBehindStrategy(manager)

	// 测试数据
	priceData := &PriceData{
		Symbol:    "BTCUSDT",
		LastPrice: 50000.0,
		Timestamp: time.Now(),
	}

	// 写入数据
	err := strategy.WritePrice(context.Background(), priceData)
	require.NoError(t, err)

	// 验证缓存中的数据
	cachedData, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", cachedData.Symbol)
	assert.Equal(t, 50000.0, cachedData.LastPrice)

	// 等待异步数据库写入完成
	time.Sleep(100 * time.Millisecond)
}

func TestCacheAsideStrategy_Consistency(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建缓存旁路策略
	strategy := NewCacheAsideStrategy(manager)

	// 读取数据（应该从数据库读取并更新缓存）
	result, err := strategy.ReadPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", result.Symbol)
	assert.Equal(t, 50000.0, result.LastPrice)

	// 等待异步缓存更新完成
	time.Sleep(100 * time.Millisecond)

	// 再次读取应该从缓存获取
	cachedData, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", cachedData.Symbol)
	assert.Equal(t, 50000.0, cachedData.LastPrice)
}

func TestCacheInvalidationStrategy_Consistency(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建缓存失效策略
	strategy := NewCacheInvalidationStrategy(manager)

	// 先设置价格数据
	priceData := &PriceData{
		Symbol:    "BTCUSDT",
		LastPrice: 50000.0,
		Timestamp: time.Now(),
	}

	err := cache.SetPrice(context.Background(), priceData)
	require.NoError(t, err)

	// 验证数据存在
	cachedData, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", cachedData.Symbol)

	// 失效缓存
	err = strategy.InvalidatePrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)

	// 验证数据已被删除
	_, err = cache.GetPrice(context.Background(), "BTCUSDT")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCacheConsistency_ConcurrentReadWrite(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建写穿透策略
	strategy := NewWriteThroughStrategy(manager)

	// 并发写入测试
	done := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func(index int) {
			priceData := &PriceData{
				Symbol:    "BTCUSDT",
				LastPrice: 50000.0 + float64(index),
				Timestamp: time.Now(),
			}

			err := strategy.WritePrice(context.Background(), priceData)
			done <- err
		}(i)
	}

	// 等待所有写入完成
	for i := 0; i < 10; i++ {
		err := <-done
		require.NoError(t, err)
	}

	// 验证最终数据
	cachedData, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", cachedData.Symbol)
	// 最终价格应该是最后一个写入的值
	assert.True(t, cachedData.LastPrice >= 50000.0)
}

func TestCacheConsistency_DataIntegrity(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建写穿透策略
	strategy := NewWriteThroughStrategy(manager)

	// 测试数据完整性
	priceData := &PriceData{
		Symbol:      "BTCUSDT",
		LastPrice:   50000.0,
		AskPrice:    &[]float64{50001.0}[0],
		BidPrice:    &[]float64{49999.0}[0],
		High24h:     &[]float64{51000.0}[0],
		Low24h:      &[]float64{49000.0}[0],
		Change24h:   &[]float64{1000.0}[0],
		BaseVolume:  &[]float64{100.0}[0],
		QuoteVolume: &[]float64{5000000.0}[0],
		Timestamp:   time.Now(),
	}

	// 写入数据
	err := strategy.WritePrice(context.Background(), priceData)
	require.NoError(t, err)

	// 验证数据完整性
	cachedData, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)

	assert.Equal(t, "BTCUSDT", cachedData.Symbol)
	assert.Equal(t, 50000.0, cachedData.LastPrice)
	assert.Equal(t, 50001.0, *cachedData.AskPrice)
	assert.Equal(t, 49999.0, *cachedData.BidPrice)
	assert.Equal(t, 51000.0, *cachedData.High24h)
	assert.Equal(t, 49000.0, *cachedData.Low24h)
	assert.Equal(t, 1000.0, *cachedData.Change24h)
	assert.Equal(t, 100.0, *cachedData.BaseVolume)
	assert.Equal(t, 5000000.0, *cachedData.QuoteVolume)
}

func TestCacheConsistency_TTLExpiration(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建写穿透策略
	strategy := NewWriteThroughStrategy(manager)

	// 写入数据
	priceData := &PriceData{
		Symbol:    "BTCUSDT",
		LastPrice: 50000.0,
		Timestamp: time.Now(),
	}

	err := strategy.WritePrice(context.Background(), priceData)
	require.NoError(t, err)

	// 验证数据存在
	cachedData, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", cachedData.Symbol)

	// 等待 TTL 过期（在测试环境中可能需要调整）
	// 注意：在真实环境中，TTL 是 60 秒，这里我们使用 miniredis 的 FastForward 功能
	server.FastForward(61 * time.Second)

	// 验证数据已过期
	_, err = cache.GetPrice(context.Background(), "BTCUSDT")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestCacheConsistency_ErrorHandling(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建写穿透策略
	strategy := NewWriteThroughStrategy(manager)

	// 测试无效数据
	invalidData := &PriceData{
		Symbol: "", // 无效的 symbol
	}

	err := strategy.WritePrice(context.Background(), invalidData)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid price data")

	// 测试 nil 数据
	err = strategy.WritePrice(context.Background(), nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid price data")
}

func TestCacheConsistency_MultipleSymbols(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 创建一致性管理器
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建写穿透策略
	strategy := NewWriteThroughStrategy(manager)

	// 测试多个交易对
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT"}

	for i, symbol := range symbols {
		priceData := &PriceData{
			Symbol:    symbol,
			LastPrice: 50000.0 + float64(i*1000),
			Timestamp: time.Now(),
		}

		err := strategy.WritePrice(context.Background(), priceData)
		require.NoError(t, err)
	}

	// 验证所有数据都已保存
	for i, symbol := range symbols {
		cachedData, err := cache.GetPrice(context.Background(), symbol)
		require.NoError(t, err)
		assert.Equal(t, symbol, cachedData.Symbol)
		assert.Equal(t, 50000.0+float64(i*1000), cachedData.LastPrice)
	}
}

func TestCacheConsistency_BatchOperations(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 测试批量获取
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}

	// 先设置一些数据
	for i, symbol := range symbols {
		priceData := &PriceData{
			Symbol:    symbol,
			LastPrice: 50000.0 + float64(i*1000),
			Timestamp: time.Now(),
		}

		err := cache.SetPrice(context.Background(), priceData)
		require.NoError(t, err)
	}

	// 批量获取数据
	results, err := cache.GetMultiplePrices(context.Background(), symbols)
	require.NoError(t, err)

	// 验证结果
	assert.Len(t, results, 3)
	for i, symbol := range symbols {
		data, exists := results[symbol]
		require.True(t, exists)
		assert.Equal(t, symbol, data.Symbol)
		assert.Equal(t, 50000.0+float64(i*1000), data.LastPrice)
	}
}

func TestCacheConsistency_ActiveSymbols(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存
	cache := NewPriceCache(NewClient(client))

	// 设置活跃交易对列表
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT"}

	err := cache.SetActiveSymbols(context.Background(), symbols)
	require.NoError(t, err)

	// 获取活跃交易对列表
	activeSymbols, err := cache.GetActiveSymbols(context.Background())
	require.NoError(t, err)

	// 验证结果
	assert.Len(t, activeSymbols, 5)
	for _, symbol := range symbols {
		assert.Contains(t, activeSymbols, symbol)
	}
}
