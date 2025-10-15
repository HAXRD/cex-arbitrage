package data_collection

import (
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"go.uber.org/zap"
)

// metricCollectorImpl 指标收集器实现
type metricCollectorImpl struct {
	logger *zap.Logger

	// 控制
	running atomic.Bool
	mu      sync.RWMutex

	// 指标存储
	metrics []Metric

	// 统计信息
	stats *MetricStats

	// 性能统计
	collectionTimes []time.Duration
}

// NewMetricCollector 创建指标收集器
func NewMetricCollector(logger *zap.Logger) MetricCollector {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &metricCollectorImpl{
		logger:  logger,
		metrics: make([]Metric, 0),
		stats: &MetricStats{
			StartTime: time.Now(),
		},
		collectionTimes: make([]time.Duration, 0),
	}
}

// Start 启动指标收集器
func (m *metricCollectorImpl) Start() error {
	if m.running.Load() {
		return fmt.Errorf("指标收集器已经在运行")
	}

	m.running.Store(true)
	m.logger.Info("指标收集器已启动")
	return nil
}

// Stop 停止指标收集器
func (m *metricCollectorImpl) Stop() error {
	if !m.running.Load() {
		return fmt.Errorf("指标收集器未运行")
	}

	m.running.Store(false)
	m.logger.Info("指标收集器已停止")
	return nil
}

// IsRunning 检查是否正在运行
func (m *metricCollectorImpl) IsRunning() bool {
	return m.running.Load()
}

// CollectCounter 收集计数器指标
func (m *metricCollectorImpl) CollectCounter(name string, value float64, labels map[string]string) error {
	if !m.running.Load() {
		return fmt.Errorf("指标收集器未运行")
	}

	start := time.Now()
	defer func() {
		m.updateCollectionStats(time.Since(start))
	}()

	metric := &CounterMetric{
		Name:      name,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	m.addMetric(metric)
	m.stats.CounterMetrics++

	return nil
}

// CollectGauge 收集仪表盘指标
func (m *metricCollectorImpl) CollectGauge(name string, value float64, labels map[string]string) error {
	if !m.running.Load() {
		return fmt.Errorf("指标收集器未运行")
	}

	start := time.Now()
	defer func() {
		m.updateCollectionStats(time.Since(start))
	}()

	metric := &GaugeMetric{
		Name:      name,
		Value:     value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	m.addMetric(metric)
	m.stats.GaugeMetrics++

	return nil
}

// CollectHistogram 收集直方图指标
func (m *metricCollectorImpl) CollectHistogram(name string, value float64, labels map[string]string) error {
	if !m.running.Load() {
		return fmt.Errorf("指标收集器未运行")
	}

	start := time.Now()
	defer func() {
		m.updateCollectionStats(time.Since(start))
	}()

	metric := &HistogramMetric{
		Name:      name,
		Value:     value,
		Count:     1,
		Sum:       value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	m.addMetric(metric)
	m.stats.HistogramMetrics++

	return nil
}

// CollectSummary 收集摘要指标
func (m *metricCollectorImpl) CollectSummary(name string, value float64, labels map[string]string) error {
	if !m.running.Load() {
		return fmt.Errorf("指标收集器未运行")
	}

	start := time.Now()
	defer func() {
		m.updateCollectionStats(time.Since(start))
	}()

	metric := &SummaryMetric{
		Name:      name,
		Value:     value,
		Count:     1,
		Sum:       value,
		Labels:    labels,
		Timestamp: time.Now(),
	}

	m.addMetric(metric)
	m.stats.SummaryMetrics++

	return nil
}

// CollectBatch 批量收集指标
func (m *metricCollectorImpl) CollectBatch(metrics []Metric) error {
	if !m.running.Load() {
		return fmt.Errorf("指标收集器未运行")
	}

	start := time.Now()
	defer func() {
		m.updateCollectionStats(time.Since(start))
	}()

	for _, metric := range metrics {
		m.addMetric(metric)

		// 更新统计
		switch metric.GetType() {
		case MetricTypeCounter:
			m.stats.CounterMetrics++
		case MetricTypeGauge:
			m.stats.GaugeMetrics++
		case MetricTypeHistogram:
			m.stats.HistogramMetrics++
		case MetricTypeSummary:
			m.stats.SummaryMetrics++
		}
	}

	return nil
}

// GetMetrics 获取所有指标
func (m *metricCollectorImpl) GetMetrics() []Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回副本
	metrics := make([]Metric, len(m.metrics))
	copy(metrics, m.metrics)
	return metrics
}

// GetMetricsByName 按名称获取指标
func (m *metricCollectorImpl) GetMetricsByName(name string) []Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Metric
	for _, metric := range m.metrics {
		if metric.GetName() == name {
			result = append(result, metric)
		}
	}
	return result
}

// GetMetricsByType 按类型获取指标
func (m *metricCollectorImpl) GetMetricsByType(metricType MetricType) []Metric {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var result []Metric
	for _, metric := range m.metrics {
		if metric.GetType() == metricType {
			result = append(result, metric)
		}
	}
	return result
}

// GetStats 获取统计信息
func (m *metricCollectorImpl) GetStats() *MetricStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 复制统计信息
	stats := *m.stats
	stats.LastUpdate = time.Now()
	stats.TotalMetrics = int64(len(m.metrics))

	// 计算收集率
	if time.Since(stats.StartTime) > 0 {
		stats.CollectionRate = float64(len(m.collectionTimes)) / time.Since(stats.StartTime).Seconds()
	}

	return &stats
}

// ResetStats 重置统计信息
func (m *metricCollectorImpl) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.stats = &MetricStats{StartTime: time.Now()}
	m.collectionTimes = m.collectionTimes[:0]
}

// addMetric 添加指标
func (m *metricCollectorImpl) addMetric(metric Metric) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.metrics = append(m.metrics, metric)
}

// updateCollectionStats 更新收集统计
func (m *metricCollectorImpl) updateCollectionStats(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.collectionTimes = append(m.collectionTimes, duration)

	// 更新性能统计
	if m.stats.AvgCollectionTime == 0 {
		m.stats.AvgCollectionTime = duration
	} else {
		m.stats.AvgCollectionTime = (m.stats.AvgCollectionTime + duration) / 2
	}

	if duration > m.stats.MaxCollectionTime {
		m.stats.MaxCollectionTime = duration
	}

	if m.stats.MinCollectionTime == 0 || duration < m.stats.MinCollectionTime {
		m.stats.MinCollectionTime = duration
	}
}

// PerformanceMetricCollector 性能指标收集器
type PerformanceMetricCollector struct {
	collector MetricCollector
	logger    *zap.Logger
}

// NewPerformanceMetricCollector 创建性能指标收集器
func NewPerformanceMetricCollector(logger *zap.Logger) *PerformanceMetricCollector {
	return &PerformanceMetricCollector{
		collector: NewMetricCollector(logger),
		logger:    logger,
	}
}

// Start 启动性能指标收集器
func (p *PerformanceMetricCollector) Start() error {
	return p.collector.Start()
}

// Stop 停止性能指标收集器
func (p *PerformanceMetricCollector) Stop() error {
	return p.collector.Stop()
}

// IsRunning 检查是否正在运行
func (p *PerformanceMetricCollector) IsRunning() bool {
	return p.collector.IsRunning()
}

// CollectPerformanceMetrics 收集性能指标
func (p *PerformanceMetricCollector) CollectPerformanceMetrics(metrics *PerformanceMetrics) error {
	labels := map[string]string{
		"service":   "data_collection",
		"timestamp": metrics.Timestamp.Format(time.RFC3339),
	}

	// 收集成功率指标
	err := p.collector.CollectGauge("success_rate", metrics.SuccessRate, labels)
	if err != nil {
		return err
	}

	// 收集错误率指标
	err = p.collector.CollectGauge("error_rate", metrics.ErrorRate, labels)
	if err != nil {
		return err
	}

	// 收集重试率指标
	err = p.collector.CollectGauge("retry_rate", metrics.RetryRate, labels)
	if err != nil {
		return err
	}

	// 收集延迟指标
	err = p.collector.CollectHistogram("avg_latency_seconds", metrics.AvgLatency.Seconds(), labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectHistogram("p50_latency_seconds", metrics.P50Latency.Seconds(), labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectHistogram("p95_latency_seconds", metrics.P95Latency.Seconds(), labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectHistogram("p99_latency_seconds", metrics.P99Latency.Seconds(), labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectHistogram("max_latency_seconds", metrics.MaxLatency.Seconds(), labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectHistogram("min_latency_seconds", metrics.MinLatency.Seconds(), labels)
	if err != nil {
		return err
	}

	// 收集吞吐量指标
	err = p.collector.CollectGauge("throughput", metrics.Throughput, labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectGauge("qps", metrics.QPS, labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectGauge("tps", metrics.TPS, labels)
	if err != nil {
		return err
	}

	// 收集资源使用指标
	err = p.collector.CollectGauge("memory_usage_bytes", float64(metrics.MemoryUsage), labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectGauge("cpu_usage", metrics.CPUUsage, labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectGauge("queue_size", float64(metrics.QueueSize), labels)
	if err != nil {
		return err
	}

	err = p.collector.CollectGauge("active_connections", float64(metrics.ActiveConnections), labels)
	if err != nil {
		return err
	}

	return nil
}

// GetMetrics 获取指标
func (p *PerformanceMetricCollector) GetMetrics() []Metric {
	return p.collector.GetMetrics()
}

// GetMetricsByName 按名称获取指标
func (p *PerformanceMetricCollector) GetMetricsByName(name string) []Metric {
	return p.collector.GetMetricsByName(name)
}

// GetStats 获取统计信息
func (p *PerformanceMetricCollector) GetStats() *MetricStats {
	return p.collector.GetStats()
}
