package data_collection

import (
	"fmt"
	"sync"
	"time"
)

// deduplicatorImpl 去重器实现
type deduplicatorImpl struct {
	config *PersistenceConfig
	cache  map[string]*PersistenceItem
	mu     sync.RWMutex
	stats  map[string]int64
}

// NewDeduplicator 创建去重器
func NewDeduplicator(config *PersistenceConfig) Deduplicator {
	return &deduplicatorImpl{
		config: config,
		cache:  make(map[string]*PersistenceItem),
		stats:  make(map[string]int64),
	}
}

// IsDuplicate 检查是否重复
func (d *deduplicatorImpl) IsDuplicate(item *PersistenceItem) bool {
	if !d.config.EnableDeduplication {
		return false
	}

	d.mu.RLock()
	defer d.mu.RUnlock()

	// 生成去重键
	key := d.generateKey(item)

	// 检查缓存中是否存在
	if cached, exists := d.cache[key]; exists {
		// 检查时间窗口
		if time.Since(cached.Timestamp) <= d.config.DeduplicationWindow {
			d.stats["duplicate_count"]++
			return true
		}
	}

	return false
}

// Add 添加项目到去重缓存
func (d *deduplicatorImpl) Add(item *PersistenceItem) {
	if !d.config.EnableDeduplication {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	// 生成去重键
	key := d.generateKey(item)

	// 添加到缓存
	d.cache[key] = item
	d.stats["add_count"]++
}

// Cleanup 清理过期项目
func (d *deduplicatorImpl) Cleanup() {
	if !d.config.EnableDeduplication {
		return
	}

	d.mu.Lock()
	defer d.mu.Unlock()

	now := time.Now()
	expiredKeys := make([]string, 0)

	// 查找过期项目
	for key, item := range d.cache {
		if now.Sub(item.Timestamp) > d.config.DeduplicationWindow {
			expiredKeys = append(expiredKeys, key)
		}
	}

	// 删除过期项目
	for _, key := range expiredKeys {
		delete(d.cache, key)
	}

	d.stats["cleanup_count"]++
	d.stats["expired_count"] += int64(len(expiredKeys))
}

// GetStats 获取统计信息
func (d *deduplicatorImpl) GetStats() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range d.stats {
		stats[k] = v
	}
	stats["cache_size"] = len(d.cache)
	stats["enabled"] = d.config.EnableDeduplication
	stats["window"] = d.config.DeduplicationWindow.String()

	return stats
}

// generateKey 生成去重键
func (d *deduplicatorImpl) generateKey(item *PersistenceItem) string {
	// 基于类型、ID和时间戳生成键
	// 这里可以根据具体需求调整键的生成策略
	return fmt.Sprintf("%s:%s:%d", item.Type, item.ID, item.Timestamp.Unix())
}
