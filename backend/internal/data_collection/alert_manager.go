package data_collection

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
)

// alertManagerImpl 告警管理器实现
type alertManagerImpl struct {
	logger *zap.Logger
	mu     sync.RWMutex

	// 告警存储
	alerts map[string]*Alert
	rules  map[string]*AlertRule

	// 告警统计
	stats *AlertStats
}

// AlertStats 告警统计
type AlertStats struct {
	TotalAlerts    int64 `json:"total_alerts"`
	ActiveAlerts   int64 `json:"active_alerts"`
	ResolvedAlerts int64 `json:"resolved_alerts"`

	// 按级别统计
	InfoAlerts      int64 `json:"info_alerts"`
	WarningAlerts   int64 `json:"warning_alerts"`
	CriticalAlerts  int64 `json:"critical_alerts"`
	EmergencyAlerts int64 `json:"emergency_alerts"`

	// 按来源统计
	SourceStats map[string]int64 `json:"source_stats"`

	// 时间戳
	StartTime  time.Time `json:"start_time"`
	LastUpdate time.Time `json:"last_update"`
}

// NewAlertManager 创建告警管理器
func NewAlertManager(logger *zap.Logger) AlertManager {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &alertManagerImpl{
		logger: logger,
		alerts: make(map[string]*Alert),
		rules:  make(map[string]*AlertRule),
		stats: &AlertStats{
			StartTime:   time.Now(),
			SourceStats: make(map[string]int64),
		},
	}
}

// CreateAlert 创建告警
func (a *alertManagerImpl) CreateAlert(alert *Alert) error {
	if alert == nil {
		return fmt.Errorf("告警不能为空")
	}

	if alert.ID == "" {
		alert.ID = fmt.Sprintf("alert_%d", time.Now().UnixNano())
	}

	if alert.Timestamp.IsZero() {
		alert.Timestamp = time.Now()
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 检查是否已存在
	if _, exists := a.alerts[alert.ID]; exists {
		return fmt.Errorf("告警ID已存在: %s", alert.ID)
	}

	// 添加告警
	a.alerts[alert.ID] = alert

	// 更新统计
	a.updateStats(alert)

	a.logger.Info("创建告警",
		zap.String("id", alert.ID),
		zap.String("level", string(alert.Level)),
		zap.String("title", alert.Title),
		zap.String("source", alert.Source))

	return nil
}

// ResolveAlert 解决告警
func (a *alertManagerImpl) ResolveAlert(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	alert, exists := a.alerts[id]
	if !exists {
		return fmt.Errorf("告警不存在: %s", id)
	}

	if alert.Resolved {
		return fmt.Errorf("告警已解决: %s", id)
	}

	// 解决告警
	now := time.Now()
	alert.Resolved = true
	alert.ResolvedAt = &now

	// 更新统计
	a.stats.ResolvedAlerts++
	a.stats.ActiveAlerts--

	a.logger.Info("解决告警",
		zap.String("id", alert.ID),
		zap.String("level", string(alert.Level)),
		zap.String("title", alert.Title))

	return nil
}

// GetAlert 获取告警
func (a *alertManagerImpl) GetAlert(id string) (*Alert, error) {
	a.mu.RLock()
	defer a.mu.RUnlock()

	alert, exists := a.alerts[id]
	if !exists {
		return nil, fmt.Errorf("告警不存在: %s", id)
	}

	// 返回副本
	alertCopy := *alert
	return &alertCopy, nil
}

// GetAlerts 获取所有告警
func (a *alertManagerImpl) GetAlerts() []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	alerts := make([]*Alert, 0, len(a.alerts))
	for _, alert := range a.alerts {
		alertCopy := *alert
		alerts = append(alerts, &alertCopy)
	}

	return alerts
}

// GetActiveAlerts 获取活跃告警
func (a *alertManagerImpl) GetActiveAlerts() []*Alert {
	a.mu.RLock()
	defer a.mu.RUnlock()

	var activeAlerts []*Alert
	for _, alert := range a.alerts {
		if !alert.Resolved {
			alertCopy := *alert
			activeAlerts = append(activeAlerts, &alertCopy)
		}
	}

	return activeAlerts
}

// AddRule 添加告警规则
func (a *alertManagerImpl) AddRule(rule *AlertRule) error {
	if rule == nil {
		return fmt.Errorf("告警规则不能为空")
	}

	if rule.ID == "" {
		rule.ID = fmt.Sprintf("rule_%d", time.Now().UnixNano())
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// 检查是否已存在
	if _, exists := a.rules[rule.ID]; exists {
		return fmt.Errorf("告警规则ID已存在: %s", rule.ID)
	}

	// 添加规则
	a.rules[rule.ID] = rule

	a.logger.Info("添加告警规则",
		zap.String("id", rule.ID),
		zap.String("name", rule.Name),
		zap.String("condition", rule.Condition),
		zap.Float64("threshold", rule.Threshold))

	return nil
}

// RemoveRule 移除告警规则
func (a *alertManagerImpl) RemoveRule(id string) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	rule, exists := a.rules[id]
	if !exists {
		return fmt.Errorf("告警规则不存在: %s", id)
	}

	delete(a.rules, id)

	a.logger.Info("移除告警规则",
		zap.String("id", rule.ID),
		zap.String("name", rule.Name))

	return nil
}

// GetRules 获取所有告警规则
func (a *alertManagerImpl) GetRules() []*AlertRule {
	a.mu.RLock()
	defer a.mu.RUnlock()

	rules := make([]*AlertRule, 0, len(a.rules))
	for _, rule := range a.rules {
		ruleCopy := *rule
		rules = append(rules, &ruleCopy)
	}

	return rules
}

// CheckAlerts 检查告警
func (a *alertManagerImpl) CheckAlerts() error {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 这里可以实现告警检查逻辑
	// 例如：检查指标是否超过阈值、检查服务健康状态等

	for _, rule := range a.rules {
		if !rule.Enabled {
			continue
		}

		// 这里应该实现具体的告警检查逻辑
		// 例如：检查指标值、检查服务状态等
		a.checkRule(rule)
	}

	return nil
}

// checkRule 检查告警规则
func (a *alertManagerImpl) checkRule(rule *AlertRule) {
	// 这里实现具体的告警检查逻辑
	// 例如：检查指标值是否超过阈值

	// 模拟检查逻辑
	// 实际实现中，这里应该检查真实的指标值
	if rule.Condition == "threshold_exceeded" {
		// 检查阈值
		// 如果超过阈值，创建告警
		a.createAlertFromRule(rule)
	}
}

// createAlertFromRule 从规则创建告警
func (a *alertManagerImpl) createAlertFromRule(rule *AlertRule) {
	alert := &Alert{
		ID:        fmt.Sprintf("auto_%s_%d", rule.ID, time.Now().UnixNano()),
		Level:     rule.Level,
		Title:     fmt.Sprintf("告警规则触发: %s", rule.Name),
		Message:   fmt.Sprintf("条件 '%s' 超过阈值 %.2f", rule.Condition, rule.Threshold),
		Source:    "alert_manager",
		Timestamp: time.Now(),
		Resolved:  false,
		Metadata: map[string]interface{}{
			"rule_id":   rule.ID,
			"rule_name": rule.Name,
			"condition": rule.Condition,
			"threshold": rule.Threshold,
		},
	}

	// 创建告警（这里需要解锁，因为CreateAlert会加锁）
	a.mu.RUnlock()
	err := a.CreateAlert(alert)
	a.mu.RLock()

	if err != nil {
		a.logger.Error("从规则创建告警失败", zap.Error(err))
	}
}

// updateStats 更新统计信息
func (a *alertManagerImpl) updateStats(alert *Alert) {
	a.stats.TotalAlerts++
	a.stats.ActiveAlerts++
	a.stats.LastUpdate = time.Now()

	// 按级别统计
	switch alert.Level {
	case AlertLevelInfo:
		a.stats.InfoAlerts++
	case AlertLevelWarning:
		a.stats.WarningAlerts++
	case AlertLevelCritical:
		a.stats.CriticalAlerts++
	case AlertLevelEmergency:
		a.stats.EmergencyAlerts++
	}

	// 按来源统计
	a.stats.SourceStats[alert.Source]++
}

// GetStats 获取告警统计
func (a *alertManagerImpl) GetStats() *AlertStats {
	a.mu.RLock()
	defer a.mu.RUnlock()

	// 复制统计信息
	stats := *a.stats
	stats.LastUpdate = time.Now()

	return &stats
}

// AlertConfig 告警配置
type AlertConfig struct {
	// 告警规则
	DefaultRules []*AlertRule `json:"default_rules" yaml:"default_rules"`

	// 告警阈值
	Thresholds map[string]float64 `json:"thresholds" yaml:"thresholds"`

	// 告警间隔
	CheckInterval time.Duration `json:"check_interval" yaml:"check_interval"`

	// 告警抑制
	SuppressionRules []*SuppressionRule `json:"suppression_rules" yaml:"suppression_rules"`

	// 告警通知
	NotificationConfig *NotificationConfig `json:"notification_config" yaml:"notification_config"`
}

// SuppressionRule 抑制规则
type SuppressionRule struct {
	ID        string        `json:"id" yaml:"id"`
	Name      string        `json:"name" yaml:"name"`
	Condition string        `json:"condition" yaml:"condition"`
	Duration  time.Duration `json:"duration" yaml:"duration"`
	Enabled   bool          `json:"enabled" yaml:"enabled"`
}

// NotificationConfig 通知配置
type NotificationConfig struct {
	Enabled    bool     `json:"enabled" yaml:"enabled"`
	Channels   []string `json:"channels" yaml:"channels"`
	Recipients []string `json:"recipients" yaml:"recipients"`
	Template   string   `json:"template" yaml:"template"`
}

// DefaultAlertConfig 创建默认告警配置
func DefaultAlertConfig() *AlertConfig {
	return &AlertConfig{
		DefaultRules: []*AlertRule{
			{
				ID:        "high_error_rate",
				Name:      "高错误率告警",
				Condition: "error_rate > 0.1",
				Threshold: 0.1,
				Duration:  5 * time.Minute,
				Level:     AlertLevelWarning,
				Enabled:   true,
			},
			{
				ID:        "high_latency",
				Name:      "高延迟告警",
				Condition: "avg_latency > 1.0",
				Threshold: 1.0,
				Duration:  3 * time.Minute,
				Level:     AlertLevelCritical,
				Enabled:   true,
			},
		},
		Thresholds: map[string]float64{
			"error_rate":   0.1,
			"avg_latency":  1.0,
			"queue_size":   1000,
			"memory_usage": 0.8,
		},
		CheckInterval: 30 * time.Second,
		SuppressionRules: []*SuppressionRule{
			{
				ID:        "maintenance_window",
				Name:      "维护窗口抑制",
				Condition: "time_between(02:00, 04:00)",
				Duration:  2 * time.Hour,
				Enabled:   true,
			},
		},
		NotificationConfig: &NotificationConfig{
			Enabled:    true,
			Channels:   []string{"email", "slack"},
			Recipients: []string{"admin@example.com"},
			Template:   "告警: {{.Title}} - {{.Message}}",
		},
	}
}
