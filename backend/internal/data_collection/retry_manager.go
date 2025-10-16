package data_collection

import (
	"fmt"
	"math"
	"sync"
	"time"
)

// retryManagerImpl 重试管理器实现
type retryManagerImpl struct {
	config *PersistenceConfig
	stats  map[string]int64
	mu     sync.RWMutex
}

// NewRetryManager 创建重试管理器
func NewRetryManager(config *PersistenceConfig) RetryManager {
	return &retryManagerImpl{
		config: config,
		stats:  make(map[string]int64),
	}
}

// ShouldRetry 检查是否应该重试
func (r *retryManagerImpl) ShouldRetry(item *PersistenceItem, err error) bool {
	// 检查重试次数
	if item.RetryCount >= r.config.MaxRetries {
		return false
	}

	// 检查错误类型（这里简化处理，实际可以根据错误类型判断）
	// 网络错误、超时错误等可以重试
	// 数据格式错误、权限错误等不应该重试
	errorStr := err.Error()
	retryableErrors := []string{
		"timeout",
		"connection",
		"network",
		"temporary",
		"busy",
		"unavailable",
	}

	for _, retryableError := range retryableErrors {
		if contains(errorStr, retryableError) {
			return true
		}
	}

	// 默认不重试
	return false
}

// CalculateDelay 计算重试延迟
func (r *retryManagerImpl) CalculateDelay(retryCount int) time.Duration {
	// 指数退避算法
	delay := float64(r.config.RetryInterval) * math.Pow(r.config.RetryBackoff, float64(retryCount))

	// 限制最大延迟
	if delay > float64(r.config.MaxRetryDelay) {
		delay = float64(r.config.MaxRetryDelay)
	}

	// 添加随机抖动（±10%）
	jitter := delay * 0.1 * (2*math.Mod(float64(time.Now().UnixNano()), 1) - 1)
	delay += jitter

	// 确保延迟为正数
	if delay < 0 {
		delay = float64(r.config.RetryInterval)
	}

	return time.Duration(delay)
}

// RecordRetry 记录重试
func (r *retryManagerImpl) RecordRetry(item *PersistenceItem, err error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.stats["retry_count"]++
	r.stats[fmt.Sprintf("retry_type_%s", item.Type)]++
	r.stats[fmt.Sprintf("retry_error_%s", err.Error())]++
}

// GetRetryStats 获取重试统计
func (r *retryManagerImpl) GetRetryStats() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	stats := make(map[string]interface{})
	for k, v := range r.stats {
		stats[k] = v
	}

	stats["max_retries"] = r.config.MaxRetries
	stats["retry_interval"] = r.config.RetryInterval.String()
	stats["retry_backoff"] = r.config.RetryBackoff
	stats["max_retry_delay"] = r.config.MaxRetryDelay.String()

	return stats
}

// contains 检查字符串是否包含子字符串（忽略大小写）
func contains(s, substr string) bool {
	return len(s) >= len(substr) &&
		(s == substr ||
			(len(s) > len(substr) &&
				(s[:len(substr)] == substr ||
					s[len(s)-len(substr):] == substr ||
					indexOf(s, substr) >= 0)))
}

// indexOf 查找子字符串在字符串中的位置
func indexOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
