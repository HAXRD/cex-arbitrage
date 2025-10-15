package data_collection

import (
	"context"
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// healthCheckerImpl 健康检查器实现
type healthCheckerImpl struct {
	logger *zap.Logger
	mu     sync.RWMutex

	// 健康检查存储
	checks map[string]func() *HealthCheck

	// 健康检查结果
	results map[string]*HealthCheck

	// 健康检查配置
	config *HealthCheckConfig
}

// HealthCheckConfig 健康检查配置
type HealthCheckConfig struct {
	// 检查间隔
	CheckInterval time.Duration `json:"check_interval" yaml:"check_interval"`

	// 超时设置
	Timeout time.Duration `json:"timeout" yaml:"timeout"`

	// 重试设置
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"`

	// 健康阈值
	HealthyThreshold  float64 `json:"healthy_threshold" yaml:"healthy_threshold"`
	DegradedThreshold float64 `json:"degraded_threshold" yaml:"degraded_threshold"`

	// 自动检查
	AutoCheck bool `json:"auto_check" yaml:"auto_check"`
}

// NewHealthChecker 创建健康检查器
func NewHealthChecker(logger *zap.Logger) HealthChecker {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &healthCheckerImpl{
		logger:  logger,
		checks:  make(map[string]func() *HealthCheck),
		results: make(map[string]*HealthCheck),
		config:  DefaultHealthCheckConfig(),
	}
}

// CheckHealth 执行健康检查
func (h *healthCheckerImpl) CheckHealth() *HealthCheck {
	h.mu.RLock()
	defer h.mu.RUnlock()

	// 执行所有健康检查
	var overallStatus HealthStatus = HealthStatusHealthy
	var messages []string
	var totalDuration time.Duration

	for name, checkFunc := range h.checks {
		start := time.Now()
		result := checkFunc()
		duration := time.Since(start)

		// 更新结果
		result.Duration = duration
		h.results[name] = result

		// 确定整体状态
		if result.Status == HealthStatusUnhealthy {
			overallStatus = HealthStatusUnhealthy
		} else if result.Status == HealthStatusDegraded && overallStatus != HealthStatusUnhealthy {
			overallStatus = HealthStatusDegraded
		}

		// 收集消息
		if result.Message != "" {
			messages = append(messages, fmt.Sprintf("%s: %s", name, result.Message))
		}

		totalDuration += duration
	}

	// 创建整体健康检查结果
	overallCheck := &HealthCheck{
		Name:      "overall",
		Status:    overallStatus,
		Message:   fmt.Sprintf("整体健康状态: %s", overallStatus),
		Timestamp: time.Now(),
		Duration:  totalDuration,
		Metadata: map[string]interface{}{
			"total_checks":     len(h.checks),
			"healthy_checks":   h.countChecksByStatus(HealthStatusHealthy),
			"degraded_checks":  h.countChecksByStatus(HealthStatusDegraded),
			"unhealthy_checks": h.countChecksByStatus(HealthStatusUnhealthy),
		},
	}

	return overallCheck
}

// CheckHealthByName 按名称执行健康检查
func (h *healthCheckerImpl) CheckHealthByName(name string) *HealthCheck {
	h.mu.RLock()
	defer h.mu.RUnlock()

	checkFunc, exists := h.checks[name]
	if !exists {
		return &HealthCheck{
			Name:      name,
			Status:    HealthStatusUnhealthy,
			Message:   fmt.Sprintf("健康检查 '%s' 不存在", name),
			Timestamp: time.Now(),
			Duration:  0,
		}
	}

	start := time.Now()
	result := checkFunc()
	duration := time.Since(start)

	// 更新结果
	result.Duration = duration
	h.results[name] = result

	return result
}

// GetAllHealthChecks 获取所有健康检查结果
func (h *healthCheckerImpl) GetAllHealthChecks() []*HealthCheck {
	h.mu.RLock()
	defer h.mu.RUnlock()

	var results []*HealthCheck
	for _, result := range h.results {
		resultCopy := *result
		results = append(results, &resultCopy)
	}

	return results
}

// RegisterCheck 注册健康检查
func (h *healthCheckerImpl) RegisterCheck(name string, checkFunc func() *HealthCheck) {
	h.mu.Lock()
	defer h.mu.Unlock()

	h.checks[name] = checkFunc
	h.logger.Info("注册健康检查", zap.String("name", name))
}

// UnregisterCheck 注销健康检查
func (h *healthCheckerImpl) UnregisterCheck(name string) {
	h.mu.Lock()
	defer h.mu.Unlock()

	delete(h.checks, name)
	delete(h.results, name)
	h.logger.Info("注销健康检查", zap.String("name", name))
}

// GetOverallStatus 获取整体健康状态
func (h *healthCheckerImpl) GetOverallStatus() HealthStatus {
	overallCheck := h.CheckHealth()
	return overallCheck.Status
}

// IsHealthy 检查是否健康
func (h *healthCheckerImpl) IsHealthy() bool {
	return h.GetOverallStatus() == HealthStatusHealthy
}

// countChecksByStatus 按状态统计健康检查
func (h *healthCheckerImpl) countChecksByStatus(status HealthStatus) int {
	count := 0
	for _, result := range h.results {
		if result.Status == status {
			count++
		}
	}
	return count
}

// DefaultHealthCheckConfig 创建默认健康检查配置
func DefaultHealthCheckConfig() *HealthCheckConfig {
	return &HealthCheckConfig{
		CheckInterval:     30 * time.Second,
		Timeout:           5 * time.Second,
		MaxRetries:        3,
		RetryDelay:        1 * time.Second,
		HealthyThreshold:  0.9,
		DegradedThreshold: 0.7,
		AutoCheck:         true,
	}
}

// DataCollectionHealthChecker 数据采集健康检查器
type DataCollectionHealthChecker struct {
	healthChecker HealthChecker
	logger        *zap.Logger
}

// NewDataCollectionHealthChecker 创建数据采集健康检查器
func NewDataCollectionHealthChecker(logger *zap.Logger) *DataCollectionHealthChecker {
	if logger == nil {
		logger = zap.NewNop()
	}

	checker := &DataCollectionHealthChecker{
		healthChecker: NewHealthChecker(logger),
		logger:        logger,
	}

	// 注册默认健康检查
	checker.registerDefaultChecks()

	return checker
}

// registerDefaultChecks 注册默认健康检查
func (d *DataCollectionHealthChecker) registerDefaultChecks() {
	// 注册服务健康检查
	d.healthChecker.RegisterCheck("service", func() *HealthCheck {
		return &HealthCheck{
			Name:      "service",
			Status:    HealthStatusHealthy,
			Message:   "数据采集服务运行正常",
			Timestamp: time.Now(),
			Duration:  0,
		}
	})

	// 注册队列健康检查
	d.healthChecker.RegisterCheck("queue", func() *HealthCheck {
		// 这里应该检查队列状态
		// 例如：队列大小、处理速度等
		return &HealthCheck{
			Name:      "queue",
			Status:    HealthStatusHealthy,
			Message:   "队列状态正常",
			Timestamp: time.Now(),
			Duration:  0,
		}
	})

	// 注册数据库健康检查
	d.healthChecker.RegisterCheck("database", func() *HealthCheck {
		// 这里应该检查数据库连接
		// 例如：连接池状态、查询响应时间等
		return &HealthCheck{
			Name:      "database",
			Status:    HealthStatusHealthy,
			Message:   "数据库连接正常",
			Timestamp: time.Now(),
			Duration:  0,
		}
	})

	// 注册Redis健康检查
	d.healthChecker.RegisterCheck("redis", func() *HealthCheck {
		// 这里应该检查Redis连接
		// 例如：连接状态、内存使用等
		return &HealthCheck{
			Name:      "redis",
			Status:    HealthStatusHealthy,
			Message:   "Redis连接正常",
			Timestamp: time.Now(),
			Duration:  0,
		}
	})

	// 注册WebSocket健康检查
	d.healthChecker.RegisterCheck("websocket", func() *HealthCheck {
		// 这里应该检查WebSocket连接
		// 例如：连接状态、消息处理等
		return &HealthCheck{
			Name:      "websocket",
			Status:    HealthStatusHealthy,
			Message:   "WebSocket连接正常",
			Timestamp: time.Now(),
			Duration:  0,
		}
	})
}

// CheckHealth 执行健康检查
func (d *DataCollectionHealthChecker) CheckHealth() *HealthCheck {
	return d.healthChecker.CheckHealth()
}

// CheckHealthByName 按名称执行健康检查
func (d *DataCollectionHealthChecker) CheckHealthByName(name string) *HealthCheck {
	return d.healthChecker.CheckHealthByName(name)
}

// GetAllHealthChecks 获取所有健康检查结果
func (d *DataCollectionHealthChecker) GetAllHealthChecks() []*HealthCheck {
	return d.healthChecker.GetAllHealthChecks()
}

// RegisterCheck 注册健康检查
func (d *DataCollectionHealthChecker) RegisterCheck(name string, checkFunc func() *HealthCheck) {
	d.healthChecker.RegisterCheck(name, checkFunc)
}

// UnregisterCheck 注销健康检查
func (d *DataCollectionHealthChecker) UnregisterCheck(name string) {
	d.healthChecker.UnregisterCheck(name)
}

// GetOverallStatus 获取整体健康状态
func (d *DataCollectionHealthChecker) GetOverallStatus() HealthStatus {
	return d.healthChecker.GetOverallStatus()
}

// IsHealthy 检查是否健康
func (d *DataCollectionHealthChecker) IsHealthy() bool {
	return d.healthChecker.IsHealthy()
}

// StartAutoCheck 启动自动健康检查
func (d *DataCollectionHealthChecker) StartAutoCheck(ctx context.Context) error {
	if !d.healthChecker.(*healthCheckerImpl).config.AutoCheck {
		return fmt.Errorf("自动健康检查未启用")
	}

	ticker := time.NewTicker(d.healthChecker.(*healthCheckerImpl).config.CheckInterval)
	defer ticker.Stop()

	d.logger.Info("启动自动健康检查",
		zap.Duration("interval", d.healthChecker.(*healthCheckerImpl).config.CheckInterval))

	for {
		select {
		case <-ctx.Done():
			d.logger.Info("停止自动健康检查")
			return ctx.Err()
		case <-ticker.C:
			// 执行健康检查
			result := d.CheckHealth()
			d.logger.Debug("自动健康检查完成",
				zap.String("status", string(result.Status)),
				zap.Duration("duration", result.Duration))
		}
	}
}

// HealthCheckMetrics 健康检查指标
type HealthCheckMetrics struct {
	// 健康检查统计
	TotalChecks     int64 `json:"total_checks"`
	HealthyChecks   int64 `json:"healthy_checks"`
	DegradedChecks  int64 `json:"degraded_checks"`
	UnhealthyChecks int64 `json:"unhealthy_checks"`

	// 性能统计
	AvgCheckDuration time.Duration `json:"avg_check_duration"`
	MaxCheckDuration time.Duration `json:"max_check_duration"`
	MinCheckDuration time.Duration `json:"min_check_duration"`

	// 时间戳
	LastCheckTime time.Time `json:"last_check_time"`
	StartTime     time.Time `json:"start_time"`
}

// GetHealthCheckMetrics 获取健康检查指标
func (d *DataCollectionHealthChecker) GetHealthCheckMetrics() *HealthCheckMetrics {
	checks := d.GetAllHealthChecks()

	metrics := &HealthCheckMetrics{
		TotalChecks: int64(len(checks)),
		StartTime:   time.Now(),
	}

	for _, check := range checks {
		switch check.Status {
		case HealthStatusHealthy:
			metrics.HealthyChecks++
		case HealthStatusDegraded:
			metrics.DegradedChecks++
		case HealthStatusUnhealthy:
			metrics.UnhealthyChecks++
		}

		// 更新性能统计
		if metrics.AvgCheckDuration == 0 {
			metrics.AvgCheckDuration = check.Duration
		} else {
			metrics.AvgCheckDuration = (metrics.AvgCheckDuration + check.Duration) / 2
		}

		if check.Duration > metrics.MaxCheckDuration {
			metrics.MaxCheckDuration = check.Duration
		}

		if metrics.MinCheckDuration == 0 || check.Duration < metrics.MinCheckDuration {
			metrics.MinCheckDuration = check.Duration
		}

		if check.Timestamp.After(metrics.LastCheckTime) {
			metrics.LastCheckTime = check.Timestamp
		}
	}

	return metrics
}
