package data_collection

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// GoroutinePoolImpl 协程池实现
type GoroutinePoolImpl struct {
	config         *PoolConfig
	logger         *zap.Logger
	workers        []*worker
	workerCount    int32
	maxWorkers     int32
	queue          chan Task
	queueSize      int32
	maxQueueSize   int32
	processedTasks int64
	failedTasks    int64
	startTime      time.Time
	lastError      string
	mu             sync.RWMutex
	running        bool
	ctx            context.Context
	cancel         context.CancelFunc
	wg             sync.WaitGroup
}

// worker 工作协程
type worker struct {
	id       int
	pool     *GoroutinePoolImpl
	taskChan chan Task
	quit     chan bool
}

// NewGoroutinePool 创建新的协程池
func NewGoroutinePool(config *PoolConfig, logger *zap.Logger) GoroutinePool {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &GoroutinePoolImpl{
		config:       config,
		logger:       logger,
		workers:      make([]*worker, 0, config.MaxWorkers),
		maxWorkers:   int32(config.MaxWorkers),
		queue:        make(chan Task, config.QueueSize),
		maxQueueSize: int32(config.QueueSize),
		startTime:    time.Now(),
	}
}

// Start 启动协程池
func (p *GoroutinePoolImpl) Start(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.running {
		return fmt.Errorf("协程池已在运行中")
	}

	p.ctx, p.cancel = context.WithCancel(ctx)
	p.running = true
	p.startTime = time.Now()

	// 启动工作协程
	for i := 0; i < int(p.maxWorkers); i++ {
		worker := &worker{
			id:       i,
			pool:     p,
			taskChan: make(chan Task, 1),
			quit:     make(chan bool, 1),
		}
		p.workers = append(p.workers, worker)

		p.wg.Add(1)
		go worker.start()
		atomic.AddInt32(&p.workerCount, 1)
	}

	p.logger.Info("协程池启动成功",
		zap.Int("max_workers", int(p.maxWorkers)),
		zap.Int("queue_size", int(p.maxQueueSize)),
	)

	return nil
}

// Stop 停止协程池
func (p *GoroutinePoolImpl) Stop(ctx context.Context) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.running {
		return fmt.Errorf("协程池未运行")
	}

	// 取消上下文
	if p.cancel != nil {
		p.cancel()
	}

	// 停止所有工作协程
	for _, worker := range p.workers {
		select {
		case worker.quit <- true:
		default:
		}
	}

	// 等待所有工作协程退出
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 正常完成
	case <-time.After(5 * time.Second):
		// 超时，强制继续
		p.logger.Warn("等待工作协程退出超时")
	}

	p.running = false
	atomic.StoreInt32(&p.workerCount, 0)
	p.workers = p.workers[:0]

	p.logger.Info("协程池已停止",
		zap.Int64("processed_tasks", atomic.LoadInt64(&p.processedTasks)),
		zap.Int64("failed_tasks", atomic.LoadInt64(&p.failedTasks)),
	)

	return nil
}

// Submit 提交任务
func (p *GoroutinePoolImpl) Submit(task Task) error {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		return fmt.Errorf("协程池未运行")
	}
	p.mu.RUnlock()

	select {
	case p.queue <- task:
		atomic.AddInt32(&p.queueSize, 1)
		return nil
	default:
		return fmt.Errorf("任务队列已满")
	}
}

// SubmitBatch 批量提交任务
func (p *GoroutinePoolImpl) SubmitBatch(tasks []Task) error {
	p.mu.RLock()
	if !p.running {
		p.mu.RUnlock()
		return fmt.Errorf("协程池未运行")
	}
	p.mu.RUnlock()

	for _, task := range tasks {
		select {
		case p.queue <- task:
			atomic.AddInt32(&p.queueSize, 1)
		default:
			return fmt.Errorf("任务队列已满，无法提交所有任务")
		}
	}

	return nil
}

// GetStatus 获取协程池状态
func (p *GoroutinePoolImpl) GetStatus() PoolStatus {
	p.mu.RLock()
	defer p.mu.RUnlock()

	status := PoolStatus{
		IsRunning:      p.running,
		WorkerCount:    int(atomic.LoadInt32(&p.workerCount)),
		MaxWorkers:     int(p.maxWorkers),
		QueueSize:      int(atomic.LoadInt32(&p.queueSize)),
		MaxQueueSize:   int(p.maxQueueSize),
		ProcessedTasks: atomic.LoadInt64(&p.processedTasks),
		FailedTasks:    atomic.LoadInt64(&p.failedTasks),
		Uptime:         time.Since(p.startTime),
		LastError:      p.lastError,
	}

	return status
}

// GetWorkerCount 获取工作协程数量
func (p *GoroutinePoolImpl) GetWorkerCount() int {
	return int(atomic.LoadInt32(&p.workerCount))
}

// GetQueueSize 获取队列大小
func (p *GoroutinePoolImpl) GetQueueSize() int {
	return int(atomic.LoadInt32(&p.queueSize))
}

// SetMaxWorkers 设置最大工作协程数
func (p *GoroutinePoolImpl) SetMaxWorkers(count int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if count > 0 {
		p.maxWorkers = int32(count)
	}
}

// SetQueueSize 设置队列大小
func (p *GoroutinePoolImpl) SetQueueSize(size int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if size > 0 {
		p.maxQueueSize = int32(size)
	}
}

// start 启动工作协程
func (w *worker) start() {
	defer w.pool.wg.Done()

	for {
		select {
		case <-w.pool.ctx.Done():
			w.pool.logger.Debug("工作协程收到停止信号", zap.Int("worker_id", w.id))
			return
		case <-w.quit:
			w.pool.logger.Debug("工作协程收到退出信号", zap.Int("worker_id", w.id))
			return
		case task := <-w.pool.queue:
			atomic.AddInt32(&w.pool.queueSize, -1)
			w.executeTask(task)
		}
	}
}

// executeTask 执行任务
func (w *worker) executeTask(task Task) {
	startTime := time.Now()

	// 设置任务超时
	taskCtx, cancel := context.WithTimeout(w.pool.ctx, w.pool.config.TaskTimeout)
	defer cancel()

	err := task.Execute(taskCtx)

	duration := time.Since(startTime)

	if err != nil {
		atomic.AddInt64(&w.pool.failedTasks, 1)
		w.pool.mu.Lock()
		w.pool.lastError = err.Error()
		w.pool.mu.Unlock()

		w.pool.logger.Error("任务执行失败",
			zap.String("task_id", task.GetID()),
			zap.Int("worker_id", w.id),
			zap.Error(err),
			zap.Duration("duration", duration),
		)

		// 重试逻辑
		if task.GetRetryCount() < w.pool.config.RetryCount {
			task.SetRetryCount(task.GetRetryCount() + 1)

			// 延迟重试
			time.Sleep(w.pool.config.RetryDelay)

			select {
			case w.pool.queue <- task:
				atomic.AddInt32(&w.pool.queueSize, 1)
				w.pool.logger.Debug("任务重试",
					zap.String("task_id", task.GetID()),
					zap.Int("retry_count", task.GetRetryCount()),
				)
			default:
				w.pool.logger.Warn("重试任务队列已满",
					zap.String("task_id", task.GetID()),
				)
			}
		}
	} else {
		atomic.AddInt64(&w.pool.processedTasks, 1)
		w.pool.logger.Debug("任务执行成功",
			zap.String("task_id", task.GetID()),
			zap.Int("worker_id", w.id),
			zap.Duration("duration", duration),
		)
	}
}
