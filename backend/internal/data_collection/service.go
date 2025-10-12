package data_collection

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// DataCollectionServiceImpl 数据采集服务实现
type DataCollectionServiceImpl struct {
	// 配置
	config *ServiceConfig

	// 日志
	logger *zap.Logger

	// 状态管理
	state     string
	startTime time.Time
	stopTime  time.Time

	// 并发控制
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 指标统计
	totalCollections      int64
	successfulCollections int64
	failedCollections     int64
	activeConnections     int32
	reconnectCount        int64

	// 事件通道
	eventChan chan DataCollectionEvent

	// 健康检查
	lastHealthCheck time.Time
	lastError       error
}

// NewDataCollectionService 创建新的数据采集服务
func NewDataCollectionService(config *ServiceConfig, logger *zap.Logger) *DataCollectionServiceImpl {
	if config == nil {
		config = DefaultServiceConfig()
	}

	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	return &DataCollectionServiceImpl{
		config:    config,
		logger:    logger,
		state:     StateStopped,
		eventChan: make(chan DataCollectionEvent, config.ChannelBufferSize),
	}
}

// Start 启动服务
func (s *DataCollectionServiceImpl) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查服务状态
	if s.state == StateRunning {
		s.logger.Warn("服务已在运行中")
		return nil
	}

	if s.state == StateStarting {
		s.logger.Warn("服务正在启动中")
		return nil
	}

	// 设置状态
	s.state = StateStarting
	s.startTime = time.Now()
	s.lastHealthCheck = time.Now()

	// 创建新的上下文
	s.ctx, s.cancel = context.WithCancel(ctx)

	// 启动服务组件
	if err := s.startComponents(); err != nil {
		s.state = StateError
		s.lastError = err
		return err
	}

	// 设置运行状态
	s.state = StateRunning

	s.logger.Info("数据采集服务启动成功",
		zap.String("state", s.state),
		zap.Time("start_time", s.startTime),
		zap.Int("symbols", len(s.config.Symbols)),
	)

	return nil
}

// Stop 停止服务
func (s *DataCollectionServiceImpl) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 检查服务状态
	if s.state == StateStopped {
		s.logger.Warn("服务已停止")
		return nil
	}

	if s.state == StateStopping {
		s.logger.Warn("服务正在停止中")
		return nil
	}

	// 设置状态
	s.state = StateStopping
	s.stopTime = time.Now()

	// 取消上下文
	if s.cancel != nil {
		s.cancel()
	}

	// 停止服务组件
	s.stopComponents()

	// 等待所有goroutine完成
	s.wg.Wait()

	// 设置停止状态
	s.state = StateStopped

	s.logger.Info("数据采集服务停止成功",
		zap.String("state", s.state),
		zap.Time("stop_time", s.stopTime),
		zap.Duration("uptime", s.stopTime.Sub(s.startTime)),
	)

	return nil
}

// IsRunning 检查服务是否运行
func (s *DataCollectionServiceImpl) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.state == StateRunning
}

// GetStatus 获取服务状态
func (s *DataCollectionServiceImpl) GetStatus() *ServiceStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status := &ServiceStatus{
		State:             s.state,
		ActiveConnections: int(atomic.LoadInt32(&s.activeConnections)),
		TotalCollections:  atomic.LoadInt64(&s.totalCollections),
		ErrorCount:        atomic.LoadInt64(&s.failedCollections),
		LastUpdated:       time.Now(),
		StartTime:         s.startTime,
	}

	// 计算运行时间
	if s.state == StateRunning && !s.startTime.IsZero() {
		status.Uptime = time.Since(s.startTime)
	} else if !s.stopTime.IsZero() && !s.startTime.IsZero() {
		status.Uptime = s.stopTime.Sub(s.startTime)
	}

	// 设置最后错误
	if s.lastError != nil {
		status.LastError = s.lastError.Error()
	}

	return status
}

// HealthCheck 健康检查
func (s *DataCollectionServiceImpl) HealthCheck() *HealthStatus {
	s.mu.RLock()
	defer s.mu.RUnlock()

	health := &HealthStatus{
		Timestamp: time.Now(),
		Details:   make(map[string]interface{}),
		Checks:    []HealthCheck{},
	}

	// 基础健康检查
	health.Checks = append(health.Checks, HealthCheck{
		Name:     "service_state",
		Status:   s.getHealthStatus(),
		Duration: 0,
	})

	// 连接健康检查
	health.Checks = append(health.Checks, HealthCheck{
		Name:     "connections",
		Status:   s.checkConnectionsHealth(),
		Duration: 0,
	})

	// 数据采集健康检查
	health.Checks = append(health.Checks, HealthCheck{
		Name:     "data_collection",
		Status:   s.checkDataCollectionHealth(),
		Duration: 0,
	})

	// 设置整体健康状态
	health.Status = s.getOverallHealthStatus(health.Checks)
	health.IsHealthy = health.Status == HealthStatusHealthy

	// 设置详细信息
	health.Details["state"] = s.state
	health.Details["uptime"] = time.Since(s.startTime).String()
	health.Details["active_connections"] = atomic.LoadInt32(&s.activeConnections)
	health.Details["total_collections"] = atomic.LoadInt64(&s.totalCollections)
	health.Details["error_rate"] = s.getErrorRate()

	return health
}

// GetConfig 获取服务配置
func (s *DataCollectionServiceImpl) GetConfig() *ServiceConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.config
}

// UpdateConfig 更新服务配置
func (s *DataCollectionServiceImpl) UpdateConfig(config *ServiceConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if config == nil {
		return ErrInvalidConfig
	}

	// 验证配置
	if err := s.validateConfig(config); err != nil {
		return err
	}

	// 更新配置
	s.config = config

	s.logger.Info("服务配置已更新",
		zap.Int("max_connections", config.MaxConnections),
		zap.Int("symbols_count", len(config.Symbols)),
	)

	return nil
}

// GetMetrics 获取服务指标
func (s *DataCollectionServiceImpl) GetMetrics() *ServiceMetrics {
	s.mu.RLock()
	defer s.mu.RUnlock()

	metrics := &ServiceMetrics{
		Uptime:                time.Since(s.startTime),
		TotalCollections:      atomic.LoadInt64(&s.totalCollections),
		SuccessfulCollections: atomic.LoadInt64(&s.successfulCollections),
		FailedCollections:     atomic.LoadInt64(&s.failedCollections),
		ErrorRate:             s.getErrorRate(),
		ActiveConnections:     int(atomic.LoadInt32(&s.activeConnections)),
		ReconnectCount:        atomic.LoadInt64(&s.reconnectCount),
		LastUpdated:           time.Now(),
	}

	// 计算吞吐量
	if metrics.Uptime > 0 {
		metrics.ThroughputPerSecond = float64(metrics.TotalCollections) / metrics.Uptime.Seconds()
	}

	return metrics
}

// 私有方法

// startComponents 启动服务组件
func (s *DataCollectionServiceImpl) startComponents() error {
	// 启动健康检查
	s.wg.Add(1)
	go s.healthCheckLoop()

	// 启动事件处理
	s.wg.Add(1)
	go s.eventHandlerLoop()

	// 启动数据采集
	s.wg.Add(1)
	go s.dataCollectionLoop()

	return nil
}

// stopComponents 停止服务组件
func (s *DataCollectionServiceImpl) stopComponents() {
	// 关闭事件通道
	close(s.eventChan)
}

// healthCheckLoop 健康检查循环
func (s *DataCollectionServiceImpl) healthCheckLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.HealthCheckInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performHealthCheck()
		}
	}
}

// eventHandlerLoop 事件处理循环
func (s *DataCollectionServiceImpl) eventHandlerLoop() {
	defer s.wg.Done()

	for {
		select {
		case <-s.ctx.Done():
			return
		case event, ok := <-s.eventChan:
			if !ok {
				return
			}
			s.handleEvent(event)
		}
	}
}

// dataCollectionLoop 数据采集循环
func (s *DataCollectionServiceImpl) dataCollectionLoop() {
	defer s.wg.Done()

	ticker := time.NewTicker(s.config.CollectionInterval)
	defer ticker.Stop()

	for {
		select {
		case <-s.ctx.Done():
			return
		case <-ticker.C:
			s.performDataCollection()
		}
	}
}

// performHealthCheck 执行健康检查
func (s *DataCollectionServiceImpl) performHealthCheck() {
	s.lastHealthCheck = time.Now()

	// 发送健康检查事件
	select {
	case s.eventChan <- DataCollectionEvent{
		Type:      EventTypeHealthCheck,
		Timestamp: time.Now(),
	}:
	default:
		// 通道已满，跳过
	}
}

// performDataCollection 执行数据采集
func (s *DataCollectionServiceImpl) performDataCollection() {
	// 这里将实现具体的数据采集逻辑
	// 目前只是模拟
	atomic.AddInt64(&s.totalCollections, 1)
	atomic.AddInt64(&s.successfulCollections, 1)
}

// handleEvent 处理事件
func (s *DataCollectionServiceImpl) handleEvent(event DataCollectionEvent) {
	s.logger.Debug("处理数据采集事件",
		zap.String("type", event.Type),
		zap.String("symbol", event.Symbol),
		zap.Time("timestamp", event.Timestamp),
	)
}

// getHealthStatus 获取健康状态
func (s *DataCollectionServiceImpl) getHealthStatus() string {
	switch s.state {
	case StateRunning:
		return HealthStatusHealthy
	case StateStopped:
		return HealthStatusUnhealthy
	case StateError:
		return HealthStatusUnhealthy
	default:
		return HealthStatusDegraded
	}
}

// checkConnectionsHealth 检查连接健康状态
func (s *DataCollectionServiceImpl) checkConnectionsHealth() string {
	activeConnections := atomic.LoadInt32(&s.activeConnections)
	if activeConnections > 0 {
		return HealthStatusHealthy
	}
	// 如果没有活跃连接，检查服务是否刚启动
	if s.state == StateRunning && time.Since(s.startTime) < 5*time.Second {
		// 服务刚启动，给一些时间建立连接
		return HealthStatusDegraded
	}
	return HealthStatusUnhealthy
}

// checkDataCollectionHealth 检查数据采集健康状态
func (s *DataCollectionServiceImpl) checkDataCollectionHealth() string {
	errorRate := s.getErrorRate()
	if errorRate < 0.1 { // 错误率小于10%
		return HealthStatusHealthy
	} else if errorRate < 0.3 { // 错误率小于30%
		return HealthStatusDegraded
	}
	return HealthStatusUnhealthy
}

// getOverallHealthStatus 获取整体健康状态
func (s *DataCollectionServiceImpl) getOverallHealthStatus(checks []HealthCheck) string {
	healthyCount := 0
	totalCount := len(checks)

	for _, check := range checks {
		if check.Status == HealthStatusHealthy {
			healthyCount++
		}
	}

	if healthyCount == totalCount {
		return HealthStatusHealthy
	} else if healthyCount > totalCount/2 {
		return HealthStatusDegraded
	}
	return HealthStatusUnhealthy
}

// getErrorRate 获取错误率
func (s *DataCollectionServiceImpl) getErrorRate() float64 {
	total := atomic.LoadInt64(&s.totalCollections)
	if total == 0 {
		return 0
	}

	failed := atomic.LoadInt64(&s.failedCollections)
	return float64(failed) / float64(total)
}

// validateConfig 验证配置
func (s *DataCollectionServiceImpl) validateConfig(config *ServiceConfig) error {
	if config.MaxConnections <= 0 {
		return ErrInvalidMaxConnections
	}

	if config.ReconnectInterval <= 0 {
		return ErrInvalidReconnectInterval
	}

	if config.HealthCheckInterval <= 0 {
		return ErrInvalidHealthCheckInterval
	}

	if len(config.Symbols) == 0 {
		return ErrEmptySymbols
	}

	return nil
}
