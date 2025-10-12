// +build integration

package cache

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"github.com/redis/go-redis/v9"
)

// setupCacheMonitoringIntegrationTest 设置缓存监控集成测试
func setupCacheMonitoringIntegrationTest(t *testing.T) (*redis.Client, *zap.Logger) {
	cfg := &config.RedisConfig{
		Host:         "localhost",
		Port:         6379,
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 5,
		MaxRetries:   3,
		PoolTimeout:  5,
		IdleTimeout:  300,
	}

	logger := zap.NewExample()
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
		PoolSize: cfg.PoolSize,
	})
	
	// 测试连接
	ctx := context.Background()
	err := client.Ping(ctx).Err()
	require.NoError(t, err, "Failed to connect to Redis")

	return client, logger
}

// TestCacheMonitor_Integration_BasicOperations 集成测试：基本操作
func TestCacheMonitor_Integration_BasicOperations(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)

	// 测试基本统计记录
	monitor.RecordHit()
	monitor.RecordHit()
	monitor.RecordMiss()
	monitor.RecordError()

	stats := monitor.GetStats()
	assert.Equal(t, int64(2), stats["hits"])
	assert.Equal(t, int64(1), stats["misses"])
	assert.Equal(t, int64(1), stats["errors"])
	assert.Equal(t, int64(3), stats["total_ops"])

	hitRate := stats["hit_rate"].(float64)
	assert.InDelta(t, 66.67, hitRate, 0.1)

	t.Logf("Cache stats: %+v", stats)
}

// TestCacheMonitor_Integration_RedisMemoryStats 集成测试：Redis 内存统计
func TestCacheMonitor_Integration_RedisMemoryStats(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)
	ctx := context.Background()

	// 测试 Redis 内存统计
	err := monitor.LogRedisMemoryStats(ctx)
	if err != nil {
		t.Logf("Redis memory stats failed (expected in some environments): %v", err)
	}

	// 验证健康状态
	status := monitor.GetHealthStatus(ctx)
	assert.NotNil(t, status)
	assert.True(t, status["healthy"].(bool))
	assert.Contains(t, status, "hit_rate")
	assert.Contains(t, status, "total_ops")

	t.Logf("Cache health status: %+v", status)
}

// TestCacheMonitor_Integration_PeriodicMonitoring 集成测试：定期监控
func TestCacheMonitor_Integration_PeriodicMonitoring(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动定期监控
	go monitor.StartPeriodicMonitoring(ctx, 1*time.Second, 2*time.Second)

	// 在监控运行期间记录一些操作
	for i := 0; i < 10; i++ {
		monitor.RecordHit()
		if i%3 == 0 {
			monitor.RecordMiss()
		}
		time.Sleep(200 * time.Millisecond)
	}

	// 等待监控结束
	cancel()
	time.Sleep(100 * time.Millisecond)

	t.Log("Periodic monitoring test completed")
}

// TestCacheMonitor_Integration_HighLoad 集成测试：高负载
func TestCacheMonitor_Integration_HighLoad(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)
	ctx := context.Background()

	// 模拟高负载场景
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// 每个 goroutine 执行 100 次操作
			for j := 0; j < 100; j++ {
				if j%3 == 0 {
					monitor.RecordMiss()
				} else {
					monitor.RecordHit()
				}
				
				// 模拟一些 Redis 操作
				key := fmt.Sprintf("test:key:%d:%d", id, j)
				client.Set(ctx, key, "value", time.Minute)
			}
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 记录统计信息
	monitor.LogStats()

	// 获取最终统计
	stats := monitor.GetStats()
	totalOps := stats["total_ops"].(int64)
	hitRate := stats["hit_rate"].(float64)

	t.Logf("High load test completed - Total ops: %d, Hit rate: %.2f%%", totalOps, hitRate)
	assert.True(t, totalOps > 0)
}

// TestCacheMonitor_Integration_ConcurrentAccess 集成测试：并发访问
func TestCacheMonitor_Integration_ConcurrentAccess(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)

	// 并发执行监控操作
	done := make(chan bool, 5)
	
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// 并发记录操作
			for j := 0; j < 20; j++ {
				if j%2 == 0 {
					monitor.RecordHit()
				} else {
					monitor.RecordMiss()
				}
			}
			
			// 并发获取统计
			stats := monitor.GetStats()
			assert.NotNil(t, stats)
			
			t.Logf("Goroutine %d completed", id)
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 5; i++ {
		<-done
	}

	// 验证最终统计
	stats := monitor.GetStats()
	totalOps := stats["total_ops"].(int64)
	assert.True(t, totalOps >= 100) // 5 goroutines * 20 operations each

	t.Log("Concurrent access test completed")
}

// TestCacheMonitor_Integration_ResetStats 集成测试：统计重置
func TestCacheMonitor_Integration_ResetStats(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)

	// 记录一些操作
	for i := 0; i < 50; i++ {
		monitor.RecordHit()
		if i%5 == 0 {
			monitor.RecordMiss()
		}
	}

	// 验证统计
	stats := monitor.GetStats()
	assert.True(t, stats["total_ops"].(int64) > 0)

	// 重置统计
	monitor.ResetStats()

	// 验证重置后的统计
	stats = monitor.GetStats()
	assert.Equal(t, int64(0), stats["hits"])
	assert.Equal(t, int64(0), stats["misses"])
	assert.Equal(t, int64(0), stats["errors"])
	assert.Equal(t, int64(0), stats["total_ops"])

	t.Log("Reset stats test completed")
}

// TestCacheMonitor_Integration_HealthCheck 集成测试：健康检查
func TestCacheMonitor_Integration_HealthCheck(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)
	ctx := context.Background()

	// 测试健康检查
	status := monitor.GetHealthStatus(ctx)
	
	assert.NotNil(t, status)
	assert.True(t, status["healthy"].(bool))
	assert.Contains(t, status, "hit_rate")
	assert.Contains(t, status, "total_ops")

	t.Logf("Health check passed: %+v", status)
}

// TestCacheMonitor_Integration_ErrorHandling 集成测试：错误处理
func TestCacheMonitor_Integration_ErrorHandling(t *testing.T) {
	// 使用无效配置测试错误处理
	cfg := &config.RedisConfig{
		Host:     "invalid-host",
		Port:     9999,
		Password: "invalid-password",
		DB:       999,
	}

	logger := zap.NewExample()
	client := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})
	
	// 测试连接
	ctx := context.Background()
	err := client.Ping(ctx).Err()
	if err != nil {
		t.Logf("Expected connection failure: %v", err)
		return
	}
	
	// 如果连接成功，测试监控
	monitor := NewCacheMonitor(client, logger)
	
	// 健康检查应该返回不健康状态
	status := monitor.GetHealthStatus(ctx)
	assert.NotNil(t, status)
	
	// 清理
	if client != nil {
		client.Close()
	}
}

// TestCacheMonitor_Integration_LongRunning 集成测试：长时间运行
func TestCacheMonitor_Integration_LongRunning(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
	defer cancel()

	// 启动长时间运行的监控
	go monitor.StartPeriodicMonitoring(ctx, 1*time.Second, 3*time.Second)

	// 在监控运行期间执行一些操作
	start := time.Now()
	for time.Since(start) < 5*time.Second {
		monitor.RecordHit()
		monitor.RecordMiss()
		
		// 执行一些 Redis 操作
		key := fmt.Sprintf("test:long:%d", time.Now().Unix())
		client.Set(ctx, key, "value", time.Minute)
		
		time.Sleep(500 * time.Millisecond)
	}

	// 等待监控结束
	cancel()
	time.Sleep(100 * time.Millisecond)

	// 获取最终统计
	stats := monitor.GetStats()
	t.Logf("Long running test completed - Final stats: %+v", stats)
}

// TestCacheMonitor_Integration_RealWorldScenario 集成测试：真实场景
func TestCacheMonitor_Integration_RealWorldScenario(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 启动定期监控
	go monitor.StartPeriodicMonitoring(ctx, 2*time.Second, 4*time.Second)

	// 模拟真实应用场景
	// 1. 正常缓存操作
	for i := 0; i < 20; i++ {
		key := fmt.Sprintf("price:BTC-USDT:%d", i)
		client.Set(ctx, key, "50000", time.Minute)
		monitor.RecordHit()
		time.Sleep(100 * time.Millisecond)
	}

	// 2. 高负载操作
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 50; j++ {
				key := fmt.Sprintf("ticker:ETH-USDT:%d", j)
				client.Set(ctx, key, "3000", time.Minute)
				monitor.RecordHit()
			}
		}()
	}

	// 等待高负载操作完成
	for i := 0; i < 3; i++ {
		<-done
	}

	// 3. 模拟缓存未命中
	for i := 0; i < 10; i++ {
		key := fmt.Sprintf("missing:key:%d", i)
		client.Get(ctx, key)
		monitor.RecordMiss()
	}

	// 4. 最终健康检查
	status := monitor.GetHealthStatus(ctx)
	assert.True(t, status["healthy"].(bool))
	
	hitRate := status["hit_rate"].(float64)
	t.Logf("Final hit rate: %.2f%%", hitRate)

	// 等待监控结束
	cancel()
	time.Sleep(100 * time.Millisecond)

	t.Log("Real world scenario test completed")
}

// TestCacheMonitor_Integration_MemoryUsage 集成测试：内存使用
func TestCacheMonitor_Integration_MemoryUsage(t *testing.T) {
	client, logger := setupCacheMonitoringIntegrationTest(t)
	defer client.Close()

	monitor := NewCacheMonitor(client, logger)

	// 执行大量操作，观察内存使用
	for i := 0; i < 1000; i++ {
		monitor.RecordHit()
		if i%10 == 0 {
			monitor.RecordMiss()
		}
		
		if i%100 == 0 {
			stats := monitor.GetStats()
			t.Logf("Completed %d operations, stats: %+v", i, stats)
		}
	}

	// 最终统计
	stats := monitor.GetStats()
	totalOps := stats["total_ops"].(int64)
	hitRate := stats["hit_rate"].(float64)

	t.Logf("Memory usage test completed - Total ops: %d, Hit rate: %.2f%%", totalOps, hitRate)
	assert.True(t, totalOps >= 1000)
}
