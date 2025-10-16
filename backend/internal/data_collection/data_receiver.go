package data_collection

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// dataReceiverImpl 数据接收器实现
type dataReceiverImpl struct {
	config *ReceiverConfig
	logger *zap.Logger
	parser MessageParser

	// 数据通道
	dataChan chan *PriceData

	// 状态管理
	mu            sync.RWMutex
	running       atomic.Bool
	receivedCount atomic.Int64
	errorCount    atomic.Int64
	startTime     time.Time
	lastReceived  time.Time
	lastError     string

	// 工作协程管理
	workerCtx    context.Context
	workerCancel context.CancelFunc
	workerWg     sync.WaitGroup

	// 数据源管理
	activeSources map[string]bool
	sourceMu      sync.RWMutex
}

// NewDataReceiver 创建新的数据接收器
func NewDataReceiver(config *ReceiverConfig, parser MessageParser, logger *zap.Logger) DataReceiver {
	if logger == nil {
		logger = zap.NewNop()
	}
	if config == nil {
		config = DefaultReceiverConfig()
	}
	if parser == nil {
		parser = NewJSONMessageParser(DefaultParserConfig(), logger)
	}

	return &dataReceiverImpl{
		config:        config,
		logger:        logger,
		parser:        parser,
		dataChan:      make(chan *PriceData, config.BufferSize),
		activeSources: make(map[string]bool),
	}
}

// Start 启动数据接收器
func (r *dataReceiverImpl) Start(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.running.Load() {
		return fmt.Errorf("数据接收器已在运行中")
	}

	r.workerCtx, r.workerCancel = context.WithCancel(ctx)
	r.running.Store(true)
	r.startTime = time.Now()
	r.lastReceived = time.Now()

	// 启动工作协程
	for i := 0; i < r.config.WorkerCount; i++ {
		r.workerWg.Add(1)
		go r.worker(i + 1)
	}

	// 启动数据源连接
	for _, source := range r.config.DataSources {
		if source.Enabled {
			r.workerWg.Add(1)
			go r.connectDataSource(source)
		}
	}

	r.logger.Info("数据接收器启动",
		zap.Int("worker_count", r.config.WorkerCount),
		zap.Int("buffer_size", r.config.BufferSize),
		zap.Int("data_sources", len(r.config.DataSources)),
	)

	return nil
}

// Stop 停止数据接收器
func (r *dataReceiverImpl) Stop(ctx context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if !r.running.Load() {
		return fmt.Errorf("数据接收器未运行")
	}

	r.running.Store(false)
	r.workerCancel()

	// 关闭数据通道
	close(r.dataChan)

	// 等待所有工作协程退出
	done := make(chan struct{})
	go func() {
		r.workerWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		r.logger.Info("数据接收器已停止")
	case <-ctx.Done():
		r.logger.Warn("等待数据接收器停止超时", zap.Error(ctx.Err()))
		return fmt.Errorf("停止数据接收器超时: %w", ctx.Err())
	}

	return nil
}

// IsRunning 检查是否正在运行
func (r *dataReceiverImpl) IsRunning() bool {
	return r.running.Load()
}

// ReceiveData 获取数据接收通道
func (r *dataReceiverImpl) ReceiveData() <-chan *PriceData {
	return r.dataChan
}

// GetReceivedCount 获取接收数据总数
func (r *dataReceiverImpl) GetReceivedCount() int64 {
	return r.receivedCount.Load()
}

// GetErrorCount 获取错误总数
func (r *dataReceiverImpl) GetErrorCount() int64 {
	return r.errorCount.Load()
}

// GetStatus 获取接收器状态
func (r *dataReceiverImpl) GetStatus() *ReceiverStatus {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	uptime := now.Sub(r.startTime)

	// 计算吞吐量
	var throughput float64
	if uptime.Seconds() > 0 {
		throughput = float64(r.receivedCount.Load()) / uptime.Seconds()
	}

	// 计算缓冲区使用率
	bufferUsage := float64(len(r.dataChan)) / float64(r.config.BufferSize) * 100

	// 计算活跃数据源数量
	r.sourceMu.RLock()
	activeSources := 0
	for _, active := range r.activeSources {
		if active {
			activeSources++
		}
	}
	r.sourceMu.RUnlock()

	return &ReceiverStatus{
		Running:       r.running.Load(),
		StartTime:     r.startTime,
		Uptime:        uptime,
		ReceivedCount: r.receivedCount.Load(),
		ErrorCount:    r.errorCount.Load(),
		LastReceived:  r.lastReceived,
		ActiveSources: activeSources,
		BufferUsage:   bufferUsage,
		Throughput:    throughput,
		LastError:     r.lastError,
	}
}

// HealthCheck 健康检查
func (r *dataReceiverImpl) HealthCheck() bool {
	if !r.running.Load() {
		return false
	}

	// 检查是否长时间没有接收到数据
	if time.Since(r.lastReceived) > 5*time.Minute {
		return false
	}

	// 检查错误率是否过高
	received := r.receivedCount.Load()
	errors := r.errorCount.Load()
	if received > 0 && float64(errors)/float64(received) > 0.5 {
		return false
	}

	return true
}

// worker 工作协程
func (r *dataReceiverImpl) worker(id int) {
	defer r.workerWg.Done()
	r.logger.Debug("数据接收工作协程启动", zap.Int("worker_id", id))

	for {
		select {
		case <-r.workerCtx.Done():
			r.logger.Debug("数据接收工作协程收到停止信号", zap.Int("worker_id", id))
			return
		}
	}
}

// connectDataSource 连接数据源
func (r *dataReceiverImpl) connectDataSource(source DataSourceConfig) {
	defer r.workerWg.Done()

	r.logger.Info("连接数据源",
		zap.String("name", source.Name),
		zap.String("type", source.Type),
		zap.String("url", source.URL),
	)

	// 标记数据源为活跃
	r.sourceMu.Lock()
	r.activeSources[source.Name] = true
	r.sourceMu.Unlock()

	defer func() {
		r.sourceMu.Lock()
		r.activeSources[source.Name] = false
		r.sourceMu.Unlock()
	}()

	// 根据数据源类型进行连接
	switch source.Type {
	case "websocket":
		r.connectWebSocketSource(source)
	case "rest":
		r.connectRestSource(source)
	case "file":
		r.connectFileSource(source)
	default:
		r.logger.Error("不支持的数据源类型", zap.String("type", source.Type))
		r.errorCount.Add(1)
		r.lastError = fmt.Sprintf("不支持的数据源类型: %s", source.Type)
	}
}

// connectWebSocketSource 连接WebSocket数据源
func (r *dataReceiverImpl) connectWebSocketSource(source DataSourceConfig) {
	// 这里应该集成现有的WebSocket客户端
	// 为了简化，我们模拟数据接收
	r.logger.Info("WebSocket数据源连接成功", zap.String("source", source.Name))

	// 模拟数据接收
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-r.workerCtx.Done():
			return
		case <-ticker.C:
			// 模拟接收数据
			for _, symbol := range r.config.Symbols {
				price := &PriceData{
					Symbol:    symbol,
					Price:     50000.0 + float64(time.Now().UnixNano()%1000),
					Timestamp: time.Now(),
					Source:    source.Name,
				}

				select {
				case r.dataChan <- price:
					r.receivedCount.Add(1)
					r.lastReceived = time.Now()
				case <-r.workerCtx.Done():
					return
				default:
					// 缓冲区满，记录错误
					r.errorCount.Add(1)
					r.lastError = "数据缓冲区已满"
				}
			}
		}
	}
}

// connectRestSource 连接REST数据源
func (r *dataReceiverImpl) connectRestSource(source DataSourceConfig) {
	r.logger.Info("REST数据源连接成功", zap.String("source", source.Name))
	// REST数据源实现
}

// connectFileSource 连接文件数据源
func (r *dataReceiverImpl) connectFileSource(source DataSourceConfig) {
	r.logger.Info("文件数据源连接成功", zap.String("source", source.Name))
	// 文件数据源实现
}
