package data_collection

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

// PrometheusExporter Prometheus指标导出器
type PrometheusExporter struct {
	collector MetricCollector
	logger    *zap.Logger
	server    *http.Server
	mu        sync.RWMutex
}

// NewPrometheusExporter 创建Prometheus导出器
func NewPrometheusExporter(collector MetricCollector, logger *zap.Logger) *PrometheusExporter {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &PrometheusExporter{
		collector: collector,
		logger:    logger,
	}
}

// Start 启动Prometheus导出器
func (p *PrometheusExporter) Start(addr string) error {
	if p.server != nil {
		return fmt.Errorf("Prometheus导出器已经在运行")
	}

	// 创建HTTP服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/metrics", p.handleMetrics)
	mux.HandleFunc("/health", p.handleHealth)

	p.server = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 启动服务器
	go func() {
		p.logger.Info("Prometheus导出器启动", zap.String("addr", addr))
		if err := p.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			p.logger.Error("Prometheus导出器启动失败", zap.Error(err))
		}
	}()

	return nil
}

// Stop 停止Prometheus导出器
func (p *PrometheusExporter) Stop(ctx context.Context) error {
	if p.server == nil {
		return fmt.Errorf("Prometheus导出器未运行")
	}

	p.logger.Info("Prometheus导出器停止")
	return p.server.Shutdown(ctx)
}

// handleMetrics 处理/metrics请求
func (p *PrometheusExporter) handleMetrics(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	// 获取所有指标
	metrics := p.collector.GetMetrics()

	// 生成Prometheus格式的指标
	output := p.generatePrometheusMetrics(metrics)

	// 写入响应
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(output))
}

// handleHealth 处理/health请求
func (p *PrometheusExporter) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"status":"healthy"}`))
}

// generatePrometheusMetrics 生成Prometheus格式的指标
func (p *PrometheusExporter) generatePrometheusMetrics(metrics []Metric) string {
	var output strings.Builder

	// 按类型分组指标
	groupedMetrics := p.groupMetricsByType(metrics)

	// 生成计数器指标
	if counterMetrics, exists := groupedMetrics[MetricTypeCounter]; exists {
		output.WriteString("# TYPE counter_total counter\n")
		for _, metric := range counterMetrics {
			line := p.formatPrometheusMetric(metric, "_total")
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// 生成仪表盘指标
	if gaugeMetrics, exists := groupedMetrics[MetricTypeGauge]; exists {
		output.WriteString("# TYPE gauge gauge\n")
		for _, metric := range gaugeMetrics {
			line := p.formatPrometheusMetric(metric, "")
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// 生成直方图指标
	if histogramMetrics, exists := groupedMetrics[MetricTypeHistogram]; exists {
		output.WriteString("# TYPE histogram_seconds histogram\n")
		for _, metric := range histogramMetrics {
			line := p.formatPrometheusMetric(metric, "_seconds")
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	// 生成摘要指标
	if summaryMetrics, exists := groupedMetrics[MetricTypeSummary]; exists {
		output.WriteString("# TYPE summary_seconds summary\n")
		for _, metric := range summaryMetrics {
			line := p.formatPrometheusMetric(metric, "_seconds")
			output.WriteString(line)
			output.WriteString("\n")
		}
	}

	return output.String()
}

// groupMetricsByType 按类型分组指标
func (p *PrometheusExporter) groupMetricsByType(metrics []Metric) map[MetricType][]Metric {
	groups := make(map[MetricType][]Metric)

	for _, metric := range metrics {
		metricType := metric.GetType()
		groups[metricType] = append(groups[metricType], metric)
	}

	return groups
}

// formatPrometheusMetric 格式化Prometheus指标
func (p *PrometheusExporter) formatPrometheusMetric(metric Metric, suffix string) string {
	var line strings.Builder

	// 指标名称
	name := metric.GetName()
	if suffix != "" {
		name += suffix
	}
	line.WriteString(name)

	// 标签
	labels := metric.GetLabels()
	if len(labels) > 0 {
		line.WriteString("{")
		labelPairs := make([]string, 0, len(labels))
		for k, v := range labels {
			labelPairs = append(labelPairs, fmt.Sprintf("%s=%q", k, v))
		}
		line.WriteString(strings.Join(labelPairs, ","))
		line.WriteString("}")
	}

	// 值
	line.WriteString(" ")
	line.WriteString(strconv.FormatFloat(metric.GetValue(), 'f', -1, 64))

	// 时间戳
	timestamp := metric.GetTimestamp().UnixMilli()
	line.WriteString(" ")
	line.WriteString(strconv.FormatInt(timestamp, 10))

	return line.String()
}

// PrometheusConfig Prometheus配置
type PrometheusConfig struct {
	// 服务器配置
	Addr         string        `json:"addr" yaml:"addr"`
	ReadTimeout  time.Duration `json:"read_timeout" yaml:"read_timeout"`
	WriteTimeout time.Duration `json:"write_timeout" yaml:"write_timeout"`

	// 指标配置
	EnableHistograms bool `json:"enable_histograms" yaml:"enable_histograms"`
	EnableSummaries  bool `json:"enable_summaries" yaml:"enable_summaries"`

	// 标签配置
	DefaultLabels map[string]string `json:"default_labels" yaml:"default_labels"`

	// 过滤配置
	IncludeMetrics []string `json:"include_metrics" yaml:"include_metrics"`
	ExcludeMetrics []string `json:"exclude_metrics" yaml:"exclude_metrics"`
}

// DefaultPrometheusConfig 创建默认Prometheus配置
func DefaultPrometheusConfig() *PrometheusConfig {
	return &PrometheusConfig{
		Addr:             ":8080",
		ReadTimeout:      30 * time.Second,
		WriteTimeout:     30 * time.Second,
		EnableHistograms: true,
		EnableSummaries:  true,
		DefaultLabels: map[string]string{
			"service": "data_collection",
		},
		IncludeMetrics: []string{},
		ExcludeMetrics: []string{},
	}
}

// PrometheusMetricFilter Prometheus指标过滤器
type PrometheusMetricFilter struct {
	config *PrometheusConfig
}

// NewPrometheusMetricFilter 创建Prometheus指标过滤器
func NewPrometheusMetricFilter(config *PrometheusConfig) *PrometheusMetricFilter {
	return &PrometheusMetricFilter{
		config: config,
	}
}

// FilterMetrics 过滤指标
func (f *PrometheusMetricFilter) FilterMetrics(metrics []Metric) []Metric {
	var filtered []Metric

	for _, metric := range metrics {
		// 检查包含列表
		if len(f.config.IncludeMetrics) > 0 {
			include := false
			for _, pattern := range f.config.IncludeMetrics {
				if strings.Contains(metric.GetName(), pattern) {
					include = true
					break
				}
			}
			if !include {
				continue
			}
		}

		// 检查排除列表
		exclude := false
		for _, pattern := range f.config.ExcludeMetrics {
			if strings.Contains(metric.GetName(), pattern) {
				exclude = true
				break
			}
		}
		if exclude {
			continue
		}

		// 检查类型过滤
		if !f.config.EnableHistograms && metric.GetType() == MetricTypeHistogram {
			continue
		}
		if !f.config.EnableSummaries && metric.GetType() == MetricTypeSummary {
			continue
		}

		filtered = append(filtered, metric)
	}

	return filtered
}

// AddDefaultLabels 添加默认标签
func (f *PrometheusMetricFilter) AddDefaultLabels(metric Metric) Metric {
	labels := metric.GetLabels()
	if labels == nil {
		labels = make(map[string]string)
	}

	// 添加默认标签
	for k, v := range f.config.DefaultLabels {
		if _, exists := labels[k]; !exists {
			labels[k] = v
		}
	}

	// 根据指标类型创建新的指标
	switch metric.GetType() {
	case MetricTypeCounter:
		return &CounterMetric{
			Name:      metric.GetName(),
			Value:     metric.GetValue(),
			Labels:    labels,
			Timestamp: metric.GetTimestamp(),
		}
	case MetricTypeGauge:
		return &GaugeMetric{
			Name:      metric.GetName(),
			Value:     metric.GetValue(),
			Labels:    labels,
			Timestamp: metric.GetTimestamp(),
		}
	case MetricTypeHistogram:
		if histMetric, ok := metric.(*HistogramMetric); ok {
			return &HistogramMetric{
				Name:      histMetric.Name,
				Value:     histMetric.Value,
				Count:     histMetric.Count,
				Sum:       histMetric.Sum,
				Labels:    labels,
				Timestamp: histMetric.Timestamp,
			}
		}
	case MetricTypeSummary:
		if sumMetric, ok := metric.(*SummaryMetric); ok {
			return &SummaryMetric{
				Name:      sumMetric.Name,
				Value:     sumMetric.Value,
				Count:     sumMetric.Count,
				Sum:       sumMetric.Sum,
				Labels:    labels,
				Timestamp: sumMetric.Timestamp,
			}
		}
	}

	return metric
}
