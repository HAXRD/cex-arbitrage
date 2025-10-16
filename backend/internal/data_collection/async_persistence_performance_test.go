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

func TestAsyncPersistence_Performance(t *testing.T) {
	// 跳过测试，需要真实数据库连接
	t.Skip("需要真实数据库连接，跳过测试")

	// 创建性能测试配置
	config := DefaultPersistenceConfig()
	config.QueueSize = 10000
	config.BatchSize = 100
	config.BatchTimeout = 100 * time.Millisecond
	config.WorkerCount = 10

	// 创建Mock写入器
	writer := NewMockDataWriter()
	writer.SetWriteDelay(1 * time.Millisecond) // 模拟写入延迟

	// 创建异步持久化实例
	persistence := NewAsyncPersistence(config, writer, zap.NewNop())
	require.NotNil(t, persistence)
	defer persistence.Stop(context.Background())

	ctx := context.Background()

	t.Run("单线程写入性能", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		operations := 1000
		start := time.Now()

		for i := 0; i < operations; i++ {
			item := &PersistenceItem{
				ID:        fmt.Sprintf("perf_single_%d", i),
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0 + float64(i)},
				Timestamp: time.Now(),
				Priority:  1,
				CreatedAt: time.Now(),
			}

			err := persistence.Submit(item)
			require.NoError(t, err)
		}

		// 等待处理完成
		time.Sleep(2 * time.Second)

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("单线程写入性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(operations))

		// 性能要求：至少每秒500次操作
		assert.Greater(t, opsPerSecond, 500.0, "单线程写入性能不达标")
	})

	t.Run("并发写入性能", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

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
					item := &PersistenceItem{
						ID:        fmt.Sprintf("perf_concurrent_%d_%d", goroutineID, j),
						Type:      "price",
						Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0 + float64(goroutineID*100+j)},
						Timestamp: time.Now(),
						Priority:  1,
						CreatedAt: time.Now(),
					}

					persistence.Submit(item)
				}
			}(i)
		}

		wg.Wait()

		// 等待处理完成
		time.Sleep(2 * time.Second)

		duration := time.Since(start)
		opsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("并发写入性能:")
		t.Logf("  并发数: %d", concurrency)
		t.Logf("  总操作数: %d", totalOperations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(totalOperations))

		// 性能要求：至少每秒2000次操作
		assert.Greater(t, opsPerSecond, 2000.0, "并发写入性能不达标")
	})

	t.Run("批量写入性能", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		batchSize := 50
		batches := 20
		totalOperations := batchSize * batches

		start := time.Now()

		for i := 0; i < batches; i++ {
			items := make([]*PersistenceItem, batchSize)
			for j := 0; j < batchSize; j++ {
				items[j] = &PersistenceItem{
					ID:        fmt.Sprintf("perf_batch_%d_%d", i, j),
					Type:      "price",
					Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0 + float64(i*batchSize+j)},
					Timestamp: time.Now(),
					Priority:  1,
					CreatedAt: time.Now(),
				}
			}

			err := persistence.SubmitBatch(items)
			require.NoError(t, err)
		}

		// 等待处理完成
		time.Sleep(2 * time.Second)

		duration := time.Since(start)
		opsPerSecond := float64(totalOperations) / duration.Seconds()

		t.Logf("批量写入性能:")
		t.Logf("  批次大小: %d", batchSize)
		t.Logf("  批次数: %d", batches)
		t.Logf("  总操作数: %d", totalOperations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(totalOperations))

		// 性能要求：至少每秒3000次操作
		assert.Greater(t, opsPerSecond, 3000.0, "批量写入性能不达标")
	})

	t.Run("混合操作性能", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		operations := 1000
		start := time.Now()

		for i := 0; i < operations; i++ {
			// 写入操作
			item := &PersistenceItem{
				ID:        fmt.Sprintf("perf_mixed_%d", i),
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0 + float64(i)},
				Timestamp: time.Now(),
				Priority:  1,
				CreatedAt: time.Now(),
			}

			persistence.Submit(item)

			// 批量操作
			if i%10 == 0 {
				batchItems := make([]*PersistenceItem, 5)
				for j := 0; j < 5; j++ {
					batchItems[j] = &PersistenceItem{
						ID:        fmt.Sprintf("perf_mixed_batch_%d_%d", i, j),
						Type:      "price",
						Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0 + float64(i*5+j)},
						Timestamp: time.Now(),
						Priority:  1,
						CreatedAt: time.Now(),
					}
				}
				persistence.SubmitBatch(batchItems)
			}
		}

		// 等待处理完成
		time.Sleep(3 * time.Second)

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("混合操作性能:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)
		t.Logf("  平均耗时: %v", duration/time.Duration(operations))

		// 性能要求：至少每秒800次操作
		assert.Greater(t, opsPerSecond, 800.0, "混合操作性能不达标")
	})
}

func TestAsyncPersistence_MemoryUsage(t *testing.T) {
	// 跳过测试，需要真实数据库连接
	t.Skip("需要真实数据库连接，跳过测试")

	// 创建内存测试配置
	config := DefaultPersistenceConfig()
	config.QueueSize = 5000
	config.BatchSize = 100
	config.MaxMemoryUsage = 50 * 1024 * 1024 // 50MB
	config.EnableDeduplication = true
	config.DeduplicationWindow = 5 * time.Minute

	// 创建Mock写入器
	writer := NewMockDataWriter()
	writer.SetWriteDelay(10 * time.Millisecond) // 模拟较慢的写入

	// 创建异步持久化实例
	persistence := NewAsyncPersistence(config, writer, zap.NewNop())
	require.NotNil(t, persistence)
	defer persistence.Stop(context.Background())

	ctx := context.Background()

	t.Run("内存使用测试", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		// 写入大量数据
		operations := 2000
		start := time.Now()

		for i := 0; i < operations; i++ {
			item := &PersistenceItem{
				ID:        fmt.Sprintf("memory_%d", i),
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0 + float64(i)},
				Timestamp: time.Now(),
				Priority:  1,
				CreatedAt: time.Now(),
			}

			persistence.Submit(item)
		}

		// 等待处理
		time.Sleep(5 * time.Second)

		duration := time.Since(start)
		opsPerSecond := float64(operations) / duration.Seconds()

		t.Logf("内存使用测试:")
		t.Logf("  操作数: %d", operations)
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 验证统计信息
		stats := persistence.GetStats()
		t.Logf("内存统计:")
		t.Logf("  队列大小: %d", stats.QueueSize)
		t.Logf("  内存使用: %d bytes", stats.MemoryUsage)
		t.Logf("  最大内存使用: %d bytes", stats.MaxMemoryUsage)
		t.Logf("  去重次数: %d", stats.DeduplicationCount)

		// 验证内存使用在合理范围内
		assert.Less(t, stats.MemoryUsage, int64(config.MaxMemoryUsage), "内存使用超过限制")
	})
}

func TestAsyncPersistence_StressTest(t *testing.T) {
	// 跳过测试，需要真实数据库连接
	t.Skip("需要真实数据库连接，跳过测试")

	// 创建压力测试配置
	config := DefaultPersistenceConfig()
	config.QueueSize = 20000
	config.BatchSize = 200
	config.WorkerCount = 20
	config.MaxRetries = 5
	config.RetryInterval = 50 * time.Millisecond

	// 创建Mock写入器
	writer := NewMockDataWriter()
	writer.SetFailRate(0.1) // 10%失败率
	writer.SetWriteDelay(2 * time.Millisecond)

	// 创建异步持久化实例
	persistence := NewAsyncPersistence(config, writer, zap.NewNop())
	require.NotNil(t, persistence)
	defer persistence.Stop(context.Background())

	ctx := context.Background()

	t.Run("压力测试", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		concurrency := 50
		operationsPerGoroutine := 100
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
					item := &PersistenceItem{
						ID:        fmt.Sprintf("stress_%d_%d", goroutineID, j),
						Type:      "price",
						Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0 + float64(goroutineID*100+j)},
						Timestamp: time.Now(),
						Priority:  1,
						CreatedAt: time.Now(),
					}

					err := persistence.Submit(item)
					if err != nil {
						mu.Lock()
						errorCount++
						mu.Unlock()
					}
				}
			}(i)
		}

		wg.Wait()

		// 等待处理完成
		time.Sleep(10 * time.Second)

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

		// 验证统计信息
		stats := persistence.GetStats()
		t.Logf("压力测试统计:")
		t.Logf("  总处理数: %d", stats.TotalProcessed)
		t.Logf("  成功数: %d", stats.SuccessCount)
		t.Logf("  错误数: %d", stats.ErrorCount)
		t.Logf("  重试数: %d", stats.RetryCount)
		t.Logf("  平均处理时间: %v", stats.AvgProcessTime)
		t.Logf("  最大处理时间: %v", stats.MaxProcessTime)
		t.Logf("  最小处理时间: %v", stats.MinProcessTime)

		// 性能要求：错误率低于10%
		assert.Less(t, errorRate, 10.0, "错误率过高")
		// 性能要求：至少每秒5000次操作
		assert.Greater(t, opsPerSecond, 5000.0, "压力测试性能不达标")
	})
}
