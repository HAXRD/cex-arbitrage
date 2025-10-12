// +build integration

package database

import (
	"context"
	"testing"
	"time"

	"github.com/haxrd/cryptosignal-hunter/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

// setupMonitoringIntegrationTest 设置监控集成测试
func setupMonitoringIntegrationTest(t *testing.T) (*gorm.DB, *zap.Logger) {
	cfg := &config.DatabaseConfig{
		Host:            "localhost",
		Port:            5432,
		User:            "postgres",
		Password:        "postgres",
		DBName:          "cryptosignal",
		SSLMode:         "disable",
		MaxOpenConns:    20,
		MaxIdleConns:    10,
		ConnMaxLifetime: 3600,
		ConnMaxIdleTime: 600,
	}

	logger := zap.NewExample()
	db, err := Connect(cfg, logger)
	require.NoError(t, err, "Failed to connect to database")

	return db, logger
}

// TestMonitoringService_Integration_ConnectionPoolStats 集成测试：连接池统计
func TestMonitoringService_Integration_ConnectionPoolStats(t *testing.T) {
	db, logger := setupMonitoringIntegrationTest(t)
	defer Close()

	monitor := NewMonitoringService(db, logger)
	ctx := context.Background()

	// 测试连接池统计记录
	err := monitor.LogConnectionPoolStats(ctx)
	require.NoError(t, err, "Failed to log connection pool stats")

	// 验证健康状态
	status := monitor.GetHealthStatus(ctx)
	assert.NotNil(t, status)
	assert.True(t, status["healthy"].(bool))
	assert.Contains(t, status, "max_open_conns")
	assert.Contains(t, status, "utilization_rate")

	t.Logf("Database health status: %+v", status)
}

// TestMonitoringService_Integration_PeriodicMonitoring 集成测试：定期监控
func TestMonitoringService_Integration_PeriodicMonitoring(t *testing.T) {
	db, logger := setupMonitoringIntegrationTest(t)
	defer Close()

	monitor := NewMonitoringService(db, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	// 启动定期监控，每 500ms 记录一次
	go monitor.StartPeriodicMonitoring(ctx, 500*time.Millisecond)

	// 等待监控运行
	time.Sleep(2 * time.Second)

	// 取消监控
	cancel()
	time.Sleep(100 * time.Millisecond)

	t.Log("Periodic monitoring test completed")
}

// TestMonitoringService_Integration_HighConnectionLoad 集成测试：高连接负载
func TestMonitoringService_Integration_HighConnectionLoad(t *testing.T) {
	db, logger := setupMonitoringIntegrationTest(t)
	defer Close()

	monitor := NewMonitoringService(db, logger)
	ctx := context.Background()

	// 模拟高连接负载
	// 通过执行多个并发查询来增加连接使用
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// 执行一些数据库操作
			var result []map[string]interface{}
			err := db.WithContext(ctx).Raw("SELECT 1 as test").Scan(&result).Error
			assert.NoError(t, err)
			
			// 短暂等待
			time.Sleep(100 * time.Millisecond)
		}()
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 记录连接池状态
	err := monitor.LogConnectionPoolStats(ctx)
	require.NoError(t, err)

	// 获取健康状态
	status := monitor.GetHealthStatus(ctx)
	assert.True(t, status["healthy"].(bool))
	
	utilizationRate := status["utilization_rate"].(float64)
	t.Logf("Connection pool utilization: %.2f%%", utilizationRate)
}

// TestMonitoringService_Integration_HealthCheck 集成测试：健康检查
func TestMonitoringService_Integration_HealthCheck(t *testing.T) {
	db, logger := setupMonitoringIntegrationTest(t)
	defer Close()

	monitor := NewMonitoringService(db, logger)
	ctx := context.Background()

	// 测试健康检查
	status := monitor.GetHealthStatus(ctx)
	
	assert.NotNil(t, status)
	assert.True(t, status["healthy"].(bool))
	
	// 验证关键指标存在
	requiredFields := []string{
		"healthy", "max_open_conns", "open_conns", 
		"in_use", "idle", "utilization_rate",
	}
	
	for _, field := range requiredFields {
		assert.Contains(t, status, field, "Missing field: %s", field)
	}

	t.Logf("Health check passed: %+v", status)
}

// TestMonitoringService_Integration_ConcurrentAccess 集成测试：并发访问
func TestMonitoringService_Integration_ConcurrentAccess(t *testing.T) {
	db, logger := setupMonitoringIntegrationTest(t)
	defer Close()

	monitor := NewMonitoringService(db, logger)
	ctx := context.Background()

	// 并发执行监控操作
	done := make(chan bool, 5)
	
	for i := 0; i < 5; i++ {
		go func(id int) {
			defer func() { done <- true }()
			
			// 并发记录连接池统计
			err := monitor.LogConnectionPoolStats(ctx)
			assert.NoError(t, err, "Goroutine %d failed", id)
			
			// 并发获取健康状态
			status := monitor.GetHealthStatus(ctx)
			assert.True(t, status["healthy"].(bool), "Goroutine %d health check failed", id)
			
			t.Logf("Goroutine %d completed monitoring", id)
		}(i)
	}

	// 等待所有 goroutine 完成
	for i := 0; i < 5; i++ {
		<-done
	}

	t.Log("Concurrent monitoring access test completed")
}

// TestMonitoringService_Integration_ErrorHandling 集成测试：错误处理
func TestMonitoringService_Integration_ErrorHandling(t *testing.T) {
	// 使用无效配置测试错误处理
	cfg := &config.DatabaseConfig{
		Host:     "invalid-host",
		Port:     9999,
		User:     "invalid-user",
		Password: "invalid-password",
		DBName:   "invalid-db",
		SSLMode:  "disable",
	}

	logger := zap.NewExample()
	db, err := Connect(cfg, logger)
	
	// 连接应该失败
	if err != nil {
		t.Logf("Expected connection failure: %v", err)
		return
	}
	
	// 如果连接成功，测试监控
	monitor := NewMonitoringService(db, logger)
	ctx := context.Background()
	
	// 健康检查应该返回不健康状态
	status := monitor.GetHealthStatus(ctx)
	assert.NotNil(t, status)
	
	// 清理
	if db != nil {
		Close()
	}
}

// TestMonitoringService_Integration_LongRunning 集成测试：长时间运行
func TestMonitoringService_Integration_LongRunning(t *testing.T) {
	db, logger := setupMonitoringIntegrationTest(t)
	defer Close()

	monitor := NewMonitoringService(db, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动长时间运行的监控
	go monitor.StartPeriodicMonitoring(ctx, 1*time.Second)

	// 在监控运行期间执行一些数据库操作
	start := time.Now()
	for time.Since(start) < 3*time.Second {
		var result []map[string]interface{}
		err := db.WithContext(ctx).Raw("SELECT NOW() as current_time").Scan(&result).Error
		assert.NoError(t, err)
		
		time.Sleep(500 * time.Millisecond)
	}

	// 等待监控结束
	cancel()
	time.Sleep(100 * time.Millisecond)

	t.Log("Long running monitoring test completed")
}

// TestMonitoringService_Integration_MemoryUsage 集成测试：内存使用
func TestMonitoringService_Integration_MemoryUsage(t *testing.T) {
	db, logger := setupMonitoringIntegrationTest(t)
	defer Close()

	monitor := NewMonitoringService(db, logger)
	ctx := context.Background()

	// 执行多次监控操作，观察内存使用
	for i := 0; i < 100; i++ {
		err := monitor.LogConnectionPoolStats(ctx)
		require.NoError(t, err)
		
		status := monitor.GetHealthStatus(ctx)
		assert.True(t, status["healthy"].(bool))
		
		if i%20 == 0 {
			t.Logf("Completed %d monitoring operations", i+1)
		}
	}

	t.Log("Memory usage test completed")
}

// TestMonitoringService_Integration_RealWorldScenario 集成测试：真实场景
func TestMonitoringService_Integration_RealWorldScenario(t *testing.T) {
	db, logger := setupMonitoringIntegrationTest(t)
	defer Close()

	monitor := NewMonitoringService(db, logger)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 启动定期监控
	go monitor.StartPeriodicMonitoring(ctx, 2*time.Second)

	// 模拟真实应用场景
	// 1. 正常数据库操作
	for i := 0; i < 5; i++ {
		var result []map[string]interface{}
		err := db.WithContext(ctx).Raw("SELECT COUNT(*) as count FROM information_schema.tables").Scan(&result).Error
		assert.NoError(t, err)
		time.Sleep(500 * time.Millisecond)
	}

	// 2. 高负载操作
	done := make(chan bool, 3)
	for i := 0; i < 3; i++ {
		go func() {
			defer func() { done <- true }()
			for j := 0; j < 10; j++ {
				var result []map[string]interface{}
				err := db.WithContext(ctx).Raw("SELECT pg_sleep(0.1)").Scan(&result).Error
				assert.NoError(t, err)
			}
		}()
	}

	// 等待高负载操作完成
	for i := 0; i < 3; i++ {
		<-done
	}

	// 3. 最终健康检查
	status := monitor.GetHealthStatus(ctx)
	assert.True(t, status["healthy"].(bool))
	
	utilizationRate := status["utilization_rate"].(float64)
	t.Logf("Final utilization rate: %.2f%%", utilizationRate)

	// 等待监控结束
	cancel()
	time.Sleep(100 * time.Millisecond)

	t.Log("Real world scenario test completed")
}
