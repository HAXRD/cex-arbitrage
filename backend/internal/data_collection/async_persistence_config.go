package data_collection

import (
	"time"
)

// DefaultPersistenceConfig 创建默认持久化配置
func DefaultPersistenceConfig() *PersistenceConfig {
	return &PersistenceConfig{
		// 队列配置
		QueueSize:     1000,
		BatchSize:     100,
		BatchTimeout:  1 * time.Second,
		FlushInterval: 5 * time.Second,

		// 重试配置
		MaxRetries:    3,
		RetryInterval: 1 * time.Second,
		RetryBackoff:  2.0,
		MaxRetryDelay: 30 * time.Second,

		// 去重配置
		EnableDeduplication: true,
		DeduplicationWindow: 5 * time.Minute,

		// 数据完整性配置
		EnableIntegrityCheck: true,
		IntegrityCheckWindow: 1 * time.Minute,

		// 性能配置
		WorkerCount:     5,
		MaxMemoryUsage:  100 * 1024 * 1024, // 100MB
		CleanupInterval: 1 * time.Minute,
	}
}

// NewPersistenceConfig 创建自定义持久化配置
func NewPersistenceConfig(queueSize, batchSize, workerCount int) *PersistenceConfig {
	config := DefaultPersistenceConfig()
	config.QueueSize = queueSize
	config.BatchSize = batchSize
	config.WorkerCount = workerCount
	return config
}

// NewPersistenceConfigWithRetry 创建带重试配置的持久化配置
func NewPersistenceConfigWithRetry(queueSize, batchSize, maxRetries int, retryInterval time.Duration) *PersistenceConfig {
	config := DefaultPersistenceConfig()
	config.QueueSize = queueSize
	config.BatchSize = batchSize
	config.MaxRetries = maxRetries
	config.RetryInterval = retryInterval
	return config
}
