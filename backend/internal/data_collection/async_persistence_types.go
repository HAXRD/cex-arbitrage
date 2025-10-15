package data_collection

import (
	"context"
	"time"
)

// PersistenceConfig 持久化配置
type PersistenceConfig struct {
	// 队列配置
	QueueSize     int           `json:"queue_size" yaml:"queue_size"`
	BatchSize     int           `json:"batch_size" yaml:"batch_size"`
	BatchTimeout  time.Duration `json:"batch_timeout" yaml:"batch_timeout"`
	FlushInterval time.Duration `json:"flush_interval" yaml:"flush_interval"`

	// 重试配置
	MaxRetries    int           `json:"max_retries" yaml:"max_retries"`
	RetryInterval time.Duration `json:"retry_interval" yaml:"retry_interval"`
	RetryBackoff  float64       `json:"retry_backoff" yaml:"retry_backoff"`
	MaxRetryDelay time.Duration `json:"max_retry_delay" yaml:"max_retry_delay"`

	// 去重配置
	EnableDeduplication bool          `json:"enable_deduplication" yaml:"enable_deduplication"`
	DeduplicationWindow time.Duration `json:"deduplication_window" yaml:"deduplication_window"`

	// 数据完整性配置
	EnableIntegrityCheck bool          `json:"enable_integrity_check" yaml:"enable_integrity_check"`
	IntegrityCheckWindow time.Duration `json:"integrity_check_window" yaml:"integrity_check_window"`

	// 性能配置
	WorkerCount     int           `json:"worker_count" yaml:"worker_count"`
	MaxMemoryUsage  int64         `json:"max_memory_usage" yaml:"max_memory_usage"` // bytes
	CleanupInterval time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// PersistenceItem 持久化项目
type PersistenceItem struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"` // price, changerate, symbol, etc.
	Data       interface{}            `json:"data"`
	Timestamp  time.Time              `json:"timestamp"`
	Priority   int                    `json:"priority"` // 优先级，数字越小优先级越高
	RetryCount int                    `json:"retry_count"`
	Metadata   map[string]interface{} `json:"metadata"`
	CreatedAt  time.Time              `json:"created_at"`
}

// PersistenceBatch 持久化批次
type PersistenceBatch struct {
	Items     []*PersistenceItem `json:"items"`
	Timestamp time.Time          `json:"timestamp"`
	Size      int                `json:"size"`
	Priority  int                `json:"priority"`
}

// PersistenceResult 持久化结果
type PersistenceResult struct {
	SuccessCount int                `json:"success_count"`
	ErrorCount   int                `json:"error_count"`
	Duration     time.Duration      `json:"duration"`
	Errors       []PersistenceError `json:"errors"`
	Timestamp    time.Time          `json:"timestamp"`
}

// PersistenceError 持久化错误
type PersistenceError struct {
	ItemID    string    `json:"item_id"`
	Error     string    `json:"error"`
	Timestamp time.Time `json:"timestamp"`
	Retryable bool      `json:"retryable"`
}

// PersistenceStats 持久化统计
type PersistenceStats struct {
	// 队列统计
	QueueSize      int64 `json:"queue_size"`
	QueueCapacity  int64 `json:"queue_capacity"`
	QueueFullCount int64 `json:"queue_full_count"`

	// 处理统计
	TotalProcessed int64 `json:"total_processed"`
	SuccessCount   int64 `json:"success_count"`
	ErrorCount     int64 `json:"error_count"`
	RetryCount     int64 `json:"retry_count"`

	// 性能统计
	AvgProcessTime time.Duration `json:"avg_process_time"`
	MaxProcessTime time.Duration `json:"max_process_time"`
	MinProcessTime time.Duration `json:"min_process_time"`

	// 批量统计
	BatchCount   int64   `json:"batch_count"`
	AvgBatchSize float64 `json:"avg_batch_size"`
	MaxBatchSize int64   `json:"max_batch_size"`

	// 去重统计
	DeduplicationCount int64 `json:"deduplication_count"`

	// 内存统计
	MemoryUsage    int64 `json:"memory_usage"`
	MaxMemoryUsage int64 `json:"max_memory_usage"`

	// 时间戳
	LastUpdate time.Time `json:"last_update"`
	StartTime  time.Time `json:"start_time"`
}

// AsyncPersistence 异步持久化接口
type AsyncPersistence interface {
	// 基础操作
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 数据提交
	Submit(item *PersistenceItem) error
	SubmitBatch(items []*PersistenceItem) error

	// 队列管理
	GetQueueSize() int
	GetQueueCapacity() int
	Flush() error

	// 统计信息
	GetStats() *PersistenceStats
	ResetStats()

	// 健康检查
	HealthCheck(ctx context.Context) error
}

// DataWriter 数据写入器接口
type DataWriter interface {
	// 写入单个数据
	Write(ctx context.Context, item *PersistenceItem) error

	// 批量写入
	WriteBatch(ctx context.Context, items []*PersistenceItem) (*PersistenceResult, error)

	// 健康检查
	HealthCheck(ctx context.Context) error

	// 关闭连接
	Close() error
}

// Deduplicator 去重器接口
type Deduplicator interface {
	// 检查是否重复
	IsDuplicate(item *PersistenceItem) bool

	// 添加项目到去重缓存
	Add(item *PersistenceItem)

	// 清理过期项目
	Cleanup()

	// 获取统计信息
	GetStats() map[string]interface{}
}

// IntegrityChecker 完整性检查器接口
type IntegrityChecker interface {
	// 检查数据完整性
	CheckIntegrity(items []*PersistenceItem) error

	// 验证数据格式
	ValidateFormat(item *PersistenceItem) error

	// 检查数据一致性
	CheckConsistency(items []*PersistenceItem) error
}

// RetryManager 重试管理器接口
type RetryManager interface {
	// 检查是否应该重试
	ShouldRetry(item *PersistenceItem, err error) bool

	// 计算重试延迟
	CalculateDelay(retryCount int) time.Duration

	// 记录重试
	RecordRetry(item *PersistenceItem, err error)

	// 获取重试统计
	GetRetryStats() map[string]interface{}
}
