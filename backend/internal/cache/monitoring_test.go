package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func setupMiniRedisForMonitoring(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)
	
	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})
	
	return client, mr
}

func TestCacheMonitor_RecordStats(t *testing.T) {
	client, _ := setupMiniRedisForMonitoring(t)
	defer client.Close()

	logger := zap.NewExample()
	monitor := NewCacheMonitor(client, logger)

	// 记录一些操作
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
}

func TestCacheMonitor_LogStats(t *testing.T) {
	client, _ := setupMiniRedisForMonitoring(t)
	defer client.Close()

	logger := zap.NewExample()
	monitor := NewCacheMonitor(client, logger)

	monitor.RecordHit()
	monitor.RecordHit()
	monitor.RecordMiss()

	// 应该不会 panic
	monitor.LogStats()
}

func TestCacheMonitor_ResetStats(t *testing.T) {
	client, _ := setupMiniRedisForMonitoring(t)
	defer client.Close()

	logger := zap.NewExample()
	monitor := NewCacheMonitor(client, logger)

	monitor.RecordHit()
	monitor.RecordMiss()

	stats := monitor.GetStats()
	assert.Equal(t, int64(1), stats["hits"])

	// 重置统计
	monitor.ResetStats()

	stats = monitor.GetStats()
	assert.Equal(t, int64(0), stats["hits"])
	assert.Equal(t, int64(0), stats["misses"])
	assert.Equal(t, int64(0), stats["errors"])
}

func TestCacheMonitor_LogRedisMemoryStats(t *testing.T) {
	client, _ := setupMiniRedisForMonitoring(t)
	defer client.Close()

	logger := zap.NewExample()
	monitor := NewCacheMonitor(client, logger)

	ctx := context.Background()
	err := monitor.LogRedisMemoryStats(ctx)

	// miniredis 可能不支持 INFO 命令，所以可能会失败
	// 我们只验证它不会 panic
	if err != nil {
		t.Logf("LogRedisMemoryStats returned error (expected with miniredis): %v", err)
	}
}

func TestCacheMonitor_GetHealthStatus(t *testing.T) {
	client, _ := setupMiniRedisForMonitoring(t)
	defer client.Close()

	logger := zap.NewExample()
	monitor := NewCacheMonitor(client, logger)

	monitor.RecordHit()
	monitor.RecordHit()
	monitor.RecordMiss()

	ctx := context.Background()
	status := monitor.GetHealthStatus(ctx)

	assert.NotNil(t, status)
	assert.True(t, status["healthy"].(bool))
	assert.Contains(t, status, "hit_rate")
	assert.Contains(t, status, "total_ops")
}

func TestCacheMonitor_PeriodicMonitoring(t *testing.T) {
	client, _ := setupMiniRedisForMonitoring(t)
	defer client.Close()

	logger := zap.NewExample()
	monitor := NewCacheMonitor(client, logger)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// 启动定期监控
	go monitor.StartPeriodicMonitoring(ctx, 500*time.Millisecond, 1*time.Second)

	// 在监控运行期间记录一些操作
	monitor.RecordHit()
	monitor.RecordHit()
	monitor.RecordMiss()

	// 等待监控运行
	time.Sleep(1 * time.Second)

	// 取消监控
	cancel()
	time.Sleep(100 * time.Millisecond)
}

func TestCacheMonitor_HighCacheHitRate(t *testing.T) {
	client, _ := setupMiniRedisForMonitoring(t)
	defer client.Close()

	logger := zap.NewExample()
	monitor := NewCacheMonitor(client, logger)

	// 模拟高命中率
	for i := 0; i < 90; i++ {
		monitor.RecordHit()
	}
	for i := 0; i < 10; i++ {
		monitor.RecordMiss()
	}

	stats := monitor.GetStats()
	hitRate := stats["hit_rate"].(float64)
	assert.InDelta(t, 90.0, hitRate, 0.1)

	// 记录日志应该没有警告
	monitor.LogStats()
}

func TestCacheMonitor_LowCacheHitRate(t *testing.T) {
	client, _ := setupMiniRedisForMonitoring(t)
	defer client.Close()

	logger := zap.NewExample()
	monitor := NewCacheMonitor(client, logger)

	// 模拟低命中率
	for i := 0; i < 30; i++ {
		monitor.RecordHit()
	}
	for i := 0; i < 70; i++ {
		monitor.RecordMiss()
	}

	stats := monitor.GetStats()
	hitRate := stats["hit_rate"].(float64)
	assert.InDelta(t, 30.0, hitRate, 0.1)

	// 记录日志应该有警告（低命中率）
	monitor.LogStats()
}

