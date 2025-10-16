package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// PerformanceMonitor 性能监控器接口
type PerformanceMonitor interface {
	// 监控管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 指标记录
	RecordLatency(operation string, duration time.Duration)
	RecordThroughput(operation string, count int64)
	RecordMemoryUsage(operation string, bytes int64)
	RecordError(operation string, errorType string)

	// 指标查询
	GetLatencyStats(operation string) *LatencyStats
	GetThroughputStats(operation string) *ThroughputStats
	GetMemoryStats(operation string) *MemoryStats
	GetErrorStats(operation string) *ErrorStats
	GetOverallStats() *OverallStats

	// 告警管理
	SetLatencyThreshold(operation string, threshold time.Duration)
	SetThroughputThreshold(operation string, threshold int64)
	SetMemoryThreshold(operation string, threshold int64)
	GetAlerts() []*PerformanceAlert

	// 配置管理
	SetSamplingRate(rate float64)
	SetAggregationInterval(interval time.Duration)
	SetRetentionPeriod(period time.Duration)
}

// LatencyStats 延迟统计
type LatencyStats struct {
	Operation   string        `json:"operation"`
	Count       int64         `json:"count"`
	MinLatency  time.Duration `json:"min_latency"`
	MaxLatency  time.Duration `json:"max_latency"`
	AvgLatency  time.Duration `json:"avg_latency"`
	P50Latency  time.Duration `json:"p50_latency"`
	P90Latency  time.Duration `json:"p90_latency"`
	P95Latency  time.Duration `json:"p95_latency"`
	P99Latency  time.Duration `json:"p99_latency"`
	LastLatency time.Duration `json:"last_latency"`
	LastUpdated time.Time     `json:"last_updated"`
}

// ThroughputStats 吞吐量统计
type ThroughputStats struct {
	Operation   string    `json:"operation"`
	Count       int64     `json:"count"`
	TotalCount  int64     `json:"total_count"`
	Rate        float64   `json:"rate"` // 每秒操作数
	PeakRate    float64   `json:"peak_rate"`
	AvgRate     float64   `json:"avg_rate"`
	LastCount   int64     `json:"last_count"`
	LastUpdated time.Time `json:"last_updated"`
	WindowStart time.Time `json:"window_start"`
}

// MemoryStats 内存统计
type MemoryStats struct {
	Operation   string    `json:"operation"`
	Count       int64     `json:"count"`
	TotalBytes  int64     `json:"total_bytes"`
	AvgBytes    int64     `json:"avg_bytes"`
	PeakBytes   int64     `json:"peak_bytes"`
	LastBytes   int64     `json:"last_bytes"`
	LastUpdated time.Time `json:"last_updated"`
}

// ErrorStats 错误统计
type ErrorStats struct {
	Operation   string           `json:"operation"`
	TotalErrors int64            `json:"total_errors"`
	ErrorRate   float64          `json:"error_rate"`
	ErrorTypes  map[string]int64 `json:"error_types"`
	LastError   time.Time        `json:"last_error"`
	LastUpdated time.Time        `json:"last_updated"`
}

// OverallStats 总体统计
type OverallStats struct {
	TotalOperations int64            `json:"total_operations"`
	TotalLatency    time.Duration    `json:"total_latency"`
	TotalThroughput int64            `json:"total_throughput"`
	TotalMemory     int64            `json:"total_memory"`
	TotalErrors     int64            `json:"total_errors"`
	AvgLatency      time.Duration    `json:"avg_latency"`
	AvgThroughput   float64          `json:"avg_throughput"`
	AvgMemory       int64            `json:"avg_memory"`
	ErrorRate       float64          `json:"error_rate"`
	StartTime       time.Time        `json:"start_time"`
	LastUpdated     time.Time        `json:"last_updated"`
	OperationStats  map[string]int64 `json:"operation_stats"`
}

// PerformanceAlert 性能告警
type PerformanceAlert struct {
	ID         string    `json:"id"`
	Type       string    `json:"type"`
	Operation  string    `json:"operation"`
	Severity   string    `json:"severity"`
	Message    string    `json:"message"`
	Value      float64   `json:"value"`
	Threshold  float64   `json:"threshold"`
	Timestamp  time.Time `json:"timestamp"`
	Resolved   bool      `json:"resolved"`
	ResolvedAt time.Time `json:"resolved_at"`
}

// PerformanceConfig 性能监控配置
type PerformanceConfig struct {
	SamplingRate         float64                  `json:"sampling_rate" yaml:"sampling_rate"`
	AggregationInterval  time.Duration            `json:"aggregation_interval" yaml:"aggregation_interval"`
	RetentionPeriod      time.Duration            `json:"retention_period" yaml:"retention_period"`
	LatencyThresholds    map[string]time.Duration `json:"latency_thresholds" yaml:"latency_thresholds"`
	ThroughputThresholds map[string]int64         `json:"throughput_thresholds" yaml:"throughput_thresholds"`
	MemoryThresholds     map[string]int64         `json:"memory_thresholds" yaml:"memory_thresholds"`
	EnableAlerts         bool                     `json:"enable_alerts" yaml:"enable_alerts"`
	AlertCooldown        time.Duration            `json:"alert_cooldown" yaml:"alert_cooldown"`
}

// DefaultPerformanceConfig 默认性能监控配置
func DefaultPerformanceConfig() *PerformanceConfig {
	return &PerformanceConfig{
		SamplingRate:         1.0,
		AggregationInterval:  1 * time.Second,
		RetentionPeriod:      24 * time.Hour,
		LatencyThresholds:    make(map[string]time.Duration),
		ThroughputThresholds: make(map[string]int64),
		MemoryThresholds:     make(map[string]int64),
		EnableAlerts:         true,
		AlertCooldown:        5 * time.Minute,
	}
}

// PerformanceMonitorImpl 性能监控器实现
type PerformanceMonitorImpl struct {
	// 配置
	config *PerformanceConfig

	// 统计数据
	latencyStats    map[string]*LatencyStats
	throughputStats map[string]*ThroughputStats
	memoryStats     map[string]*MemoryStats
	errorStats      map[string]*ErrorStats
	overallStats    *OverallStats

	// 锁
	latencyMu    sync.RWMutex
	throughputMu sync.RWMutex
	memoryMu     sync.RWMutex
	errorMu      sync.RWMutex
	overallMu    sync.RWMutex

	// 告警
	alerts         []*PerformanceAlert
	alertMu        sync.RWMutex
	alertCooldowns map[string]time.Time

	// 控制
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.RWMutex

	// 聚合定时器
	aggregationTicker *time.Ticker

	// 日志记录器
	logger interface{}
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(config *PerformanceConfig, logger interface{}) PerformanceMonitor {
	if config == nil {
		config = DefaultPerformanceConfig()
	}

	return &PerformanceMonitorImpl{
		config:          config,
		latencyStats:    make(map[string]*LatencyStats),
		throughputStats: make(map[string]*ThroughputStats),
		memoryStats:     make(map[string]*MemoryStats),
		errorStats:      make(map[string]*ErrorStats),
		overallStats:    &OverallStats{StartTime: time.Now()},
		alerts:          make([]*PerformanceAlert, 0),
		alertCooldowns:  make(map[string]time.Time),
		logger:          logger,
	}
}

// Start 启动性能监控器
func (pm *PerformanceMonitorImpl) Start(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if pm.running {
		return &PerformanceError{
			Type:    "ALREADY_RUNNING",
			Message: "性能监控器已在运行",
		}
	}

	pm.ctx, pm.cancel = context.WithCancel(ctx)
	pm.running = true

	// 启动聚合定时器
	pm.aggregationTicker = time.NewTicker(pm.config.AggregationInterval)
	go pm.aggregationLoop()

	return nil
}

// Stop 停止性能监控器
func (pm *PerformanceMonitorImpl) Stop(ctx context.Context) error {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	if !pm.running {
		return &PerformanceError{
			Type:    "NOT_RUNNING",
			Message: "性能监控器未运行",
		}
	}

	pm.running = false
	pm.cancel()

	// 停止聚合定时器
	if pm.aggregationTicker != nil {
		pm.aggregationTicker.Stop()
	}

	return nil
}

// IsRunning 检查是否运行
func (pm *PerformanceMonitorImpl) IsRunning() bool {
	pm.mu.RLock()
	defer pm.mu.RUnlock()
	return pm.running
}

// RecordLatency 记录延迟
func (pm *PerformanceMonitorImpl) RecordLatency(operation string, duration time.Duration) {
	if !pm.shouldSample() {
		return
	}

	pm.latencyMu.Lock()
	defer pm.latencyMu.Unlock()

	stats, exists := pm.latencyStats[operation]
	if !exists {
		stats = &LatencyStats{
			Operation:   operation,
			MinLatency:  duration,
			MaxLatency:  duration,
			AvgLatency:  duration,
			P50Latency:  duration,
			P90Latency:  duration,
			P95Latency:  duration,
			P99Latency:  duration,
			LastLatency: duration,
			LastUpdated: time.Now(),
		}
		pm.latencyStats[operation] = stats
	}

	// 更新统计
	stats.Count++
	stats.LastLatency = duration
	stats.LastUpdated = time.Now()

	if duration < stats.MinLatency {
		stats.MinLatency = duration
	}
	if duration > stats.MaxLatency {
		stats.MaxLatency = duration
	}

	// 简单的移动平均
	stats.AvgLatency = (stats.AvgLatency + duration) / 2

	// 更新总体统计
	pm.updateOverallStats("latency", duration)

	// 检查告警
	pm.checkLatencyAlert(operation, duration)
}

// RecordThroughput 记录吞吐量
func (pm *PerformanceMonitorImpl) RecordThroughput(operation string, count int64) {
	if !pm.shouldSample() {
		return
	}

	pm.throughputMu.Lock()
	defer pm.throughputMu.Unlock()

	stats, exists := pm.throughputStats[operation]
	if !exists {
		stats = &ThroughputStats{
			Operation:   operation,
			WindowStart: time.Now(),
			LastUpdated: time.Now(),
		}
		pm.throughputStats[operation] = stats
	}

	// 更新统计
	stats.Count += count
	stats.TotalCount += count
	stats.LastCount = count
	stats.LastUpdated = time.Now()

	// 计算速率
	elapsed := time.Since(stats.WindowStart)
	if elapsed > 0 {
		stats.Rate = float64(stats.Count) / elapsed.Seconds()
		if stats.Rate > stats.PeakRate {
			stats.PeakRate = stats.Rate
		}
		stats.AvgRate = float64(stats.TotalCount) / time.Since(stats.WindowStart).Seconds()
	}

	// 更新总体统计
	pm.updateOverallStats("throughput", count)

	// 检查告警
	pm.checkThroughputAlert(operation, stats.Rate)
}

// RecordMemoryUsage 记录内存使用
func (pm *PerformanceMonitorImpl) RecordMemoryUsage(operation string, bytes int64) {
	if !pm.shouldSample() {
		return
	}

	pm.memoryMu.Lock()
	defer pm.memoryMu.Unlock()

	stats, exists := pm.memoryStats[operation]
	if !exists {
		stats = &MemoryStats{
			Operation:   operation,
			PeakBytes:   bytes,
			LastUpdated: time.Now(),
		}
		pm.memoryStats[operation] = stats
	}

	// 更新统计
	stats.Count++
	stats.TotalBytes += bytes
	stats.LastBytes = bytes
	stats.LastUpdated = time.Now()

	// 计算平均值
	stats.AvgBytes = stats.TotalBytes / stats.Count

	if bytes > stats.PeakBytes {
		stats.PeakBytes = bytes
	}

	// 更新总体统计
	pm.updateOverallStats("memory", bytes)

	// 检查告警
	pm.checkMemoryAlert(operation, bytes)
}

// RecordError 记录错误
func (pm *PerformanceMonitorImpl) RecordError(operation string, errorType string) {
	if !pm.shouldSample() {
		return
	}

	pm.errorMu.Lock()
	defer pm.errorMu.Unlock()

	stats, exists := pm.errorStats[operation]
	if !exists {
		stats = &ErrorStats{
			Operation:   operation,
			ErrorTypes:  make(map[string]int64),
			LastUpdated: time.Now(),
		}
		pm.errorStats[operation] = stats
	}

	// 更新统计
	stats.TotalErrors++
	stats.ErrorTypes[errorType]++
	stats.LastError = time.Now()
	stats.LastUpdated = time.Now()

	// 计算错误率
	pm.overallMu.RLock()
	totalOps := pm.overallStats.TotalOperations
	pm.overallMu.RUnlock()

	if totalOps > 0 {
		stats.ErrorRate = float64(stats.TotalErrors) / float64(totalOps)
	}

	// 更新总体统计
	pm.updateOverallStats("error", 1)
}

// GetLatencyStats 获取延迟统计
func (pm *PerformanceMonitorImpl) GetLatencyStats(operation string) *LatencyStats {
	pm.latencyMu.RLock()
	defer pm.latencyMu.RUnlock()

	stats, exists := pm.latencyStats[operation]
	if !exists {
		return nil
	}

	// 返回副本
	return &LatencyStats{
		Operation:   stats.Operation,
		Count:       stats.Count,
		MinLatency:  stats.MinLatency,
		MaxLatency:  stats.MaxLatency,
		AvgLatency:  stats.AvgLatency,
		P50Latency:  stats.P50Latency,
		P90Latency:  stats.P90Latency,
		P95Latency:  stats.P95Latency,
		P99Latency:  stats.P99Latency,
		LastLatency: stats.LastLatency,
		LastUpdated: stats.LastUpdated,
	}
}

// GetThroughputStats 获取吞吐量统计
func (pm *PerformanceMonitorImpl) GetThroughputStats(operation string) *ThroughputStats {
	pm.throughputMu.RLock()
	defer pm.throughputMu.RUnlock()

	stats, exists := pm.throughputStats[operation]
	if !exists {
		return nil
	}

	// 返回副本
	return &ThroughputStats{
		Operation:   stats.Operation,
		Count:       stats.Count,
		TotalCount:  stats.TotalCount,
		Rate:        stats.Rate,
		PeakRate:    stats.PeakRate,
		AvgRate:     stats.AvgRate,
		LastCount:   stats.LastCount,
		LastUpdated: stats.LastUpdated,
		WindowStart: stats.WindowStart,
	}
}

// GetMemoryStats 获取内存统计
func (pm *PerformanceMonitorImpl) GetMemoryStats(operation string) *MemoryStats {
	pm.memoryMu.RLock()
	defer pm.memoryMu.RUnlock()

	stats, exists := pm.memoryStats[operation]
	if !exists {
		return nil
	}

	// 返回副本
	return &MemoryStats{
		Operation:   stats.Operation,
		Count:       stats.Count,
		TotalBytes:  stats.TotalBytes,
		AvgBytes:    stats.AvgBytes,
		PeakBytes:   stats.PeakBytes,
		LastBytes:   stats.LastBytes,
		LastUpdated: stats.LastUpdated,
	}
}

// GetErrorStats 获取错误统计
func (pm *PerformanceMonitorImpl) GetErrorStats(operation string) *ErrorStats {
	pm.errorMu.RLock()
	defer pm.errorMu.RUnlock()

	stats, exists := pm.errorStats[operation]
	if !exists {
		return nil
	}

	// 返回副本
	errorTypes := make(map[string]int64)
	for k, v := range stats.ErrorTypes {
		errorTypes[k] = v
	}

	return &ErrorStats{
		Operation:   stats.Operation,
		TotalErrors: stats.TotalErrors,
		ErrorRate:   stats.ErrorRate,
		ErrorTypes:  errorTypes,
		LastError:   stats.LastError,
		LastUpdated: stats.LastUpdated,
	}
}

// GetOverallStats 获取总体统计
func (pm *PerformanceMonitorImpl) GetOverallStats() *OverallStats {
	pm.overallMu.RLock()
	defer pm.overallMu.RUnlock()

	// 返回副本
	operationStats := make(map[string]int64)
	for k, v := range pm.overallStats.OperationStats {
		operationStats[k] = v
	}

	return &OverallStats{
		TotalOperations: pm.overallStats.TotalOperations,
		TotalLatency:    pm.overallStats.TotalLatency,
		TotalThroughput: pm.overallStats.TotalThroughput,
		TotalMemory:     pm.overallStats.TotalMemory,
		TotalErrors:     pm.overallStats.TotalErrors,
		AvgLatency:      pm.overallStats.AvgLatency,
		AvgThroughput:   pm.overallStats.AvgThroughput,
		AvgMemory:       pm.overallStats.AvgMemory,
		ErrorRate:       pm.overallStats.ErrorRate,
		StartTime:       pm.overallStats.StartTime,
		LastUpdated:     pm.overallStats.LastUpdated,
		OperationStats:  operationStats,
	}
}

// SetLatencyThreshold 设置延迟阈值
func (pm *PerformanceMonitorImpl) SetLatencyThreshold(operation string, threshold time.Duration) {
	pm.config.LatencyThresholds[operation] = threshold
}

// SetThroughputThreshold 设置吞吐量阈值
func (pm *PerformanceMonitorImpl) SetThroughputThreshold(operation string, threshold int64) {
	pm.config.ThroughputThresholds[operation] = threshold
}

// SetMemoryThreshold 设置内存阈值
func (pm *PerformanceMonitorImpl) SetMemoryThreshold(operation string, threshold int64) {
	pm.config.MemoryThresholds[operation] = threshold
}

// GetAlerts 获取告警
func (pm *PerformanceMonitorImpl) GetAlerts() []*PerformanceAlert {
	pm.alertMu.RLock()
	defer pm.alertMu.RUnlock()

	// 返回副本
	alerts := make([]*PerformanceAlert, len(pm.alerts))
	for i, alert := range pm.alerts {
		alerts[i] = &PerformanceAlert{
			ID:         alert.ID,
			Type:       alert.Type,
			Operation:  alert.Operation,
			Severity:   alert.Severity,
			Message:    alert.Message,
			Value:      alert.Value,
			Threshold:  alert.Threshold,
			Timestamp:  alert.Timestamp,
			Resolved:   alert.Resolved,
			ResolvedAt: alert.ResolvedAt,
		}
	}

	return alerts
}

// SetSamplingRate 设置采样率
func (pm *PerformanceMonitorImpl) SetSamplingRate(rate float64) {
	pm.config.SamplingRate = rate
}

// SetAggregationInterval 设置聚合间隔
func (pm *PerformanceMonitorImpl) SetAggregationInterval(interval time.Duration) {
	pm.config.AggregationInterval = interval
}

// SetRetentionPeriod 设置保留期
func (pm *PerformanceMonitorImpl) SetRetentionPeriod(period time.Duration) {
	pm.config.RetentionPeriod = period
}

// 辅助方法

func (pm *PerformanceMonitorImpl) shouldSample() bool {
	// 简单的采样逻辑，实际实现中可以使用更复杂的采样策略
	return true
}

func (pm *PerformanceMonitorImpl) updateOverallStats(metricType string, value interface{}) {
	pm.overallMu.Lock()
	defer pm.overallMu.Unlock()

	pm.overallStats.LastUpdated = time.Now()

	switch metricType {
	case "latency":
		if duration, ok := value.(time.Duration); ok {
			pm.overallStats.TotalLatency += duration
			if pm.overallStats.TotalOperations > 0 {
				pm.overallStats.AvgLatency = pm.overallStats.TotalLatency / time.Duration(pm.overallStats.TotalOperations)
			}
		}
	case "throughput":
		if count, ok := value.(int64); ok {
			pm.overallStats.TotalThroughput += count
			elapsed := time.Since(pm.overallStats.StartTime).Seconds()
			if elapsed > 0 {
				pm.overallStats.AvgThroughput = float64(pm.overallStats.TotalThroughput) / elapsed
			}
		}
	case "memory":
		if bytes, ok := value.(int64); ok {
			pm.overallStats.TotalMemory += bytes
			if pm.overallStats.TotalOperations > 0 {
				pm.overallStats.AvgMemory = pm.overallStats.TotalMemory / pm.overallStats.TotalOperations
			}
		}
	case "error":
		pm.overallStats.TotalErrors++
		if pm.overallStats.TotalOperations > 0 {
			pm.overallStats.ErrorRate = float64(pm.overallStats.TotalErrors) / float64(pm.overallStats.TotalOperations)
		}
	}

	pm.overallStats.TotalOperations++
}

func (pm *PerformanceMonitorImpl) checkLatencyAlert(operation string, latency time.Duration) {
	if !pm.config.EnableAlerts {
		return
	}

	threshold, exists := pm.config.LatencyThresholds[operation]
	if !exists {
		return
	}

	if latency > threshold {
		pm.createAlert("latency", operation, "high",
			fmt.Sprintf("延迟超过阈值: %v > %v", latency, threshold),
			float64(latency), float64(threshold))
	}
}

func (pm *PerformanceMonitorImpl) checkThroughputAlert(operation string, rate float64) {
	if !pm.config.EnableAlerts {
		return
	}

	threshold, exists := pm.config.ThroughputThresholds[operation]
	if !exists {
		return
	}

	if int64(rate) < threshold {
		pm.createAlert("throughput", operation, "low",
			fmt.Sprintf("吞吐量低于阈值: %.2f < %d", rate, threshold),
			rate, float64(threshold))
	}
}

func (pm *PerformanceMonitorImpl) checkMemoryAlert(operation string, bytes int64) {
	if !pm.config.EnableAlerts {
		return
	}

	threshold, exists := pm.config.MemoryThresholds[operation]
	if !exists {
		return
	}

	if bytes > threshold {
		pm.createAlert("memory", operation, "high",
			fmt.Sprintf("内存使用超过阈值: %d > %d", bytes, threshold),
			float64(bytes), float64(threshold))
	}
}

func (pm *PerformanceMonitorImpl) createAlert(alertType, operation, severity, message string, value, threshold float64) {
	alertKey := fmt.Sprintf("%s_%s_%s", alertType, operation, severity)

	// 检查冷却期
	if lastAlert, exists := pm.alertCooldowns[alertKey]; exists {
		if time.Since(lastAlert) < pm.config.AlertCooldown {
			return
		}
	}

	alert := &PerformanceAlert{
		ID:        fmt.Sprintf("alert_%d", time.Now().UnixNano()),
		Type:      alertType,
		Operation: operation,
		Severity:  severity,
		Message:   message,
		Value:     value,
		Threshold: threshold,
		Timestamp: time.Now(),
		Resolved:  false,
	}

	pm.alertMu.Lock()
	pm.alerts = append(pm.alerts, alert)
	pm.alertCooldowns[alertKey] = time.Now()
	pm.alertMu.Unlock()
}

func (pm *PerformanceMonitorImpl) aggregationLoop() {
	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-pm.aggregationTicker.C:
			pm.performAggregation()
		}
	}
}

func (pm *PerformanceMonitorImpl) performAggregation() {
	// 这里可以执行数据聚合、清理过期数据等操作
	// 实际实现中会包含更复杂的聚合逻辑
}

// PerformanceError 性能监控错误
type PerformanceError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (e *PerformanceError) Error() string {
	return e.Message
}
