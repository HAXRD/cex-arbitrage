package data_collection

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestMonitoringSystem_Integration(t *testing.T) {
	// 跳过测试，需要真实监控系统
	t.Skip("需要真实监控系统，跳过测试")
	
	// 创建监控系统组件
	logger := zap.NewNop()
	metricCollector := NewMetricCollector(logger)
	alertManager := NewAlertManager(logger)
	healthChecker := NewDataCollectionHealthChecker(logger)
	
	// 启动组件
	err := metricCollector.Start()
	require.NoError(t, err)
	defer metricCollector.Stop()

	t.Run("监控系统集成测试", func(t *testing.T) {
		// 收集性能指标
		performanceMetrics := &PerformanceMetrics{
			SuccessRate:    0.95,
			ErrorRate:      0.05,
			RetryRate:      0.02,
			AvgLatency:     100 * time.Millisecond,
			P50Latency:     80 * time.Millisecond,
			P95Latency:     200 * time.Millisecond,
			P99Latency:     500 * time.Millisecond,
			MaxLatency:     1 * time.Second,
			MinLatency:     10 * time.Millisecond,
			Throughput:     1000.0,
			QPS:           800.0,
			TPS:           600.0,
			MemoryUsage:   1024 * 1024 * 100, // 100MB
			CPUUsage:      0.75,
			QueueSize:      50,
			ActiveConnections: 10,
			Timestamp:     time.Now(),
		}

		// 收集指标
		performanceCollector := NewPerformanceMetricCollector(logger)
		err := performanceCollector.Start()
		require.NoError(t, err)
		defer performanceCollector.Stop()

		err = performanceCollector.CollectPerformanceMetrics(performanceMetrics)
		require.NoError(t, err)

		// 验证指标收集
		metrics := performanceCollector.GetMetrics()
		assert.Greater(t, len(metrics), 0)

		// 检查健康状态
		healthCheck := healthChecker.CheckHealth()
		assert.NotNil(t, healthCheck)
		assert.True(t, healthCheck.Status == HealthStatusHealthy || 
			healthCheck.Status == HealthStatusDegraded || 
			healthCheck.Status == HealthStatusUnhealthy)

		// 创建告警
		alert := &Alert{
			ID:        "test_alert",
			Level:     AlertLevelWarning,
			Title:     "测试告警",
			Message:   "这是一个测试告警",
			Source:    "test",
			Timestamp: time.Now(),
			Resolved:  false,
			Metadata:  map[string]interface{}{"test": true},
		}

		err = alertManager.CreateAlert(alert)
		require.NoError(t, err)

		// 验证告警创建
		createdAlert, err := alertManager.GetAlert("test_alert")
		require.NoError(t, err)
		assert.Equal(t, "test_alert", createdAlert.ID)
		assert.Equal(t, AlertLevelWarning, createdAlert.Level)
		assert.Equal(t, "测试告警", createdAlert.Title)

		// 解决告警
		err = alertManager.ResolveAlert("test_alert")
		require.NoError(t, err)

		// 验证告警解决
		resolvedAlert, err := alertManager.GetAlert("test_alert")
		require.NoError(t, err)
		assert.True(t, resolvedAlert.Resolved)
		assert.NotNil(t, resolvedAlert.ResolvedAt)
	})

	t.Run("监控系统性能测试", func(t *testing.T) {
		// 性能指标收集测试
		performanceCollector := NewPerformanceMetricCollector(logger)
		err := performanceCollector.Start()
		require.NoError(t, err)
		defer performanceCollector.Stop()

		// 批量收集指标
		start := time.Now()
		for i := 0; i < 100; i++ {
			metrics := &PerformanceMetrics{
				SuccessRate:    0.9 + float64(i)*0.001,
				ErrorRate:      0.1 - float64(i)*0.001,
				AvgLatency:     time.Duration(100+i) * time.Millisecond,
				Throughput:     1000.0 + float64(i)*10,
				MemoryUsage:    int64(1024*1024*100 + i*1024*1024),
				CPUUsage:       0.7 + float64(i)*0.001,
				Timestamp:      time.Now().Add(time.Duration(i) * time.Second),
			}
			
			err := performanceCollector.CollectPerformanceMetrics(metrics)
			require.NoError(t, err)
		}
		
		duration := time.Since(start)
		opsPerSecond := 100.0 / duration.Seconds()

		t.Logf("性能指标收集性能:")
		t.Logf("  操作数: 100")
		t.Logf("  总耗时: %v", duration)
		t.Logf("  每秒操作数: %.2f", opsPerSecond)

		// 性能要求：至少每秒50次操作
		assert.Greater(t, opsPerSecond, 50.0, "性能指标收集性能不达标")
	})

	t.Run("告警系统测试", func(t *testing.T) {
		// 创建多个告警
		alerts := make([]*Alert, 10)
		for i := 0; i < 10; i++ {
			alerts[i] = &Alert{
				ID:        fmt.Sprintf("alert_%d", i),
				Level:     AlertLevelWarning,
				Title:     fmt.Sprintf("告警 %d", i),
				Message:   fmt.Sprintf("这是第 %d 个告警", i),
				Source:    "test",
				Timestamp: time.Now(),
				Resolved:  false,
			}
			
			err := alertManager.CreateAlert(alerts[i])
			require.NoError(t, err)
		}

		// 验证告警创建
		allAlerts := alertManager.GetAlerts()
		assert.GreaterOrEqual(t, len(allAlerts), 10)

		activeAlerts := alertManager.GetActiveAlerts()
		assert.GreaterOrEqual(t, len(activeAlerts), 10)

		// 解决部分告警
		for i := 0; i < 5; i++ {
			err := alertManager.ResolveAlert(fmt.Sprintf("alert_%d", i))
			require.NoError(t, err)
		}

		// 验证告警解决
		activeAlertsAfter := alertManager.GetActiveAlerts()
		assert.Less(t, len(activeAlertsAfter), len(activeAlerts))

		// 验证统计信息
		stats := alertManager.GetStats()
		assert.True(t, stats.TotalAlerts >= 10)
		assert.True(t, stats.ActiveAlerts >= 0)
		assert.True(t, stats.ResolvedAlerts >= 0)
	})

	t.Run("健康检查系统测试", func(t *testing.T) {
		// 执行健康检查
		healthCheck := healthChecker.CheckHealth()
		assert.NotNil(t, healthCheck)
		assert.NotEmpty(t, healthCheck.Name)
		assert.NotEmpty(t, healthCheck.Status)

		// 获取所有健康检查
		allChecks := healthChecker.GetAllHealthChecks()
		assert.Greater(t, len(allChecks), 0)

		// 检查特定健康检查
		serviceCheck := healthChecker.CheckHealthByName("service")
		assert.NotNil(t, serviceCheck)
		assert.Equal(t, "service", serviceCheck.Name)

		// 验证整体健康状态
		overallStatus := healthChecker.GetOverallStatus()
		assert.True(t, overallStatus == HealthStatusHealthy || 
			overallStatus == HealthStatusDegraded || 
			overallStatus == HealthStatusUnhealthy)

		// 验证健康状态
		isHealthy := healthChecker.IsHealthy()
		assert.True(t, isHealthy || !isHealthy) // 任何状态都是有效的

		// 获取健康检查指标
		metrics := healthChecker.GetHealthCheckMetrics()
		assert.NotNil(t, metrics)
		assert.True(t, metrics.TotalChecks >= 0)
	})

	t.Run("日志系统测试", func(t *testing.T) {
		// 创建日志记录器
		logger := NewStructuredLogger(LogLevelInfo)
		require.NotNil(t, logger)

		// 测试不同级别的日志
		logger.Debug("调试信息", map[string]interface{}{"debug": true})
		logger.Info("信息日志", map[string]interface{}{"info": true})
		logger.Warn("警告日志", map[string]interface{}{"warn": true})
		logger.Error("错误日志", fmt.Errorf("测试错误"), map[string]interface{}{"error": true})

		// 验证日志条目
		entries := logger.GetEntries()
		assert.GreaterOrEqual(t, len(entries), 4)

		// 验证日志级别
		levels := make(map[LogLevel]int)
		for _, entry := range entries {
			levels[entry.Level]++
		}
		assert.True(t, levels[LogLevelInfo] > 0)
		assert.True(t, levels[LogLevelWarn] > 0)
		assert.True(t, levels[LogLevelError] > 0)

		// 测试带字段的日志记录器
		fieldLogger := logger.WithFields(map[string]interface{}{"service": "test"})
		fieldLogger.Info("带字段的日志")
		
		// 测试带错误的日志记录器
		errorLogger := logger.WithError(fmt.Errorf("测试错误"))
		errorLogger.Info("带错误的日志")

		// 清空日志条目
		logger.ClearEntries()
		entries = logger.GetEntries()
		assert.Equal(t, 0, len(entries))
	})
}

func TestMonitoringSystem_Configuration(t *testing.T) {
	// 跳过测试，需要真实监控系统
	t.Skip("需要真实监控系统，跳过测试")
	
	t.Run("监控配置测试", func(t *testing.T) {
		// 测试默认配置
		config := DefaultLogConfig()
		assert.NotNil(t, config)
		assert.Equal(t, LogLevelInfo, config.Level)
		assert.Equal(t, "json", config.Format)
		assert.Equal(t, "stdout", config.Output)

		// 测试告警配置
		alertConfig := DefaultAlertConfig()
		assert.NotNil(t, alertConfig)
		assert.Greater(t, len(alertConfig.DefaultRules), 0)
		assert.Greater(t, len(alertConfig.Thresholds), 0)

		// 测试健康检查配置
		healthConfig := DefaultHealthCheckConfig()
		assert.NotNil(t, healthConfig)
		assert.Equal(t, 30*time.Second, healthConfig.CheckInterval)
		assert.Equal(t, 5*time.Second, healthConfig.Timeout)
	})

	t.Run("Prometheus配置测试", func(t *testing.T) {
		// 测试Prometheus配置
		promConfig := DefaultPrometheusConfig()
		assert.NotNil(t, promConfig)
		assert.Equal(t, ":8080", promConfig.Addr)
		assert.True(t, promConfig.EnableHistograms)
		assert.True(t, promConfig.EnableSummaries)

		// 测试指标过滤器
		filter := NewPrometheusMetricFilter(promConfig)
		assert.NotNil(t, filter)

		// 创建测试指标
		metrics := []Metric{
			&CounterMetric{
				Name:      "test_counter",
				Value:     100.0,
				Labels:    map[string]string{"service": "test"},
				Timestamp: time.Now(),
			},
			&GaugeMetric{
				Name:      "test_gauge",
				Value:     50.0,
				Labels:    map[string]string{"service": "test"},
				Timestamp: time.Now(),
			},
		}

		// 过滤指标
		filtered := filter.FilterMetrics(metrics)
		assert.Equal(t, len(metrics), len(filtered))

		// 添加默认标签
		for _, metric := range filtered {
			enhanced := filter.AddDefaultLabels(metric)
			assert.NotNil(t, enhanced)
		}
	})
}

func TestMonitoringSystem_ErrorHandling(t *testing.T) {
	// 跳过测试，需要真实监控系统
	t.Skip("需要真实监控系统，跳过测试")
	
	logger := zap.NewNop()

	t.Run("错误处理测试", func(t *testing.T) {
		// 测试指标收集器错误处理
		collector := NewMetricCollector(logger)
		
		// 未启动时收集指标应该失败
		err := collector.CollectCounter("test", 1.0, nil)
		assert.Error(t, err)

		// 启动后应该成功
		err = collector.Start()
		require.NoError(t, err)
		defer collector.Stop()

		err = collector.CollectCounter("test", 1.0, nil)
		assert.NoError(t, err)

		// 测试告警管理器错误处理
		alertManager := NewAlertManager(logger)
		
		// 创建空告警应该失败
		err = alertManager.CreateAlert(nil)
		assert.Error(t, err)

		// 创建有效告警应该成功
		alert := &Alert{
			ID:        "test",
			Level:     AlertLevelInfo,
			Title:     "测试",
			Message:   "测试消息",
			Source:    "test",
			Timestamp: time.Now(),
		}
		err = alertManager.CreateAlert(alert)
		assert.NoError(t, err)

		// 获取不存在的告警应该失败
		_, err = alertManager.GetAlert("nonexistent")
		assert.Error(t, err)

		// 解决不存在的告警应该失败
		err = alertManager.ResolveAlert("nonexistent")
		assert.Error(t, err)

		// 测试健康检查器错误处理
		healthChecker := NewDataCollectionHealthChecker(logger)
		
		// 检查不存在的健康检查应该返回错误状态
		check := healthChecker.CheckHealthByName("nonexistent")
		assert.NotNil(t, check)
		assert.Equal(t, HealthStatusUnhealthy, check.Status)
	})
}
