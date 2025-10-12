package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zaptest"
	"gorm.io/gorm"
)

func TestRetryManager_Execute_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rm := NewRetryManager(DefaultRetryConfig(), logger)

	attempts := 0
	err := rm.Execute("test_operation", func() error {
		attempts++
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, attempts)
}

func TestRetryManager_Execute_RetryableError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &RetryConfig{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}
	rm := NewRetryManager(config, logger)

	attempts := 0
	err := rm.Execute("test_operation", func() error {
		attempts++
		if attempts < 3 {
			return fmt.Errorf("connection refused")
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 3, attempts)
}

func TestRetryManager_Execute_NonRetryableError(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &RetryConfig{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}
	rm := NewRetryManager(config, logger)

	attempts := 0
	err := rm.Execute("test_operation", func() error {
		attempts++
		return fmt.Errorf("invalid input")
	})

	require.Error(t, err)
	assert.Equal(t, 1, attempts)
	assert.Contains(t, err.Error(), "invalid input")
}

func TestRetryManager_Execute_MaxRetriesExceeded(t *testing.T) {
	logger := zaptest.NewLogger(t)
	config := &RetryConfig{
		MaxRetries:    2,
		InitialDelay:  10 * time.Millisecond,
		MaxDelay:      100 * time.Millisecond,
		BackoffFactor: 2.0,
		Jitter:        false,
	}
	rm := NewRetryManager(config, logger)

	attempts := 0
	err := rm.Execute("test_operation", func() error {
		attempts++
		return fmt.Errorf("connection refused")
	})

	require.Error(t, err)
	assert.Equal(t, 3, attempts) // 1 initial + 2 retries
	assert.Contains(t, err.Error(), "failed after 3 attempts")
}

func TestRetryManager_ExecuteWithContext_Cancelled(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rm := NewRetryManager(DefaultRetryConfig(), logger)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // 立即取消

	attempts := 0
	err := rm.ExecuteWithContext(ctx, "test_operation", func(ctx context.Context) error {
		attempts++
		return fmt.Errorf("connection refused")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
	assert.Equal(t, 0, attempts) // 应该没有执行
}

func TestRetryManager_ExecuteWithContext_Timeout(t *testing.T) {
	logger := zaptest.NewLogger(t)
	rm := NewRetryManager(DefaultRetryConfig(), logger)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	attempts := 0
	err := rm.ExecuteWithContext(ctx, "test_operation", func(ctx context.Context) error {
		attempts++
		time.Sleep(100 * time.Millisecond) // 模拟慢操作
		return fmt.Errorf("connection refused")
	})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "operation cancelled")
}

func TestDatabaseRetryManager_ExecuteTransaction(t *testing.T) {
	logger := zaptest.NewLogger(t)
	drm := NewDatabaseRetryManager(logger)

	// 模拟数据库操作
	attempts := 0
	err := drm.ExecuteTransaction("test_transaction", func(tx *gorm.DB) error {
		attempts++
		if attempts < 2 {
			return fmt.Errorf("connection refused")
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestCacheRetryManager_ExecuteCacheOperation(t *testing.T) {
	logger := zaptest.NewLogger(t)
	crm := NewCacheRetryManager(logger)

	// 模拟缓存操作
	attempts := 0
	err := crm.ExecuteCacheOperation("test_cache_operation", func() error {
		attempts++
		if attempts < 2 {
			return fmt.Errorf("connection refused")
		}
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 2, attempts)
}

func TestCircuitBreaker_Execute_Success(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(3, 100*time.Millisecond, logger)

	attempts := 0
	err := cb.Execute("test_operation", func() error {
		attempts++
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, 1, attempts)
	assert.Equal(t, CircuitStateClosed, cb.GetState())
}

func TestCircuitBreaker_Execute_OpenCircuit(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(2, 100*time.Millisecond, logger)

	// 连续失败，触发熔断
	for i := 0; i < 3; i++ {
		err := cb.Execute("test_operation", func() error {
			return fmt.Errorf("connection refused")
		})
		require.Error(t, err)
	}

	// 熔断器应该开启
	assert.Equal(t, CircuitStateOpen, cb.GetState())
	assert.Equal(t, 3, cb.GetFailureCount())

	// 再次执行应该被拒绝
	err := cb.Execute("test_operation", func() error {
		return nil
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "circuit breaker is open")
}

func TestCircuitBreaker_Execute_HalfOpenRecovery(t *testing.T) {
	logger := zaptest.NewLogger(t)
	cb := NewCircuitBreaker(2, 50*time.Millisecond, logger)

	// 触发熔断
	for i := 0; i < 3; i++ {
		cb.Execute("test_operation", func() error {
			return fmt.Errorf("connection refused")
		})
	}

	assert.Equal(t, CircuitStateOpen, cb.GetState())

	// 等待重置超时
	time.Sleep(60 * time.Millisecond)

	// 成功操作应该恢复熔断器
	err := cb.Execute("test_operation", func() error {
		return nil
	})

	require.NoError(t, err)
	assert.Equal(t, CircuitStateClosed, cb.GetState())
	assert.Equal(t, 0, cb.GetFailureCount())
}

func TestBloomFilter_AddAndContains(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)
	bf := NewBloomFilter(client, "test:bloom", 1000, 0.01, logger)

	// 添加元素
	err := bf.Add(context.Background(), "BTCUSDT")
	require.NoError(t, err)

	err = bf.Add(context.Background(), "ETHUSDT")
	require.NoError(t, err)

	// 检查存在的元素
	exists, err := bf.Contains(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.True(t, exists)

	exists, err = bf.Contains(context.Background(), "ETHUSDT")
	require.NoError(t, err)
	assert.True(t, exists)

	// 检查不存在的元素（可能返回 false positive）
	exists, err = bf.Contains(context.Background(), "ADAUSDT")
	require.NoError(t, err)
	// 注意：布隆过滤器可能返回 false positive，但不返回 false negative
}

func TestNullValueCache_SetAndCheck(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)
	nvc := NewNullValueCache(client, 60*time.Second, logger)

	// 设置空值
	err := nvc.SetNullValue(context.Background(), "INVALID")
	require.NoError(t, err)

	// 检查空值
	assert.True(t, nvc.IsNullValue(context.Background(), "INVALID"))
	assert.False(t, nvc.IsNullValue(context.Background(), "BTCUSDT"))

	// 移除空值
	err = nvc.RemoveNullValue(context.Background(), "INVALID")
	require.NoError(t, err)

	assert.False(t, nvc.IsNullValue(context.Background(), "INVALID"))
}

func TestCacheProtection_GetPriceWithProtection(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存和布隆过滤器
	cache := NewPriceCache(NewClient(client))
	bf := NewBloomFilter(client, "test:bloom", 1000, 0.01, logger)
	nvc := NewNullValueCache(client, 60*time.Second, logger)

	// 创建缓存保护
	cp := NewCacheProtection(cache, bf, nvc, logger)

	// 添加符号到布隆过滤器
	err := bf.Add(context.Background(), "BTCUSDT")
	require.NoError(t, err)

	// 设置价格数据
	priceData := &PriceData{
		Symbol:    "BTCUSDT",
		LastPrice: 50000.0,
		Timestamp: time.Now(),
	}

	err = cp.SetPriceWithProtection(context.Background(), priceData)
	require.NoError(t, err)

	// 获取价格数据
	result, err := cp.GetPriceWithProtection(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", result.Symbol)
	assert.Equal(t, 50000.0, result.LastPrice)
}

func TestCacheProtection_GetPriceWithProtection_NotFound(t *testing.T) {
	// 使用 miniredis 进行测试
	server := miniredis.RunT(t)
	defer server.Close()

	client := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer client.Close()

	logger := zaptest.NewLogger(t)

	// 创建缓存和布隆过滤器
	cache := NewPriceCache(NewClient(client))
	bf := NewBloomFilter(client, "test:bloom", 1000, 0.01, logger)
	nvc := NewNullValueCache(client, 60*time.Second, logger)

	// 创建缓存保护
	cp := NewCacheProtection(cache, bf, nvc, logger)

	// 设置空值缓存
	err := nvc.SetNullValue(context.Background(), "INVALID")
	require.NoError(t, err)

	// 尝试获取不存在的价格
	result, err := cp.GetPriceWithProtection(context.Background(), "INVALID")
	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")
}

func TestWriteThroughStrategy_WritePrice(t *testing.T) {
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

	// 创建一致性管理器（这里需要模拟数据库）
	manager := &CacheConsistencyManager{
		cache:  cache,
		logger: logger,
	}

	// 创建写穿透策略
	strategy := NewWriteThroughStrategy(manager)

	// 写入价格数据
	priceData := &PriceData{
		Symbol:    "BTCUSDT",
		LastPrice: 50000.0,
		Timestamp: time.Now(),
	}

	err := strategy.WritePrice(context.Background(), priceData)
	require.NoError(t, err)

	// 验证缓存中的数据
	result, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", result.Symbol)
	assert.Equal(t, 50000.0, result.LastPrice)
}

func TestWriteBehindStrategy_WritePrice(t *testing.T) {
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

	// 写入价格数据
	priceData := &PriceData{
		Symbol:    "BTCUSDT",
		LastPrice: 50000.0,
		Timestamp: time.Now(),
	}

	err := strategy.WritePrice(context.Background(), priceData)
	require.NoError(t, err)

	// 验证缓存中的数据
	result, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", result.Symbol)
	assert.Equal(t, 50000.0, result.LastPrice)

	// 等待异步数据库写入完成
	time.Sleep(100 * time.Millisecond)
}

func TestCacheAsideStrategy_ReadPrice(t *testing.T) {
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

	// 读取价格数据（应该从数据库读取并更新缓存）
	result, err := strategy.ReadPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", result.Symbol)
	assert.Equal(t, 50000.0, result.LastPrice)

	// 等待异步缓存更新完成
	time.Sleep(100 * time.Millisecond)

	// 再次读取应该从缓存获取
	result2, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", result2.Symbol)
	assert.Equal(t, 50000.0, result2.LastPrice)
}

func TestCacheInvalidationStrategy_InvalidatePrice(t *testing.T) {
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
	result, err := cache.GetPrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)
	assert.Equal(t, "BTCUSDT", result.Symbol)

	// 失效缓存
	err = strategy.InvalidatePrice(context.Background(), "BTCUSDT")
	require.NoError(t, err)

	// 验证数据已被删除
	_, err = cache.GetPrice(context.Background(), "BTCUSDT")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDatabaseErrorHandling(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "Connection error",
			err:      fmt.Errorf("connection refused"),
			expected: true,
		},
		{
			name:     "Timeout error",
			err:      fmt.Errorf("i/o timeout"),
			expected: true,
		},
		{
			name:     "Network unreachable",
			err:      fmt.Errorf("network is unreachable"),
			expected: true,
		},
		{
			name:     "Context deadline exceeded",
			err:      fmt.Errorf("context deadline exceeded"),
			expected: true,
		},
		{
			name:     "Invalid input",
			err:      fmt.Errorf("invalid input"),
			expected: false,
		},
		{
			name:     "Duplicate key",
			err:      fmt.Errorf("duplicate key violation"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			logger := zaptest.NewLogger(t)
			rm := NewRetryManager(DefaultRetryConfig(), logger)

			// 使用反射访问私有方法（仅用于测试）
			retryable := rm.isRetryableError(tt.err)
			assert.Equal(t, tt.expected, retryable)
		})
	}
}
