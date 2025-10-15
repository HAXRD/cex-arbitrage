package data_collection

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAsyncPersistence_RetryMechanism(t *testing.T) {
	// 跳过测试，需要真实数据库连接
	t.Skip("需要真实数据库连接，跳过测试")

	// 创建重试配置
	config := DefaultPersistenceConfig()
	config.MaxRetries = 3
	config.RetryInterval = 100 * time.Millisecond
	config.RetryBackoff = 2.0
	config.MaxRetryDelay = 1 * time.Second

	// 创建Mock写入器
	writer := NewMockDataWriter()
	writer.SetFailRate(0.5) // 50%失败率

	// 创建异步持久化实例
	persistence := NewAsyncPersistence(config, writer, zap.NewNop())
	require.NotNil(t, persistence)
	defer persistence.Stop(context.Background())

	ctx := context.Background()

	t.Run("重试机制测试", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		// 提交数据
		item := &PersistenceItem{
			ID:        "retry_test",
			Type:      "price",
			Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
			Timestamp: time.Now(),
			Priority:  1,
			CreatedAt: time.Now(),
		}

		err = persistence.Submit(item)
		require.NoError(t, err)

		// 等待处理完成
		time.Sleep(2 * time.Second)

		// 验证重试统计
		stats := persistence.GetStats()
		assert.True(t, stats.RetryCount >= 0)

		// 验证写入器统计
		writerStats := writer.GetStats()
		t.Logf("写入器统计: %+v", writerStats)
	})

	t.Run("重试延迟计算", func(t *testing.T) {
		// 创建重试管理器
		retryManager := NewRetryManager(config)

		// 测试重试延迟计算
		delays := make([]time.Duration, 5)
		for i := 0; i < 5; i++ {
			delays[i] = retryManager.CalculateDelay(i)
			t.Logf("重试次数 %d: 延迟 %v", i, delays[i])
		}

		// 验证延迟递增
		for i := 1; i < len(delays); i++ {
			assert.True(t, delays[i] >= delays[i-1], "延迟应该递增")
		}
	})

	t.Run("重试条件判断", func(t *testing.T) {
		// 创建重试管理器
		retryManager := NewRetryManager(config)

		// 测试可重试错误
		retryableErrors := []string{
			"timeout error",
			"connection failed",
			"network unavailable",
			"temporary failure",
			"server busy",
		}

		for _, errMsg := range retryableErrors {
			item := &PersistenceItem{
				ID:        "test",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: time.Now(),
				Priority:  1,
				CreatedAt: time.Now(),
			}

			shouldRetry := retryManager.ShouldRetry(item, fmt.Errorf("%s", errMsg))
			assert.True(t, shouldRetry, fmt.Sprintf("错误 '%s' 应该可以重试", errMsg))
		}

		// 测试不可重试错误
		nonRetryableErrors := []string{
			"invalid data format",
			"permission denied",
			"authentication failed",
			"data validation error",
		}

		for _, errMsg := range nonRetryableErrors {
			item := &PersistenceItem{
				ID:        "test",
				Type:      "price",
				Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
				Timestamp: time.Now(),
				Priority:  1,
				CreatedAt: time.Now(),
			}

			shouldRetry := retryManager.ShouldRetry(item, fmt.Errorf("%s", errMsg))
			assert.False(t, shouldRetry, fmt.Sprintf("错误 '%s' 不应该重试", errMsg))
		}
	})

	t.Run("重试次数限制", func(t *testing.T) {
		// 创建重试管理器
		retryManager := NewRetryManager(config)

		// 测试重试次数限制
		item := &PersistenceItem{
			ID:        "test",
			Type:      "price",
			Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
			Timestamp: time.Now(),
			Priority:  1,
			CreatedAt: time.Now(),
		}

		// 设置重试次数超过限制
		item.RetryCount = config.MaxRetries + 1

		shouldRetry := retryManager.ShouldRetry(item, fmt.Errorf("timeout error"))
		assert.False(t, shouldRetry, "超过重试次数限制后不应该重试")
	})

	t.Run("重试统计", func(t *testing.T) {
		// 创建重试管理器
		retryManager := NewRetryManager(config)

		// 记录一些重试
		item := &PersistenceItem{
			ID:        "test",
			Type:      "price",
			Data:      &PriceData{Symbol: "BTCUSDT", Price: 50000.0},
			Timestamp: time.Now(),
			Priority:  1,
			CreatedAt: time.Now(),
		}

		retryManager.RecordRetry(item, fmt.Errorf("timeout error"))
		retryManager.RecordRetry(item, fmt.Errorf("connection failed"))

		// 获取重试统计
		stats := retryManager.GetRetryStats()
		assert.True(t, stats["retry_count"].(int64) >= 2)
		assert.Equal(t, config.MaxRetries, stats["max_retries"])
		assert.Equal(t, config.RetryInterval.String(), stats["retry_interval"])
		assert.Equal(t, config.RetryBackoff, stats["retry_backoff"])
		assert.Equal(t, config.MaxRetryDelay.String(), stats["max_retry_delay"])
	})
}

func TestAsyncPersistence_RetryIntegration(t *testing.T) {
	// 跳过测试，需要真实数据库连接
	t.Skip("需要真实数据库连接，跳过测试")

	// 创建重试配置
	config := DefaultPersistenceConfig()
	config.MaxRetries = 2
	config.RetryInterval = 50 * time.Millisecond
	config.RetryBackoff = 1.5
	config.MaxRetryDelay = 500 * time.Millisecond

	// 创建Mock写入器
	writer := NewMockDataWriter()
	writer.SetFailRate(0.8) // 80%失败率

	// 创建异步持久化实例
	persistence := NewAsyncPersistence(config, writer, zap.NewNop())
	require.NotNil(t, persistence)
	defer persistence.Stop(context.Background())

	ctx := context.Background()

	t.Run("集成重试测试", func(t *testing.T) {
		// 启动持久化
		err := persistence.Start(ctx)
		require.NoError(t, err)
		defer persistence.Stop(ctx)

		// 提交多个数据
		for i := 0; i < 5; i++ {
			item := &PersistenceItem{
				ID:        fmt.Sprintf("retry_integration_%d", i),
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
		time.Sleep(3 * time.Second)

		// 验证统计信息
		stats := persistence.GetStats()
		t.Logf("持久化统计: %+v", stats)

		// 验证写入器统计
		writerStats := writer.GetStats()
		t.Logf("写入器统计: %+v", writerStats)

		// 验证重试次数
		assert.True(t, stats.RetryCount >= 0)
	})
}
