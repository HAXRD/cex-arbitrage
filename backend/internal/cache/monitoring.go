package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// CacheMonitor Redis 缓存监控
type CacheMonitor struct {
	client      *redis.Client
	logger      *zap.Logger
	hitCount    atomic.Int64
	missCount   atomic.Int64
	errorCount  atomic.Int64
	lastReset   time.Time
	mu          sync.RWMutex
}

// NewCacheMonitor 创建缓存监控实例
func NewCacheMonitor(client *redis.Client, logger *zap.Logger) *CacheMonitor {
	if logger == nil {
		logger = zap.NewNop()
	}
	return &CacheMonitor{
		client:    client,
		logger:    logger,
		lastReset: time.Now(),
	}
}

// RecordHit 记录缓存命中
func (m *CacheMonitor) RecordHit() {
	m.hitCount.Add(1)
}

// RecordMiss 记录缓存未命中
func (m *CacheMonitor) RecordMiss() {
	m.missCount.Add(1)
}

// RecordError 记录错误
func (m *CacheMonitor) RecordError() {
	m.errorCount.Add(1)
}

// GetStats 获取缓存统计信息
func (m *CacheMonitor) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	hits := m.hitCount.Load()
	misses := m.missCount.Load()
	errors := m.errorCount.Load()
	total := hits + misses

	var hitRate float64
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"hits":          hits,
		"misses":        misses,
		"errors":        errors,
		"total_ops":     total,
		"hit_rate":      hitRate,
		"miss_rate":     100 - hitRate,
		"since":         m.lastReset,
		"duration":      time.Since(m.lastReset),
	}
}

// LogStats 记录缓存统计信息到日志
func (m *CacheMonitor) LogStats() {
	stats := m.GetStats()
	
	m.logger.Info("Cache statistics",
		zap.Int64("hits", stats["hits"].(int64)),
		zap.Int64("misses", stats["misses"].(int64)),
		zap.Int64("errors", stats["errors"].(int64)),
		zap.Int64("total_ops", stats["total_ops"].(int64)),
		zap.Float64("hit_rate", stats["hit_rate"].(float64)),
		zap.Duration("duration", stats["duration"].(time.Duration)),
	)

	// 警告：命中率过低
	hitRate := stats["hit_rate"].(float64)
	if hitRate < 70 && stats["total_ops"].(int64) > 100 {
		m.logger.Warn("Low cache hit rate detected",
			zap.Float64("hit_rate", hitRate),
			zap.Int64("total_ops", stats["total_ops"].(int64)),
		)
	}
}

// ResetStats 重置统计信息
func (m *CacheMonitor) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.hitCount.Store(0)
	m.missCount.Store(0)
	m.errorCount.Store(0)
	m.lastReset = time.Now()

	m.logger.Info("Cache statistics reset")
}

// LogRedisMemoryStats 记录 Redis 内存使用情况
func (m *CacheMonitor) LogRedisMemoryStats(ctx context.Context) error {
	info, err := m.client.Info(ctx, "memory").Result()
	if err != nil {
		m.logger.Error("Failed to get Redis memory info", zap.Error(err))
		return err
	}

	// 解析内存信息（简化版本）
	m.logger.Info("Redis memory info",
		zap.String("info", info),
	)

	// 获取内存使用统计
	memoryUsage, err := m.client.Info(ctx, "memory").Result()
	if err == nil {
		m.logger.Debug("Redis memory usage",
			zap.String("usage", memoryUsage),
		)
	}

	return nil
}

// StartPeriodicMonitoring 启动定期监控
func (m *CacheMonitor) StartPeriodicMonitoring(ctx context.Context, statsInterval, memoryInterval time.Duration) {
	statsTicker := time.NewTicker(statsInterval)
	memoryTicker := time.NewTicker(memoryInterval)
	defer statsTicker.Stop()
	defer memoryTicker.Stop()

	m.logger.Info("Started periodic cache monitoring",
		zap.Duration("stats_interval", statsInterval),
		zap.Duration("memory_interval", memoryInterval),
	)

	for {
		select {
		case <-ctx.Done():
			m.logger.Info("Stopping periodic cache monitoring")
			return
		case <-statsTicker.C:
			m.LogStats()
		case <-memoryTicker.C:
			if err := m.LogRedisMemoryStats(ctx); err != nil {
				m.logger.Error("Failed to log Redis memory stats", zap.Error(err))
			}
		}
	}
}

// GetHealthStatus 获取 Redis 健康状态
func (m *CacheMonitor) GetHealthStatus(ctx context.Context) map[string]interface{} {
	// Ping Redis
	if err := m.client.Ping(ctx).Err(); err != nil {
		return map[string]interface{}{
			"healthy": false,
			"error":   err.Error(),
		}
	}

	// 获取缓存统计
	stats := m.GetStats()
	
	return map[string]interface{}{
		"healthy":   true,
		"hit_rate":  stats["hit_rate"],
		"total_ops": stats["total_ops"],
	}
}

