package data_collection

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// asyncPersistenceImpl 异步持久化实现
type asyncPersistenceImpl struct {
	config *PersistenceConfig
	writer DataWriter
	logger *zap.Logger

	// 队列
	queue chan *PersistenceItem

	// 控制
	running atomic.Bool
	ctx     context.Context
	cancel  context.CancelFunc
	wg      sync.WaitGroup

	// 统计
	stats *PersistenceStats
	mu    sync.RWMutex

	// 去重器
	deduplicator Deduplicator

	// 完整性检查器
	integrityChecker IntegrityChecker

	// 重试管理器
	retryManager RetryManager
}

// NewAsyncPersistence 创建异步持久化实例
func NewAsyncPersistence(config *PersistenceConfig, writer DataWriter, logger *zap.Logger) AsyncPersistence {
	if logger == nil {
		logger = zap.NewNop()
	}

	if config == nil {
		config = DefaultPersistenceConfig()
	}

	// 创建去重器
	deduplicator := NewDeduplicator(config)

	// 创建完整性检查器
	integrityChecker := NewIntegrityChecker(config)

	// 创建重试管理器
	retryManager := NewRetryManager(config)

	return &asyncPersistenceImpl{
		config:           config,
		writer:           writer,
		logger:           logger,
		queue:            make(chan *PersistenceItem, config.QueueSize),
		stats:            &PersistenceStats{StartTime: time.Now()},
		deduplicator:     deduplicator,
		integrityChecker: integrityChecker,
		retryManager:     retryManager,
	}
}

// Start 启动异步持久化
func (a *asyncPersistenceImpl) Start(ctx context.Context) error {
	if a.running.Load() {
		return fmt.Errorf("异步持久化已经在运行")
	}

	// 创建上下文
	a.ctx, a.cancel = context.WithCancel(ctx)

	// 启动工作协程
	for i := 0; i < a.config.WorkerCount; i++ {
		a.wg.Add(1)
		go a.worker(i)
	}

	// 启动清理协程
	a.wg.Add(1)
	go a.cleanupWorker()

	// 启动统计协程
	a.wg.Add(1)
	go a.statsWorker()

	a.running.Store(true)
	a.logger.Info("异步持久化已启动", zap.Int("worker_count", a.config.WorkerCount))

	return nil
}

// Stop 停止异步持久化
func (a *asyncPersistenceImpl) Stop(ctx context.Context) error {
	if !a.running.Load() {
		return fmt.Errorf("异步持久化未运行")
	}

	// 取消上下文
	a.cancel()

	// 等待所有协程完成
	done := make(chan struct{})
	go func() {
		a.wg.Wait()
		close(done)
	}()

	// 等待完成或超时
	select {
	case <-done:
		a.logger.Info("异步持久化已停止")
	case <-ctx.Done():
		a.logger.Warn("异步持久化停止超时")
		return ctx.Err()
	}

	a.running.Store(false)
	return nil
}

// IsRunning 检查是否正在运行
func (a *asyncPersistenceImpl) IsRunning() bool {
	return a.running.Load()
}

// Submit 提交单个数据
func (a *asyncPersistenceImpl) Submit(item *PersistenceItem) error {
	if !a.running.Load() {
		return fmt.Errorf("异步持久化未运行")
	}

	// 设置创建时间
	if item.CreatedAt.IsZero() {
		item.CreatedAt = time.Now()
	}

	// 去重检查
	if a.config.EnableDeduplication && a.deduplicator.IsDuplicate(item) {
		a.updateDeduplicationCount()
		return nil // 重复数据，直接返回
	}

	// 添加到去重缓存
	if a.config.EnableDeduplication {
		a.deduplicator.Add(item)
	}

	// 提交到队列
	select {
	case a.queue <- item:
		a.updateQueueStats(1)
		return nil
	default:
		a.updateQueueFullCount()
		return fmt.Errorf("队列已满，无法提交数据")
	}
}

// SubmitBatch 批量提交数据
func (a *asyncPersistenceImpl) SubmitBatch(items []*PersistenceItem) error {
	if !a.running.Load() {
		return fmt.Errorf("异步持久化未运行")
	}

	if len(items) == 0 {
		return nil
	}

	// 处理每个项目
	for _, item := range items {
		// 设置创建时间
		if item.CreatedAt.IsZero() {
			item.CreatedAt = time.Now()
		}

		// 去重检查
		if a.config.EnableDeduplication && a.deduplicator.IsDuplicate(item) {
			a.updateDeduplicationCount()
			continue
		}

		// 添加到去重缓存
		if a.config.EnableDeduplication {
			a.deduplicator.Add(item)
		}

		// 提交到队列
		select {
		case a.queue <- item:
			a.updateQueueStats(1)
		default:
			a.updateQueueFullCount()
			return fmt.Errorf("队列已满，无法提交数据")
		}
	}

	return nil
}

// GetQueueSize 获取队列大小
func (a *asyncPersistenceImpl) GetQueueSize() int {
	return len(a.queue)
}

// GetQueueCapacity 获取队列容量
func (a *asyncPersistenceImpl) GetQueueCapacity() int {
	return cap(a.queue)
}

// Flush 强制刷新队列
func (a *asyncPersistenceImpl) Flush() error {
	if !a.running.Load() {
		return fmt.Errorf("异步持久化未运行")
	}

	// 等待队列清空
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("刷新超时")
		case <-ticker.C:
			if len(a.queue) == 0 {
				return nil
			}
		}
	}
}

// GetStats 获取统计信息
func (a *asyncPersistenceImpl) GetStats() *PersistenceStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 复制统计信息
	stats := *a.stats
	stats.LastUpdate = time.Now()
	stats.QueueSize = int64(len(a.queue))
	stats.QueueCapacity = int64(cap(a.queue))

	return &stats
}

// ResetStats 重置统计信息
func (a *asyncPersistenceImpl) ResetStats() {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stats = &PersistenceStats{StartTime: time.Now()}
}

// HealthCheck 健康检查
func (a *asyncPersistenceImpl) HealthCheck(ctx context.Context) error {
	if !a.running.Load() {
		return fmt.Errorf("异步持久化未运行")
	}

	// 检查写入器健康状态
	if err := a.writer.HealthCheck(ctx); err != nil {
		return fmt.Errorf("写入器健康检查失败: %w", err)
	}

	// 检查队列状态
	queueSize := len(a.queue)
	queueCapacity := cap(a.queue)
	if float64(queueSize)/float64(queueCapacity) > 0.9 {
		return fmt.Errorf("队列使用率过高: %d/%d", queueSize, queueCapacity)
	}

	return nil
}

// worker 工作协程
func (a *asyncPersistenceImpl) worker(id int) {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.BatchTimeout)
	defer ticker.Stop()

	var batch []*PersistenceItem

	for {
		select {
		case <-a.ctx.Done():
			// 处理剩余批次
			if len(batch) > 0 {
				a.processBatch(batch)
			}
			return

		case item := <-a.queue:
			batch = append(batch, item)

			// 达到批量大小，立即处理
			if len(batch) >= a.config.BatchSize {
				a.processBatch(batch)
				batch = batch[:0]
			}

		case <-ticker.C:
			// 超时处理批次
			if len(batch) > 0 {
				a.processBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// processBatch 处理批次
func (a *asyncPersistenceImpl) processBatch(batch []*PersistenceItem) {
	if len(batch) == 0 {
		return
	}

	start := time.Now()

	// 完整性检查
	if a.config.EnableIntegrityCheck {
		if err := a.integrityChecker.CheckIntegrity(batch); err != nil {
			a.logger.Error("数据完整性检查失败", zap.Error(err))
			a.updateErrorCount()
			return
		}
	}

	// 批量写入
	result, err := a.writer.WriteBatch(a.ctx, batch)
	if err != nil {
		a.logger.Error("批量写入失败", zap.Error(err))
		a.updateErrorCount()

		// 处理重试
		a.handleRetry(batch, err)
		return
	}

	// 更新统计
	duration := time.Since(start)
	a.updateProcessStats(len(batch), duration, result)

	a.logger.Debug("批次处理完成",
		zap.Int("batch_size", len(batch)),
		zap.Int("success_count", result.SuccessCount),
		zap.Int("error_count", result.ErrorCount),
		zap.Duration("duration", duration))
}

// handleRetry 处理重试
func (a *asyncPersistenceImpl) handleRetry(batch []*PersistenceItem, err error) {
	for _, item := range batch {
		if a.retryManager.ShouldRetry(item, err) {
			item.RetryCount++
			delay := a.retryManager.CalculateDelay(item.RetryCount)

			// 异步重试
			go func(item *PersistenceItem) {
				time.Sleep(delay)
				select {
				case a.queue <- item:
					a.retryManager.RecordRetry(item, err)
				default:
					a.logger.Warn("重试时队列已满", zap.String("item_id", item.ID))
				}
			}(item)
		}
	}
}

// cleanupWorker 清理工作协程
func (a *asyncPersistenceImpl) cleanupWorker() {
	defer a.wg.Done()

	ticker := time.NewTicker(a.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			// 清理去重缓存
			if a.deduplicator != nil {
				a.deduplicator.Cleanup()
			}

			// 清理统计信息
			a.cleanupStats()
		}
	}
}

// statsWorker 统计工作协程
func (a *asyncPersistenceImpl) statsWorker() {
	defer a.wg.Done()

	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-a.ctx.Done():
			return
		case <-ticker.C:
			a.updateStats()
		}
	}
}

// 统计更新方法
func (a *asyncPersistenceImpl) updateQueueStats(count int) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats.TotalProcessed += int64(count)
}

func (a *asyncPersistenceImpl) updateQueueFullCount() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats.QueueFullCount++
}

func (a *asyncPersistenceImpl) updateDeduplicationCount() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats.DeduplicationCount++
}

func (a *asyncPersistenceImpl) updateErrorCount() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.stats.ErrorCount++
}

func (a *asyncPersistenceImpl) updateProcessStats(batchSize int, duration time.Duration, result *PersistenceResult) {
	a.mu.Lock()
	defer a.mu.Unlock()

	a.stats.SuccessCount += int64(result.SuccessCount)
	a.stats.ErrorCount += int64(result.ErrorCount)
	a.stats.BatchCount++

	// 更新性能统计
	if a.stats.AvgProcessTime == 0 {
		a.stats.AvgProcessTime = duration
	} else {
		a.stats.AvgProcessTime = (a.stats.AvgProcessTime + duration) / 2
	}

	if duration > a.stats.MaxProcessTime {
		a.stats.MaxProcessTime = duration
	}

	if a.stats.MinProcessTime == 0 || duration < a.stats.MinProcessTime {
		a.stats.MinProcessTime = duration
	}

	// 更新批量统计
	if a.stats.AvgBatchSize == 0 {
		a.stats.AvgBatchSize = float64(batchSize)
	} else {
		a.stats.AvgBatchSize = (a.stats.AvgBatchSize + float64(batchSize)) / 2
	}

	if int64(batchSize) > a.stats.MaxBatchSize {
		a.stats.MaxBatchSize = int64(batchSize)
	}
}

func (a *asyncPersistenceImpl) updateStats() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 更新内存使用（简化计算）
	a.stats.MemoryUsage = int64(len(a.queue)) * 1024 // 假设每个项目1KB
	if a.stats.MemoryUsage > a.stats.MaxMemoryUsage {
		a.stats.MaxMemoryUsage = a.stats.MemoryUsage
	}
}

func (a *asyncPersistenceImpl) cleanupStats() {
	a.mu.Lock()
	defer a.mu.Unlock()

	// 清理过期的统计信息
	// 这里可以添加更复杂的清理逻辑
}
