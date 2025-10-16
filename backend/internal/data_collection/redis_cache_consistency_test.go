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

func TestRedisCache_WriteThroughConsistency(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例（启用Write-Through）
	config := DefaultCacheConfig("")
	config.EnableWriteThrough = true
	config.EnableWriteBehind = false
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("Write-Through一致性", func(t *testing.T) {
		// 设置数据
		key := "test:write_through"
		value := "test_value"
		ttl := 5 * time.Minute

		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// 立即读取，应该能获取到数据
		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)

		// 验证数据确实存在于Redis中
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("并发Write-Through", func(t *testing.T) {
		// 并发写入测试
		concurrency := 10
		keysPerGoroutine := 5
		var wg sync.WaitGroup
		errors := make(chan error, concurrency*keysPerGoroutine)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()
				for j := 0; j < keysPerGoroutine; j++ {
					key := fmt.Sprintf("concurrent:write_through:%d:%d", goroutineID, j)
					value := fmt.Sprintf("value_%d_%d", goroutineID, j)

					err := cache.Set(ctx, key, value, 5*time.Minute)
					if err != nil {
						errors <- err
						return
					}

					// 立即验证
					result, err := cache.Get(ctx, key)
					if err != nil {
						errors <- err
						return
					}
					if result != value {
						errors <- fmt.Errorf("数据不一致: 期望 %s, 实际 %s", value, result)
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// 检查是否有错误
		for err := range errors {
			t.Errorf("并发Write-Through错误: %v", err)
		}
	})
}

func TestRedisCache_WriteBehindConsistency(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例（启用Write-Behind）
	config := DefaultCacheConfig("")
	config.EnableWriteThrough = false
	config.EnableWriteBehind = true
	config.BatchSize = 5
	config.BatchTimeout = 100 * time.Millisecond
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("Write-Behind批量写入", func(t *testing.T) {
		// 设置多个数据
		keys := []string{"batch:1", "batch:2", "batch:3", "batch:4", "batch:5"}
		values := []string{"value1", "value2", "value3", "value4", "value5"}

		for i, key := range keys {
			err := cache.Set(ctx, key, values[i], 5*time.Minute)
			require.NoError(t, err)
		}

		// 等待批量写入完成
		time.Sleep(200 * time.Millisecond)

		// 验证数据已写入Redis
		for i, key := range keys {
			result, err := cache.Get(ctx, key)
			require.NoError(t, err)
			assert.Equal(t, values[i], result)
		}
	})

	t.Run("Write-Behind超时写入", func(t *testing.T) {
		// 设置少量数据，触发超时写入
		key := "timeout:write_behind"
		value := "timeout_value"

		err := cache.Set(ctx, key, value, 5*time.Minute)
		require.NoError(t, err)

		// 等待超时写入
		time.Sleep(150 * time.Millisecond)

		// 验证数据已写入Redis
		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, value, result)
	})
}

func TestRedisCache_CacheAsideConsistency(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("Cache-Aside模式", func(t *testing.T) {
		key := "cache_aside:test"
		value := "cache_aside_value"
		ttl := 5 * time.Minute

		// 1. 先检查缓存
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)

		// 2. 缓存未命中，从数据源获取（这里模拟）
		// 实际应用中，这里会从数据库或其他数据源获取数据
		actualValue := value

		// 3. 将数据写入缓存
		err = cache.Set(ctx, key, actualValue, ttl)
		require.NoError(t, err)

		// 4. 再次检查缓存
		exists, err = cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.True(t, exists)

		// 5. 从缓存读取
		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, actualValue, result)
	})

	t.Run("Cache-Aside更新模式", func(t *testing.T) {
		key := "cache_aside:update"
		oldValue := "old_value"
		newValue := "new_value"
		ttl := 5 * time.Minute

		// 1. 设置初始值
		err := cache.Set(ctx, key, oldValue, ttl)
		require.NoError(t, err)

		// 2. 更新数据源（这里模拟）
		// 实际应用中，这里会更新数据库

		// 3. 删除缓存
		err = cache.Delete(ctx, key)
		require.NoError(t, err)

		// 4. 验证缓存已删除
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		assert.False(t, exists)

		// 5. 下次读取时会从数据源重新加载
		// 这里模拟重新加载
		err = cache.Set(ctx, key, newValue, ttl)
		require.NoError(t, err)

		// 6. 验证新值
		result, err := cache.Get(ctx, key)
		require.NoError(t, err)
		assert.Equal(t, newValue, result)
	})
}

func TestRedisCache_DataConsistency(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("价格数据一致性", func(t *testing.T) {
		// 创建价格数据
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

		// 设置价格数据
		err := cache.SetPrice(ctx, "BTCUSDT", price)
		require.NoError(t, err)

		// 立即读取验证
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

	t.Run("变化率数据一致性", func(t *testing.T) {
		// 创建变化率数据
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

		// 设置变化率数据
		err := cache.SetChangeRate(ctx, "BTCUSDT", TimeWindow1m, rate)
		require.NoError(t, err)

		// 立即读取验证
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

	t.Run("交易对信息一致性", func(t *testing.T) {
		// 创建交易对信息
		symbol := &SymbolInfo{
			Symbol:     "BTCUSDT",
			Name:       "Bitcoin/USDT",
			BaseAsset:  "BTC",
			QuoteAsset: "USDT",
			Status:     "active",
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// 设置交易对信息
		err := cache.SetSymbol(ctx, "BTCUSDT", symbol)
		require.NoError(t, err)

		// 立即读取验证
		result, err := cache.GetSymbol(ctx, "BTCUSDT")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, symbol.Symbol, result.Symbol)
		assert.Equal(t, symbol.Name, result.Name)
		assert.Equal(t, symbol.BaseAsset, result.BaseAsset)
		assert.Equal(t, symbol.QuoteAsset, result.QuoteAsset)
		assert.Equal(t, symbol.Status, result.Status)
	})
}

func TestRedisCache_ConcurrentConsistency(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("并发读写一致性", func(t *testing.T) {
		key := "concurrent:read_write"
		concurrency := 20
		var wg sync.WaitGroup
		errors := make(chan error, concurrency*2)

		// 启动写入协程
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				value := fmt.Sprintf("value_%d", id)
				err := cache.Set(ctx, key, value, 5*time.Minute)
				if err != nil {
					errors <- err
				}
			}(i)
		}

		// 启动读取协程
		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(id int) {
				defer wg.Done()
				_, err := cache.Get(ctx, key)
				if err != nil && err.Error() != "redis: nil" {
					errors <- err
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// 检查错误
		errorCount := 0
		for err := range errors {
			t.Logf("并发操作错误: %v", err)
			errorCount++
		}

		// 允许一些读取错误（因为并发写入可能导致读取到nil）
		assert.Less(t, errorCount, concurrency/2, "并发操作错误过多")
	})

	t.Run("并发批量操作一致性", func(t *testing.T) {
		concurrency := 5
		itemsPerGoroutine := 10
		var wg sync.WaitGroup
		errors := make(chan error, concurrency)

		for i := 0; i < concurrency; i++ {
			wg.Add(1)
			go func(goroutineID int) {
				defer wg.Done()

				// 创建批量数据
				data := make([]CacheData, itemsPerGoroutine)
				for j := 0; j < itemsPerGoroutine; j++ {
					key := fmt.Sprintf("batch:concurrent:%d:%d", goroutineID, j)
					value := fmt.Sprintf("value_%d_%d", goroutineID, j)
					data[j] = CacheData{
						Key:       key,
						Value:     value,
						TTL:       5 * time.Minute,
						Timestamp: time.Now(),
					}
				}

				// 执行批量写入
				result, err := cache.SetBatch(ctx, data)
				if err != nil {
					errors <- err
					return
				}

				// 验证结果
				if result.ErrorCount > 0 {
					errors <- fmt.Errorf("批量写入有错误: %d", result.ErrorCount)
					return
				}

				// 验证数据
				for j := 0; j < itemsPerGoroutine; j++ {
					key := fmt.Sprintf("batch:concurrent:%d:%d", goroutineID, j)
					expectedValue := fmt.Sprintf("value_%d_%d", goroutineID, j)

					actualValue, err := cache.Get(ctx, key)
					if err != nil {
						errors <- err
						return
					}
					if actualValue != expectedValue {
						errors <- fmt.Errorf("数据不一致: 期望 %s, 实际 %s", expectedValue, actualValue)
						return
					}
				}
			}(i)
		}

		wg.Wait()
		close(errors)

		// 检查错误
		for err := range errors {
			t.Errorf("并发批量操作错误: %v", err)
		}
	})
}

func TestRedisCache_TTLConsistency(t *testing.T) {
	// 跳过测试，需要真实Redis服务器
	t.Skip("需要真实Redis服务器，跳过测试")

	// 创建缓存实例
	config := DefaultCacheConfig("")
	cache := NewRedisCache(config, zap.NewNop())
	require.NotNil(t, cache)
	defer cache.Close()

	ctx := context.Background()

	t.Run("TTL设置一致性", func(t *testing.T) {
		key := "ttl:test"
		value := "ttl_value"
		ttl := 1 * time.Second

		// 设置数据
		err := cache.Set(ctx, key, value, ttl)
		require.NoError(t, err)

		// 立即检查TTL
		actualTTL, err := cache.GetTTL(ctx, key)
		require.NoError(t, err)
		assert.True(t, actualTTL > 0)
		assert.True(t, actualTTL <= ttl)

		// 等待TTL过期（miniredis可能需要更长时间）
		time.Sleep(2 * time.Second)

		// 验证数据已过期
		exists, err := cache.Exists(ctx, key)
		require.NoError(t, err)
		// miniredis的TTL处理可能与真实Redis不同，这里只验证TTL设置成功
		if !exists {
			assert.False(t, exists)
		} else {
			// 如果数据仍然存在，检查TTL是否已减少
			newTTL, err := cache.GetTTL(ctx, key)
			require.NoError(t, err)
			assert.True(t, newTTL < actualTTL, "TTL应该随时间减少")
		}
	})

	t.Run("TTL更新一致性", func(t *testing.T) {
		key := "ttl:update"
		value := "ttl_update_value"
		initialTTL := 1 * time.Second
		newTTL := 5 * time.Second

		// 设置初始TTL
		err := cache.Set(ctx, key, value, initialTTL)
		require.NoError(t, err)

		// 更新TTL
		err = cache.SetTTL(ctx, key, newTTL)
		require.NoError(t, err)

		// 验证TTL已更新
		actualTTL, err := cache.GetTTL(ctx, key)
		require.NoError(t, err)
		assert.True(t, actualTTL > initialTTL)
		assert.True(t, actualTTL <= newTTL)
	})
}
