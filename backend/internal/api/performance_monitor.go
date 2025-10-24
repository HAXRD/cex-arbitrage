package api

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// PerformanceMonitor API性能监控器
type PerformanceMonitor struct {
	// 统计数据
	requestCount    atomic.Int64
	responseTimeSum atomic.Int64
	errorCount      atomic.Int64
	activeRequests  atomic.Int64
	
	// 时间窗口统计
	windowSize      time.Duration
	windowData      []*WindowData
	windowIndex     int
	windowMu        sync.RWMutex
	
	// 告警阈值
	latencyThreshold time.Duration
	errorRateThreshold float64
	
	// 日志记录器
	logger *zap.Logger
	
	// 控制
	ctx    context.Context
	cancel context.CancelFunc
}

// WindowData 时间窗口数据
type WindowData struct {
	Timestamp    time.Time
	RequestCount int64
	ResponseTime time.Duration
	ErrorCount   int64
}

// PerformanceStats 性能统计信息
type PerformanceStats struct {
	TotalRequests    int64         `json:"total_requests"`
	AverageLatency   time.Duration `json:"average_latency"`
	ErrorRate        float64       `json:"error_rate"`
	ActiveRequests   int64         `json:"active_requests"`
	RequestsPerSecond float64      `json:"requests_per_second"`
	P95Latency       time.Duration `json:"p95_latency"`
	P99Latency       time.Duration `json:"p99_latency"`
	WindowSize       time.Duration `json:"window_size"`
	LastUpdated      time.Time     `json:"last_updated"`
}

// NewPerformanceMonitor 创建性能监控器
func NewPerformanceMonitor(logger *zap.Logger) *PerformanceMonitor {
	ctx, cancel := context.WithCancel(context.Background())
	
	pm := &PerformanceMonitor{
		windowSize:          60 * time.Second, // 1分钟窗口
		windowData:          make([]*WindowData, 60), // 60个1秒窗口
		latencyThreshold:    200 * time.Millisecond,
		errorRateThreshold:  0.05, // 5%
		logger:              logger,
		ctx:                 ctx,
		cancel:              cancel,
	}
	
	// 启动监控循环
	go pm.monitoringLoop()
	
	return pm
}

// Middleware 性能监控中间件
func (pm *PerformanceMonitor) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		
		// 增加活跃请求计数
		pm.activeRequests.Add(1)
		defer pm.activeRequests.Add(-1)
		
		// 处理请求
		c.Next()
		
		// 记录请求完成
		latency := time.Since(start)
		pm.recordRequest(latency, c.Writer.Status())
	}
}

// recordRequest 记录请求
func (pm *PerformanceMonitor) recordRequest(latency time.Duration, statusCode int) {
	// 更新总计数
	pm.requestCount.Add(1)
	pm.responseTimeSum.Add(int64(latency))
	
	// 记录错误
	if statusCode >= 400 {
		pm.errorCount.Add(1)
	}
	
	// 更新时间窗口数据
	pm.updateWindowData(latency, statusCode >= 400)
}

// updateWindowData 更新时间窗口数据
func (pm *PerformanceMonitor) updateWindowData(latency time.Duration, isError bool) {
	pm.windowMu.Lock()
	defer pm.windowMu.Unlock()
	
	now := time.Now()
	windowIndex := int(now.Unix()) % len(pm.windowData)
	
	// 如果窗口数据过期，重置
	if pm.windowData[windowIndex] == nil || 
		now.Sub(pm.windowData[windowIndex].Timestamp) > time.Second {
		pm.windowData[windowIndex] = &WindowData{
			Timestamp:    now,
			RequestCount: 0,
			ResponseTime: 0,
			ErrorCount:   0,
		}
	}
	
	// 更新窗口数据
	pm.windowData[windowIndex].RequestCount++
	pm.windowData[windowIndex].ResponseTime += latency
	if isError {
		pm.windowData[windowIndex].ErrorCount++
	}
}

// GetStats 获取性能统计信息
func (pm *PerformanceMonitor) GetStats() *PerformanceStats {
	totalRequests := pm.requestCount.Load()
	responseTimeSum := pm.responseTimeSum.Load()
	errorCount := pm.errorCount.Load()
	activeRequests := pm.activeRequests.Load()
	
	var avgLatency time.Duration
	if totalRequests > 0 {
		avgLatency = time.Duration(responseTimeSum / totalRequests)
	}
	
	var errorRate float64
	if totalRequests > 0 {
		errorRate = float64(errorCount) / float64(totalRequests)
	}
	
	// 计算每秒请求数
	requestsPerSecond := pm.calculateRequestsPerSecond()
	
	// 计算P95和P99延迟
	p95Latency, p99Latency := pm.calculatePercentiles()
	
	return &PerformanceStats{
		TotalRequests:     totalRequests,
		AverageLatency:   avgLatency,
		ErrorRate:        errorRate,
		ActiveRequests:   activeRequests,
		RequestsPerSecond: requestsPerSecond,
		P95Latency:       p95Latency,
		P99Latency:       p99Latency,
		WindowSize:       pm.windowSize,
		LastUpdated:      time.Now(),
	}
}

// calculateRequestsPerSecond 计算每秒请求数
func (pm *PerformanceMonitor) calculateRequestsPerSecond() float64 {
	pm.windowMu.RLock()
	defer pm.windowMu.RUnlock()
	
	var totalRequests int64
	var validWindows int
	
	now := time.Now()
	for _, data := range pm.windowData {
		if data != nil && now.Sub(data.Timestamp) <= time.Second {
			totalRequests += data.RequestCount
			validWindows++
		}
	}
	
	if validWindows == 0 {
		return 0
	}
	
	return float64(totalRequests) / float64(validWindows)
}

// calculatePercentiles 计算百分位延迟
func (pm *PerformanceMonitor) calculatePercentiles() (time.Duration, time.Duration) {
	pm.windowMu.RLock()
	defer pm.windowMu.RUnlock()
	
	var latencies []time.Duration
	now := time.Now()
	
	// 收集最近窗口的延迟数据
	for _, data := range pm.windowData {
		if data != nil && now.Sub(data.Timestamp) <= pm.windowSize {
			if data.RequestCount > 0 {
				avgLatency := data.ResponseTime / time.Duration(data.RequestCount)
				latencies = append(latencies, avgLatency)
			}
		}
	}
	
	if len(latencies) == 0 {
		return 0, 0
	}
	
	// 简单排序（实际应用中应该使用更高效的算法）
	for i := 0; i < len(latencies); i++ {
		for j := i + 1; j < len(latencies); j++ {
			if latencies[i] > latencies[j] {
				latencies[i], latencies[j] = latencies[j], latencies[i]
			}
		}
	}
	
	// 计算P95和P99
	p95Index := int(float64(len(latencies)) * 0.95)
	p99Index := int(float64(len(latencies)) * 0.99)
	
	if p95Index >= len(latencies) {
		p95Index = len(latencies) - 1
	}
	if p99Index >= len(latencies) {
		p99Index = len(latencies) - 1
	}
	
	return latencies[p95Index], latencies[p99Index]
}

// monitoringLoop 监控循环
func (pm *PerformanceMonitor) monitoringLoop() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-pm.ctx.Done():
			return
		case <-ticker.C:
			pm.checkAlerts()
		}
	}
}

// checkAlerts 检查告警
func (pm *PerformanceMonitor) checkAlerts() {
	stats := pm.GetStats()
	
	// 检查延迟告警
	if stats.AverageLatency > pm.latencyThreshold {
		pm.logger.Warn("高延迟告警",
			zap.Duration("average_latency", stats.AverageLatency),
			zap.Duration("threshold", pm.latencyThreshold),
		)
	}
	
	// 检查错误率告警
	if stats.ErrorRate > pm.errorRateThreshold {
		pm.logger.Warn("高错误率告警",
			zap.Float64("error_rate", stats.ErrorRate),
			zap.Float64("threshold", pm.errorRateThreshold),
		)
	}
	
	// 记录性能统计
	pm.logger.Info("性能统计",
		zap.Int64("total_requests", stats.TotalRequests),
		zap.Duration("average_latency", stats.AverageLatency),
		zap.Float64("error_rate", stats.ErrorRate),
		zap.Int64("active_requests", stats.ActiveRequests),
		zap.Float64("requests_per_second", stats.RequestsPerSecond),
		zap.Duration("p95_latency", stats.P95Latency),
		zap.Duration("p99_latency", stats.P99Latency),
	)
}

// Reset 重置统计信息
func (pm *PerformanceMonitor) Reset() {
	pm.requestCount.Store(0)
	pm.responseTimeSum.Store(0)
	pm.errorCount.Store(0)
	pm.activeRequests.Store(0)
	
	pm.windowMu.Lock()
	defer pm.windowMu.Unlock()
	
	pm.windowData = make([]*WindowData, len(pm.windowData))
	pm.windowIndex = 0
	
	pm.logger.Info("性能监控统计已重置")
}

// Stop 停止监控
func (pm *PerformanceMonitor) Stop() {
	pm.cancel()
	pm.logger.Info("性能监控已停止")
}

// SetLatencyThreshold 设置延迟阈值
func (pm *PerformanceMonitor) SetLatencyThreshold(threshold time.Duration) {
	pm.latencyThreshold = threshold
	pm.logger.Info("延迟阈值已更新", zap.Duration("threshold", threshold))
}

// SetErrorRateThreshold 设置错误率阈值
func (pm *PerformanceMonitor) SetErrorRateThreshold(threshold float64) {
	pm.errorRateThreshold = threshold
	pm.logger.Info("错误率阈值已更新", zap.Float64("threshold", threshold))
}

// GetHealthStatus 获取健康状态
func (pm *PerformanceMonitor) GetHealthStatus() map[string]interface{} {
	stats := pm.GetStats()
	
	// 健康状态评估
	healthScore := 100.0
	
	// 延迟评分
	if stats.AverageLatency > pm.latencyThreshold {
		healthScore -= 30
	} else if stats.AverageLatency > pm.latencyThreshold/2 {
		healthScore -= 15
	}
	
	// 错误率评分
	if stats.ErrorRate > pm.errorRateThreshold {
		healthScore -= 40
	} else if stats.ErrorRate > pm.errorRateThreshold/2 {
		healthScore -= 20
	}
	
	// 吞吐量评分
	if stats.RequestsPerSecond < 100 {
		healthScore -= 10
	}
	
	// 确定健康状态
	var status string
	if healthScore >= 90 {
		status = "healthy"
	} else if healthScore >= 70 {
		status = "warning"
	} else {
		status = "critical"
	}
	
	return map[string]interface{}{
		"status":           status,
		"health_score":     healthScore,
		"stats":           stats,
		"thresholds": map[string]interface{}{
			"latency_threshold":    pm.latencyThreshold,
			"error_rate_threshold": pm.errorRateThreshold,
		},
	}
}
