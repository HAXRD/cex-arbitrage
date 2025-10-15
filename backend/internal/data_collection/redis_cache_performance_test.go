package data_collection

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestRedisCache_Performance(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("单线程写入性能", func(t *testing.T) {
		operations := 1000
		start := time.Now()

		for i := 0; i < operations; i++ {
			key := fmt.Sprintf("perf:write:%d", i)
			value := fmt.Sprintf("value_%d", i)
			err := cache.Set(ctx, key, value, 5*time.Minute)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("单线程写入性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(operations))

		// 性能要求：至少每秒1000次操作
		assert.Greater(t, opsPerSecond, 1000.0, "写入性能不达标")
	})

	t.Run("单线程读取性能", func(t *testing.T) {
		// 先写入一些数据
		operations := 1000
		for i := 0; i < operations; i++ {
			key := fmt.Sprintf("perf:read:%d", i)
			value := fmt.Sprintf("value_%d", i)
			cache.Set(ctx, key, value, 5*time.Minute)
		}

		start := time.Now()

		for i := 0; i < operations; i++ {
			key := fmt.Sprintf("perf:read:%d", i)
			_, err := cache.Get(ctx, key)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("单线程读取性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(operations))

		// 性能要求：至少每秒2000次操作
		assert.Greater(t, opsPerSecond, 2000.0, "读取性能不达标")
	})

	t.Run("并发写入性能", func(t *testing.T) {
		concurrency := 10
		operationsPerGoroutine := 100
		totalOperations := concurrency * operationsPerGoroutine

		var wg sync.WaitGroup
		start := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					key := fmt.Sprintf("perf:concurrent:%d:%d", goroutineID, j)
					value := fmt.Sprintf("value_%d_%d", goroutineID, j)
					cache.Set(ctx, key, value, 5*time.Minute)
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)
		opsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("并发写入性能:")
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总操作数: %d", totalOperations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(totalOperations))

		// 性能要求：至少每秒5000次操作
		assert.Greater(t, opsPerSecond, 5000.0, "并发写入性能不达标")
	})

	t.Run("批量写入性能", func(t *testing.T) {
		batchSize := 100
		batches := 10
		totalOperations := batchSize * batches

		start := time.Now()

		for i := 0; i < batches; i++ {
			data := make([]CacheData, batchSize)
			for j := 0; j < batchSize; j++ {
				key := fmt.Sprintf("perf:batch:%d:%d", i, j)
				value := fmt.Sprintf("value_%d_%d", i, j)
				data[j] = CacheData{
					Key:       key,
					Value:     value,
					TTL:       5 * time.Minute,
					Timestamp: time.Now(),
				}
			}

			result, err := cache.SetBatch(ctx, data)
			require.NoError(t, err)
			assert.Equal(t, batchSize, result.SuccessCount)
		}

		duration := time.Since(start)
		opsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("批量写入性能:")
		t.Logf("  批次大小: %d", batchSize)
		t.Logf("  批次数: %d", batches)
		t.Logf("  总操作数: %d", totalOperations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(totalOperations))

		// 性能要求：至少每秒10000次操作
		assert.Greater(t, opsPerSecond, 10000.0, "批量写入性能不达标")
	})

	t.Run("价格数据写入性能", func(t *testing.T) {
		operations := 1000
		start := time.Now()

		for i := 0; i < operations; i++ {
			price := &PriceData{
				Symbol:    fmt.Sprintf("BTCUSDT_%d", i),
				Price:     50000.0 + float64(i),
				BidPrice:  49999.0 + float64(i),
				AskPrice:  50001.0 + float64(i),
				Volume:    1000.0 + float64(i),
				Timestamp: time.Now(),
				Source:    "test",
				Latency:   10 * time.Millisecond,
			}

			err := cache.SetPrice(ctx, price.Symbol, price)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("价格数据写入性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(operations))

		// 性能要求：至少每秒800次操作
		assert.Greater(t, opsPerSecond, 800.0, "价格数据写入性能不达标")
	})

	t.Run("变化率数据写入性能", func(t *testing.T) {
		operations := 1000
		start := time.Now()

		for i := 0; i < operations; i++ {
			rate := &ProcessedPriceChangeRate{
				Symbol:     fmt.Sprintf("BTCUSDT_%d", i),
				TimeWindow: "1m",
				ChangeRate: 2.5 + float64(i)*0.01,
				StartPrice: 50000.0,
				EndPrice:   51250.0 + float64(i),
				Timestamp:  time.Now(),
				IsValid:    true,
				IsAnomaly:  false,
			}

			err := cache.SetChangeRate(ctx, rate.Symbol, TimeWindow1m, rate)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("变化率数据写入性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(operations))

		// 性能要求：至少每秒800次操作
		assert.Greater(t, opsPerSecond, 800.0, "变化率数据写入性能不达标")
	})

	t.Run("混合操作性能", func(t *testing.T) {
		operations := 1000
		start := time.Now()

		for i := 0; i < operations; i++ {
			// 写入操作
			key := fmt.Sprintf("perf:mixed:%d", i)
			value := fmt.Sprintf("value_%d", i)
			cache.Set(ctx, key, value, 5*time.Minute)

			// 读取操作
			if i%2 == 0 {
				cache.Get(ctx, key)
			}

			// 删除操作
			if i%3 == 0 {
				cache.Delete(ctx, key)
			}
		}

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("混合操作性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(operations))

		// 性能要求：至少每秒1500次操作
		assert.Greater(t, opsPerSecond, 1500.0, "混合操作性能不达标")
	})
}

func TestRedisCache_MemoryUsage(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("内存使用测试", func(t *testing.T) {
		// 写入大量数据
		operations := 10000
		start := time.Now()

		for i := 0; i < operations; i++ {
			key := fmt.Sprintf("memory:test:%d", i)
			value := fmt.Sprintf("large_value_%d_%s", i, string(make([]byte, 1000))) // 1KB数据
			err := cache.Set(ctx, key, value, 5*time.Minute)
			require.NoError(t, err)
		}

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("内存使用测试:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  数据大小: 1KB per key")
		t.Logf("  总数据量: %.2f MB", float64(operations)/1024)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 验证数据完整性
		verifyCount := 100
		for i := 0; i < verifyCount; i++ {
			key := fmt.Sprintf("memory:test:%d", i)
			value, err := cache.Get(ctx, key)
			require.NoError(t, err)
			assert.Contains(t, value, fmt.Sprintf("large_value_%d", i))
		}

		t.Logf("数据完整性验证: %d/%d 通过", verifyCount, verifyCount)
	})
}

func TestRedisCache_StressTest(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("压力测试", func(t *testing.T) {
		concurrency := 50
		operationsPerGoroutine := 200
		totalOperations := concurrency * operationsPerGoroutine

		var wg sync.WaitGroup
		errorCount := 0
		var mu sync.Mutex

		start := time.Now()

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < operationsPerGoroutine; j++ {
					key := fmt.Sprintf("stress:%d:%d", goroutineID, j)
					value := fmt.Sprintf("stress_value_%d_%d", goroutineID, j)

					// 随机操作
					switch j % 4 {
					case 0: // 写入
						err := cache.Set(ctx, key, value, 5*time.Minute)
						if err != nil {
							mu.Lock()
							errorCount++
							mu.Unlock()
						}
					case 1: // 读取
						_, err := cache.Get(ctx, key)
						if err != nil && err.Error() != "redis: nil" {
							mu.Lock()
							errorCount++
							mu.Unlock()
						}
					case 2: // 检查存在性
						_, err := cache.Exists(ctx, key)
						if err != nil {
							mu.Lock()
							errorCount++
							mu.Unlock()
						}
					case 3: // 删除
						err := cache.Delete(ctx, key)
						if err != nil {
							mu.Lock()
							errorCount++
							mu.Unlock()
						}
					}
				}
			}(i)
		}

		wg.Wait()
		duration := time.Since(start)
		opsPerSecond := float64(totalOperations) / duration.Seconds()
		errorRate := float64(errorCount) / float64(totalOperations) * 100

		t.Logf("压力测试结果:")
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总操作数: %d", totalOperations)
		t.Logf("  错误数: %d", errorCount)
		t.Logf("  错误率: %.2f%%", errorRate)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能要求：错误率低于5%
		assert.Less(t, errorRate, 5.0, "错误率过高")
		// 性能要求：至少每秒10000次操作
		assert.Greater(t, opsPerSecond, 10000.0, "压力测试性能不达标")
	})
}
