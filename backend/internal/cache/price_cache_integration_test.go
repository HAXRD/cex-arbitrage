// +build integration

package cache

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupRealRedis 创建真实 Redis 客户端
func setupRealRedis(t *testing.T) *Client {
	// 从环境变量读取 Redis 连接信息
	redisHost := os.Getenv("TEST_REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}

	cfg := &Config{
		Host: redisHost,
		Port: 6379,
	}

	client, err := NewClient(cfg, nil)
	if err != nil {
		t.Skipf("跳过集成测试：Redis 服务不可用: %v", err)
		return nil
	}

	// 清理测试数据
	ctx := context.Background()
	pattern := "cryptosignal:*"
	iter := client.GetClient().Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		client.GetClient().Del(ctx, iter.Val())
	}

	return client
}

// TestPriceCache_Integration_SetGetPrice 集成测试：设置和获取价格
func TestPriceCache_Integration_SetGetPrice(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	client := setupRealRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("设置并获取价格", func(t *testing.T) {
		data := createTestPriceData("BTCUSDT")
		data.LastPrice = 50000.123456

		err := cache.SetPrice(ctx, data)
		require.NoError(t, err)

		// 获取
		retrieved, err := cache.GetPrice(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.Equal(t, "BTCUSDT", retrieved.Symbol)
		assert.Equal(t, 50000.123456, retrieved.LastPrice)
	})

	t.Run("验证 TTL 过期", func(t *testing.T) {
		data := createTestPriceData("EXPIRETEST")

		// 设置缓存
		err := cache.SetPrice(ctx, data)
		require.NoError(t, err)

		// 立即获取应该成功
		_, err = cache.GetPrice(ctx, "EXPIRETEST")
		assert.NoError(t, err)

		// 使用较短的 TTL 进行测试
		// 注意：这里需要直接设置短 TTL 来测试过期
		key := BuildLatestPriceKey("EXPIRETEST")
		client.GetClient().Expire(ctx, key, 1*time.Second)

		// 等待过期
		time.Sleep(2 * time.Second)

		// 再次获取应该失败
		_, err = cache.GetPrice(ctx, "EXPIRETEST")
		assert.Error(t, err)
	})
}

// TestPriceCache_Integration_BatchOperations 集成测试：批量操作
func TestPriceCache_Integration_BatchOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	client := setupRealRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("批量设置和获取价格", func(t *testing.T) {
		// 设置100个交易对的价格
		symbols := make([]string, 100)
		for i := 0; i < 100; i++ {
			symbol := fmt.Sprintf("SYMBOL%dUSDT", i)
			symbols[i] = symbol

			data := createTestPriceData(symbol)
			data.LastPrice = float64(1000 + i)

			err := cache.SetPrice(ctx, data)
			require.NoError(t, err)
		}

		// 批量获取
		start := time.Now()
		result, err := cache.GetMultiplePrices(ctx, symbols)
		duration := time.Since(start)

		require.NoError(t, err)
		assert.Len(t, result, 100)
		t.Logf("批量获取100个价格耗时: %v", duration)

		// 验证数据正确性
		for i, symbol := range symbols {
			data, exists := result[symbol]
			assert.True(t, exists)
			assert.Equal(t, float64(1000+i), data.LastPrice)
		}

		// 性能要求：< 100ms
		assert.Less(t, duration.Milliseconds(), int64(100))
	})
}

// TestPriceCache_Integration_Metrics 集成测试：指标缓存
func TestPriceCache_Integration_Metrics(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	client := setupRealRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("设置并获取指标数据", func(t *testing.T) {
		data := createTestMetricsData("BTCUSDT")
		data.PriceChange = 5.25
		data.Volume24h = 1234567.89

		err := cache.SetMetrics(ctx, data)
		require.NoError(t, err)

		// 获取
		retrieved, err := cache.GetMetrics(ctx, "BTCUSDT")
		require.NoError(t, err)
		assert.Equal(t, "BTCUSDT", retrieved.Symbol)
		assert.Equal(t, 5.25, retrieved.PriceChange)
		assert.Equal(t, 1234567.89, retrieved.Volume24h)
	})
}

// TestPriceCache_Integration_ActiveSymbols 集成测试：活跃交易对列表
func TestPriceCache_Integration_ActiveSymbols(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	client := setupRealRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("设置并获取活跃交易对列表", func(t *testing.T) {
		symbols := []string{"BTCUSDT", "ETHUSDT", "BNBUSDT", "ADAUSDT", "DOGEUSDT"}

		err := cache.SetActiveSymbols(ctx, symbols)
		require.NoError(t, err)

		// 获取
		retrieved, err := cache.GetActiveSymbols(ctx)
		require.NoError(t, err)
		assert.Len(t, retrieved, 5)

		// 验证所有交易对都在
		for _, symbol := range symbols {
			assert.Contains(t, retrieved, symbol)
		}
	})

	t.Run("更新活跃交易对列表", func(t *testing.T) {
		// 第一次设置
		symbols1 := []string{"BTC", "ETH", "BNB"}
		err := cache.SetActiveSymbols(ctx, symbols1)
		require.NoError(t, err)

		// 第二次设置（完全不同的列表）
		symbols2 := []string{"ADA", "DOGE", "SOL", "DOT"}
		err = cache.SetActiveSymbols(ctx, symbols2)
		require.NoError(t, err)

		// 获取应该只有新列表
		retrieved, err := cache.GetActiveSymbols(ctx)
		require.NoError(t, err)
		assert.Len(t, retrieved, 4)

		for _, symbol := range symbols2 {
			assert.Contains(t, retrieved, symbol)
		}

		// 旧列表的项不应该存在
		for _, symbol := range symbols1 {
			assert.NotContains(t, retrieved, symbol)
		}
	})
}

// TestPriceCache_Integration_ConcurrentAccess 集成测试：并发访问
func TestPriceCache_Integration_ConcurrentAccess(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	client := setupRealRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	cache := NewPriceCache(client)

	t.Run("并发设置价格", func(t *testing.T) {
		// 100个 goroutine 并发设置价格
		concurrency := 100
		errChan := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			go func(id int) {
				ctx := context.Background()
				symbol := fmt.Sprintf("CONCURRENT%dUSDT", id)
				data := createTestPriceData(symbol)
				data.LastPrice = float64(id)

				err := cache.SetPrice(ctx, data)
				errChan <- err
			}(i)
		}

		// 等待所有完成
		for i := 0; i < concurrency; i++ {
			err := <-errChan
			assert.NoError(t, err)
		}

		t.Logf("并发设置完成：%d 个并发", concurrency)
	})

	t.Run("并发获取价格", func(t *testing.T) {
		ctx := context.Background()

		// 先设置一些数据
		for i := 0; i < 10; i++ {
			symbol := fmt.Sprintf("READ%dUSDT", i)
			data := createTestPriceData(symbol)
			cache.SetPrice(ctx, data)
		}

		// 100个 goroutine 并发读取
		concurrency := 100
		errChan := make(chan error, concurrency)

		start := time.Now()
		for i := 0; i < concurrency; i++ {
			go func(id int) {
				ctx := context.Background()
				symbol := fmt.Sprintf("READ%dUSDT", id%10)
				_, err := cache.GetPrice(ctx, symbol)
				errChan <- err
			}(i)
		}

		// 等待所有完成
		for i := 0; i < concurrency; i++ {
			err := <-errChan
			assert.NoError(t, err)
		}

		duration := time.Since(start)
		t.Logf("并发读取完成：%d 个并发，耗时: %v", concurrency, duration)
	})
}

// TestPriceCache_Integration_MemoryUsage 集成测试：内存使用
func TestPriceCache_Integration_MemoryUsage(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过集成测试")
	}

	client := setupRealRedis(t)
	if client == nil {
		return
	}
	defer client.Close()

	cache := NewPriceCache(client)
	ctx := context.Background()

	t.Run("大量数据写入后的内存使用", func(t *testing.T) {
		// 获取初始内存使用
		memBefore, err := client.GetMemoryUsage(ctx)
		require.NoError(t, err)
		t.Logf("初始内存使用: %v", memBefore)

		// 写入1000条价格数据
		for i := 0; i < 1000; i++ {
			symbol := fmt.Sprintf("MEM%dUSDT", i)
			data := createTestPriceData(symbol)
			err := cache.SetPrice(ctx, data)
			require.NoError(t, err)
		}

		// 获取写入后的内存使用
		memAfter, err := client.GetMemoryUsage(ctx)
		require.NoError(t, err)
		t.Logf("写入1000条数据后的内存使用: %v", memAfter)
	})
}

