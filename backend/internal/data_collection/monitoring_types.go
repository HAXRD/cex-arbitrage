package data_collection

import (
	"time"
)

// MetricType 指标类型
type MetricType string

const (
	MetricTypeCounter   MetricType = "counter"
	MetricTypeGauge     MetricType = "gauge"
	MetricTypeHistogram MetricType = "histogram"
	MetricTypeSummary   MetricType = "summary"
)

// Metric 指标接口
type Metric interface {
	GetName() string
	GetType() MetricType
	GetValue() float64
	GetLabels() map[string]string
	GetTimestamp() time.Time
}

// CounterMetric 计数器指标
type CounterMetric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

func (c *CounterMetric) GetName() string              { return c.Name }
func (c *CounterMetric) GetType() MetricType          { return MetricTypeCounter }
func (c *CounterMetric) GetValue() float64            { return c.Value }
func (c *CounterMetric) GetLabels() map[string]string { return c.Labels }
func (c *CounterMetric) GetTimestamp() time.Time      { return c.Timestamp }

// GaugeMetric 仪表盘指标
type GaugeMetric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

func (g *GaugeMetric) GetName() string              { return g.Name }
func (g *GaugeMetric) GetType() MetricType          { return MetricTypeGauge }
func (g *GaugeMetric) GetValue() float64            { return g.Value }
func (g *GaugeMetric) GetLabels() map[string]string { return g.Labels }
func (g *GaugeMetric) GetTimestamp() time.Time      { return g.Timestamp }

// HistogramMetric 直方图指标
type HistogramMetric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Count     uint64            `json:"count"`
	Sum       float64           `json:"sum"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

func (h *HistogramMetric) GetName() string              { return h.Name }
func (h *HistogramMetric) GetType() MetricType          { return MetricTypeHistogram }
func (h *HistogramMetric) GetValue() float64            { return h.Value }
func (h *HistogramMetric) GetLabels() map[string]string { return h.Labels }
func (h *HistogramMetric) GetTimestamp() time.Time      { return h.Timestamp }

// SummaryMetric 摘要指标
type SummaryMetric struct {
	Name      string            `json:"name"`
	Value     float64           `json:"value"`
	Count     uint64            `json:"count"`
	Sum       float64           `json:"sum"`
	Labels    map[string]string `json:"labels"`
	Timestamp time.Time         `json:"timestamp"`
}

func (s *SummaryMetric) GetName() string              { return s.Name }
func (s *SummaryMetric) GetType() MetricType          { return MetricTypeSummary }
func (s *SummaryMetric) GetValue() float64            { return s.Value }
func (s *SummaryMetric) GetLabels() map[string]string { return s.Labels }
func (s *SummaryMetric) GetTimestamp() time.Time      { return s.Timestamp }

// MetricCollector 指标收集器接口
type MetricCollector interface {
	// 基础操作
	Start() error
	Stop() error
	IsRunning() bool

	// 指标收集
	CollectCounter(name string, value float64, labels map[string]string) error
	CollectGauge(name string, value float64, labels map[string]string) error
	CollectHistogram(name string, value float64, labels map[string]string) error
	CollectSummary(name string, value float64, labels map[string]string) error

	// 批量收集
	CollectBatch(metrics []Metric) error

	// 指标查询
	GetMetrics() []Metric
	GetMetricsByName(name string) []Metric
	GetMetricsByType(metricType MetricType) []Metric

	// 统计信息
	GetStats() *MetricStats
	ResetStats()
}

// MetricStats 指标统计
type MetricStats struct {
	TotalMetrics     int64 `json:"total_metrics"`
	CounterMetrics   int64 `json:"counter_metrics"`
	GaugeMetrics     int64 `json:"gauge_metrics"`
	HistogramMetrics int64 `json:"histogram_metrics"`
	SummaryMetrics   int64 `json:"summary_metrics"`

	// 性能统计
	CollectionRate    float64       `json:"collection_rate"` // 每秒收集次数
	AvgCollectionTime time.Duration `json:"avg_collection_time"`
	MaxCollectionTime time.Duration `json:"max_collection_time"`
	MinCollectionTime time.Duration `json:"min_collection_time"`

	// 时间戳
	StartTime  time.Time `json:"start_time"`
	LastUpdate time.Time `json:"last_update"`
}

// PerformanceMetrics 性能指标
type PerformanceMetrics struct {
	// 成功率指标
	SuccessRate float64 `json:"success_rate"`
	ErrorRate   float64 `json:"error_rate"`
	RetryRate   float64 `json:"retry_rate"`

	// 延迟指标
	AvgLatency time.Duration `json:"avg_latency"`
	P50Latency time.Duration `json:"p50_latency"`
	P95Latency time.Duration `json:"p95_latency"`
	P99Latency time.Duration `json:"p99_latency"`
	MaxLatency time.Duration `json:"max_latency"`
	MinLatency time.Duration `json:"min_latency"`

	// 吞吐量指标
	Throughput float64 `json:"throughput"` // 每秒操作数
	QPS        float64 `json:"qps"`        // 每秒查询数
	TPS        float64 `json:"tps"`        // 每秒事务数

	// 资源使用指标
	MemoryUsage       int64   `json:"memory_usage"`
	CPUUsage          float64 `json:"cpu_usage"`
	QueueSize         int64   `json:"queue_size"`
	ActiveConnections int64   `json:"active_connections"`

	// 时间戳
	Timestamp time.Time `json:"timestamp"`
}

// LogLevel 日志级别
type LogLevel string

const (
	LogLevelDebug LogLevel = "debug"
	LogLevelInfo  LogLevel = "info"
	LogLevelWarn  LogLevel = "warn"
	LogLevelError LogLevel = "error"
	LogLevelFatal LogLevel = "fatal"
)

// LogEntry 日志条目
type LogEntry struct {
	Level     LogLevel               `json:"level"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Fields    map[string]interface{} `json:"fields"`
	Error     error                  `json:"error,omitempty"`
	Stack     string                 `json:"stack,omitempty"`
}

// Logger 日志记录器接口
type Logger interface {
	// 基础日志方法
	Debug(msg string, fields ...map[string]interface{})
	Info(msg string, fields ...map[string]interface{})
	Warn(msg string, fields ...map[string]interface{})
	Error(msg string, err error, fields ...map[string]interface{})
	Fatal(msg string, err error, fields ...map[string]interface{})

	// 结构化日志
	WithFields(fields map[string]interface{}) Logger
	WithError(err error) Logger

	// 日志配置
	SetLevel(level LogLevel)
	GetLevel() LogLevel

	// 日志输出
	GetEntries() []LogEntry
	ClearEntries()
}

// AlertLevel 告警级别
type AlertLevel string

const (
	AlertLevelInfo      AlertLevel = "info"
	AlertLevelWarning   AlertLevel = "warning"
	AlertLevelCritical  AlertLevel = "critical"
	AlertLevelEmergency AlertLevel = "emergency"
)

// Alert 告警
type Alert struct {
	ID         string                 `json:"id"`
	Level      AlertLevel             `json:"level"`
	Title      string                 `json:"title"`
	Message    string                 `json:"message"`
	Source     string                 `json:"source"`
	Timestamp  time.Time              `json:"timestamp"`
	Resolved   bool                   `json:"resolved"`
	ResolvedAt *time.Time             `json:"resolved_at,omitempty"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// AlertManager 告警管理器接口
type AlertManager interface {
	// 告警管理
	CreateAlert(alert *Alert) error
	ResolveAlert(id string) error
	GetAlert(id string) (*Alert, error)
	GetAlerts() []*Alert
	GetActiveAlerts() []*Alert

	// 告警规则
	AddRule(rule *AlertRule) error
	RemoveRule(id string) error
	GetRules() []*AlertRule

	// 告警检查
	CheckAlerts() error

	// 统计信息
	GetStats() *AlertStats
}

// AlertRule 告警规则
type AlertRule struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Condition string                 `json:"condition"`
	Threshold float64                `json:"threshold"`
	Duration  time.Duration          `json:"duration"`
	Level     AlertLevel             `json:"level"`
	Enabled   bool                   `json:"enabled"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// HealthStatus 健康状态
type HealthStatus string

const (
	HealthStatusHealthy   HealthStatus = "healthy"
	HealthStatusDegraded  HealthStatus = "degraded"
	HealthStatusUnhealthy HealthStatus = "unhealthy"
)

// HealthCheck 健康检查
type HealthCheck struct {
	Name      string                 `json:"name"`
	Status    HealthStatus           `json:"status"`
	Message   string                 `json:"message"`
	Timestamp time.Time              `json:"timestamp"`
	Duration  time.Duration          `json:"duration"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// HealthChecker 健康检查器接口
type HealthChecker interface {
	// 健康检查
	CheckHealth() *HealthCheck
	CheckHealthByName(name string) *HealthCheck
	GetAllHealthChecks() []*HealthCheck

	// 健康检查注册
	RegisterCheck(name string, checkFunc func() *HealthCheck)
	UnregisterCheck(name string)

	// 健康状态
	GetOverallStatus() HealthStatus
	IsHealthy() bool
}
