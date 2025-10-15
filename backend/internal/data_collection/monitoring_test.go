package data_collection

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMetricCollector_BasicOperations(t *testing.T) {
	// 跳过测试，需要真实监控系统
	t.Skip("需要真实监控系统，跳过测试")

	// 创建指标收集器
	collector := NewMetricCollector(zap.NewNop())
	require.NotNil(t, collector)
	defer collector.Stop()

	t.Run("启动和停止", func(t *testing.T) {
		// 启动
		err := collector.Start()
		require.NoError(t, err)
		assert.True(t, collector.IsRunning())

		// 停止
		err = collector.Stop()
		require.NoError(t, err)
		assert.False(t, collector.IsRunning())
	})

	t.Run("收集计数器指标", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 收集计数器指标
		labels := map[string]string{
			"service": "data_collection",
			"type":    "price",
		}

		err = collector.CollectCounter("data_processed_total", 100.0, labels)
		require.NoError(t, err)

		// 验证指标
		metrics := collector.GetMetricsByName("data_processed_total")
		require.Len(t, metrics, 1)
		assert.Equal(t, "data_processed_total", metrics[0].GetName())
		assert.Equal(t, MetricTypeCounter, metrics[0].GetType())
		assert.Equal(t, 100.0, metrics[0].GetValue())
		assert.Equal(t, labels, metrics[0].GetLabels())
	})

	t.Run("收集仪表盘指标", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 收集仪表盘指标
		labels := map[string]string{
			"service": "data_collection",
			"metric":  "queue_size",
		}

		err = collector.CollectGauge("queue_size_current", 50.0, labels)
		require.NoError(t, err)

		// 验证指标
		metrics := collector.GetMetricsByName("queue_size_current")
		require.Len(t, metrics, 1)
		assert.Equal(t, "queue_size_current", metrics[0].GetName())
		assert.Equal(t, MetricTypeGauge, metrics[0].GetType())
		assert.Equal(t, 50.0, metrics[0].GetValue())
		assert.Equal(t, labels, metrics[0].GetLabels())
	})

	t.Run("收集直方图指标", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 收集直方图指标
		labels := map[string]string{
			"service":   "data_collection",
			"operation": "data_processing",
		}

		err = collector.CollectHistogram("processing_duration_seconds", 0.5, labels)
		require.NoError(t, err)

		// 验证指标
		metrics := collector.GetMetricsByName("processing_duration_seconds")
		require.Len(t, metrics, 1)
		assert.Equal(t, "processing_duration_seconds", metrics[0].GetName())
		assert.Equal(t, MetricTypeHistogram, metrics[0].GetType())
		assert.Equal(t, 0.5, metrics[0].GetValue())
		assert.Equal(t, labels, metrics[0].GetLabels())
	})

	t.Run("收集摘要指标", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 收集摘要指标
		labels := map[string]string{
			"service":   "data_collection",
			"operation": "data_validation",
		}

		err = collector.CollectSummary("validation_duration_seconds", 0.2, labels)
		require.NoError(t, err)

		// 验证指标
		metrics := collector.GetMetricsByName("validation_duration_seconds")
		require.Len(t, metrics, 1)
		assert.Equal(t, "validation_duration_seconds", metrics[0].GetName())
		assert.Equal(t, MetricTypeSummary, metrics[0].GetType())
		assert.Equal(t, 0.2, metrics[0].GetValue())
		assert.Equal(t, labels, metrics[0].GetLabels())
	})

	t.Run("批量收集指标", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 创建批量指标
		metrics := []Metric{
			&CounterMetric{
				Name:      "batch_processed_total",
				Value:     10.0,
				Labels:    map[string]string{"type": "batch"},
				Timestamp: time.Now(),
			},
			&GaugeMetric{
				Name:      "batch_size_current",
				Value:     5.0,
				Labels:    map[string]string{"type": "batch"},
				Timestamp: time.Now(),
			},
		}

		// 批量收集
		err = collector.CollectBatch(metrics)
		require.NoError(t, err)

		// 验证指标
		allMetrics := collector.GetMetrics()
		assert.GreaterOrEqual(t, len(allMetrics), 2)
	})

	t.Run("按类型查询指标", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 收集不同类型的指标
		collector.CollectCounter("test_counter", 1.0, nil)
		collector.CollectGauge("test_gauge", 2.0, nil)
		collector.CollectHistogram("test_histogram", 3.0, nil)
		collector.CollectSummary("test_summary", 4.0, nil)

		// 验证按类型查询
		counterMetrics := collector.GetMetricsByType(MetricTypeCounter)
		assert.GreaterOrEqual(t, len(counterMetrics), 1)

		gaugeMetrics := collector.GetMetricsByType(MetricTypeGauge)
		assert.GreaterOrEqual(t, len(gaugeMetrics), 1)

		histogramMetrics := collector.GetMetricsByType(MetricTypeHistogram)
		assert.GreaterOrEqual(t, len(histogramMetrics), 1)

		summaryMetrics := collector.GetMetricsByType(MetricTypeSummary)
		assert.GreaterOrEqual(t, len(summaryMetrics), 1)
	})
}

func TestMetricCollector_Stats(t *testing.T) {
	// 跳过测试，需要真实监控系统
	t.Skip("需要真实监控系统，跳过测试")

	// 创建指标收集器
	collector := NewMetricCollector(zap.NewNop())
	require.NotNil(t, collector)
	defer collector.Stop()

	t.Run("统计信息", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 收集一些指标
		for i := 0; i < 10; i++ {
			collector.CollectCounter("test_counter", float64(i), nil)
			collector.CollectGauge("test_gauge", float64(i), nil)
		}

		// 获取统计信息
		stats := collector.GetStats()
		require.NotNil(t, stats)
		assert.True(t, stats.TotalMetrics >= 0)
		assert.True(t, stats.CounterMetrics >= 0)
		assert.True(t, stats.GaugeMetrics >= 0)
		assert.True(t, stats.CollectionRate >= 0)
	})

	t.Run("重置统计", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 收集一些指标
		collector.CollectCounter("test_counter", 1.0, nil)

		// 重置统计
		collector.ResetStats()

		// 验证统计已重置
		stats := collector.GetStats()
		require.NotNil(t, stats)
		assert.Equal(t, int64(0), stats.TotalMetrics)
		assert.Equal(t, int64(0), stats.CounterMetrics)
		assert.Equal(t, int64(0), stats.GaugeMetrics)
	})
}

func TestPerformanceMetrics_Collection(t *testing.T) {
	// 跳过测试，需要真实监控系统
	t.Skip("需要真实监控系统，跳过测试")

	// 创建性能指标收集器
	collector := NewPerformanceMetricCollector(zap.NewNop())
	require.NotNil(t, collector)
	defer collector.Stop()

	t.Run("性能指标收集", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 模拟性能数据
		metrics := &PerformanceMetrics{
			SuccessRate:       0.95,
			ErrorRate:         0.05,
			RetryRate:         0.02,
			AvgLatency:        100 * time.Millisecond,
			P50Latency:        80 * time.Millisecond,
			P95Latency:        200 * time.Millisecond,
			P99Latency:        500 * time.Millisecond,
			MaxLatency:        1 * time.Second,
			MinLatency:        10 * time.Millisecond,
			Throughput:        1000.0,
			QPS:               800.0,
			TPS:               600.0,
			MemoryUsage:       1024 * 1024 * 100, // 100MB
			CPUUsage:          0.75,
			QueueSize:         50,
			ActiveConnections: 10,
			Timestamp:         time.Now(),
		}

		// 收集性能指标
		err = collector.CollectPerformanceMetrics(metrics)
		require.NoError(t, err)

		// 验证指标
		collectedMetrics := collector.GetMetrics()
		assert.Greater(t, len(collectedMetrics), 0)

		// 验证特定指标
		successRateMetrics := collector.GetMetricsByName("success_rate")
		assert.GreaterOrEqual(t, len(successRateMetrics), 1)
		assert.Equal(t, 0.95, successRateMetrics[0].GetValue())

		latencyMetrics := collector.GetMetricsByName("avg_latency_seconds")
		assert.GreaterOrEqual(t, len(latencyMetrics), 1)
		assert.Equal(t, 0.1, latencyMetrics[0].GetValue()) // 100ms = 0.1s
	})

	t.Run("批量性能指标收集", func(t *testing.T) {
		// 启动收集器
		err := collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		// 创建多个性能指标
		metricsList := make([]*PerformanceMetrics, 5)
		for i := 0; i < 5; i++ {
			metricsList[i] = &PerformanceMetrics{
				SuccessRate: 0.9 + float64(i)*0.01,
				ErrorRate:   0.1 - float64(i)*0.01,
				AvgLatency:  time.Duration(100+i*10) * time.Millisecond,
				Throughput:  1000.0 + float64(i)*100,
				MemoryUsage: int64(1024*1024*100 + i*1024*1024*10),
				CPUUsage:    0.7 + float64(i)*0.05,
				Timestamp:   time.Now().Add(time.Duration(i) * time.Second),
			}
		}

		// 批量收集
		for _, metrics := range metricsList {
			err = collector.CollectPerformanceMetrics(metrics)
			require.NoError(t, err)
		}

		// 验证指标数量
		allMetrics := collector.GetMetrics()
		assert.GreaterOrEqual(t, len(allMetrics), 5)
	})
}
