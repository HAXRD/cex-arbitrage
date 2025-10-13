package data_collection

import (
	"context"
	"time"
)

// GoroutinePool 协程池接口
type GoroutinePool interface {
	// Start 启动协程池
	Start(ctx context.Context) error

	// Stop 停止协程池
	Stop(ctx context.Context) error

	// Submit 提交任务
	Submit(task Task) error

	// SubmitBatch 批量提交任务
	SubmitBatch(tasks []Task) error

	// GetStatus 获取协程池状态
	GetStatus() PoolStatus

	// GetWorkerCount 获取工作协程数量
	GetWorkerCount() int

	// GetQueueSize 获取队列大小
	GetQueueSize() int

	// SetMaxWorkers 设置最大工作协程数
	SetMaxWorkers(count int)

	// SetQueueSize 设置队列大小
	SetQueueSize(size int)
}

// Task 任务接口
type Task interface {
	// Execute 执行任务
	Execute(ctx context.Context) error

	// GetID 获取任务ID
	GetID() string

	// GetPriority 获取任务优先级
	GetPriority() int

	// GetRetryCount 获取重试次数
	GetRetryCount() int

	// SetRetryCount 设置重试次数
	SetRetryCount(count int)
}

// PoolStatus 协程池状态
type PoolStatus struct {
	IsRunning      bool          `json:"is_running"`
	WorkerCount    int           `json:"worker_count"`
	MaxWorkers     int           `json:"max_workers"`
	QueueSize      int           `json:"queue_size"`
	MaxQueueSize   int           `json:"max_queue_size"`
	ProcessedTasks int64         `json:"processed_tasks"`
	FailedTasks    int64         `json:"failed_tasks"`
	Uptime         time.Duration `json:"uptime"`
	LastError      string        `json:"last_error,omitempty"`
}

// PoolConfig 协程池配置
type PoolConfig struct {
	MaxWorkers    int           `json:"max_workers" yaml:"max_workers"`
	QueueSize     int           `json:"queue_size" yaml:"queue_size"`
	WorkerTimeout time.Duration `json:"worker_timeout" yaml:"worker_timeout"`
	TaskTimeout   time.Duration `json:"task_timeout" yaml:"task_timeout"`
	RetryCount    int           `json:"retry_count" yaml:"retry_count"`
	RetryDelay    time.Duration `json:"retry_delay" yaml:"retry_delay"`
}

// DefaultPoolConfig 默认协程池配置
func DefaultPoolConfig() *PoolConfig {
	return &PoolConfig{
		MaxWorkers:    10,
		QueueSize:     1000,
		WorkerTimeout: 30 * time.Second,
		TaskTimeout:   10 * time.Second,
		RetryCount:    3,
		RetryDelay:    1 * time.Second,
	}
}

// DataCollectionTask 数据采集任务
type DataCollectionTask struct {
	ID           string    `json:"id"`
	Symbol       string    `json:"symbol"`
	Channel      string    `json:"channel"`
	Priority     int       `json:"priority"`
	RetryCount   int       `json:"retry_count"`
	CreatedAt    time.Time `json:"created_at"`
	LastAttempt  time.Time `json:"last_attempt,omitempty"`
	ErrorMessage string    `json:"error_message,omitempty"`
}

// Execute 执行数据采集任务
func (t *DataCollectionTask) Execute(ctx context.Context) error {
	// 这里将实现具体的数据采集逻辑
	// 暂时返回nil，实际实现时会调用WebSocket客户端进行数据采集
	return nil
}

// GetID 获取任务ID
func (t *DataCollectionTask) GetID() string {
	return t.ID
}

// GetPriority 获取任务优先级
func (t *DataCollectionTask) GetPriority() int {
	return t.Priority
}

// GetRetryCount 获取重试次数
func (t *DataCollectionTask) GetRetryCount() int {
	return t.RetryCount
}

// SetRetryCount 设置重试次数
func (t *DataCollectionTask) SetRetryCount(count int) {
	t.RetryCount = count
}
