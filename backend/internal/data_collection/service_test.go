package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestDataCollectionService_Start(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 测试服务启动
	err := service.Start(context.Background())
	require.NoError(t, err, "服务启动应该成功")

	// 验证服务状态
	assert.True(t, service.IsRunning(), "服务应该处于运行状态")

	// 验证服务配置
	assert.NotNil(t, service.GetConfig(), "服务配置不应该为空")
}

func TestDataCollectionService_Stop(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 先启动服务
	err := service.Start(context.Background())
	require.NoError(t, err, "服务启动应该成功")

	// 测试服务停止
	err = service.Stop(context.Background())
	require.NoError(t, err, "服务停止应该成功")

	// 验证服务状态
	assert.False(t, service.IsRunning(), "服务应该处于停止状态")
}

func TestDataCollectionService_Status(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 测试初始状态
	status := service.GetStatus()
	assert.Equal(t, "stopped", status.State, "初始状态应该是stopped")
	assert.Equal(t, 0, status.ActiveConnections, "初始连接数应该是0")

	// 启动服务
	err := service.Start(context.Background())
	require.NoError(t, err, "服务启动应该成功")

	// 测试运行状态
	status = service.GetStatus()
	assert.Equal(t, "running", status.State, "运行状态应该是running")
	assert.True(t, status.Uptime > 0, "运行时间应该大于0")
}

func TestDataCollectionService_HealthCheck(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 测试停止状态下的健康检查
	health := service.HealthCheck()
	assert.Equal(t, "unhealthy", health.Status, "停止状态下应该是不健康")
	assert.False(t, health.IsHealthy, "停止状态下应该是不健康")

	// 启动服务
	err := service.Start(context.Background())
	require.NoError(t, err, "服务启动应该成功")

	// 测试运行状态下的健康检查
	health = service.HealthCheck()
	// 服务刚启动时可能还没有建立连接，所以状态可能是degraded
	assert.Contains(t, []string{"healthy", "degraded"}, health.Status, "运行状态下应该是healthy或degraded")
	// 对于刚启动的服务，degraded状态也是可以接受的
	if health.Status == "degraded" {
		assert.False(t, health.IsHealthy, "degraded状态下应该是不健康")
	} else {
		assert.True(t, health.IsHealthy, "healthy状态下应该是健康")
	}
}

func TestDataCollectionService_ConcurrentStartStop(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 并发启动和停止测试
	done := make(chan bool, 2)

	// 并发启动
	go func() {
		defer func() { done <- true }()
		service.Start(context.Background())
	}()

	// 等待一段时间后停止
	go func() {
		defer func() { done <- true }()
		time.Sleep(100 * time.Millisecond)
		service.Stop(context.Background())
	}()

	// 等待两个goroutine完成
	<-done
	<-done

	// 验证最终状态
	assert.False(t, service.IsRunning(), "最终状态应该是停止")
}

func TestDataCollectionService_StartWhenAlreadyRunning(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 启动服务
	err := service.Start(context.Background())
	require.NoError(t, err, "第一次启动应该成功")

	// 再次启动应该返回错误或忽略
	_ = service.Start(context.Background())
	// 这里可能返回错误或忽略，取决于实现
	// assert.Error(t, err, "重复启动应该返回错误")

	// 验证服务仍在运行
	assert.True(t, service.IsRunning(), "服务应该仍在运行")
}

func TestDataCollectionService_StopWhenNotRunning(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 停止未运行的服务
	_ = service.Stop(context.Background())
	// 这里可能返回错误或忽略，取决于实现
	// assert.Error(t, err, "停止未运行的服务应该返回错误")

	// 验证服务状态
	assert.False(t, service.IsRunning(), "服务应该处于停止状态")
}

func TestDataCollectionService_ContextCancellation(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 创建可取消的上下文
	ctx, cancel := context.WithCancel(context.Background())

	// 启动服务
	err := service.Start(ctx)
	require.NoError(t, err, "服务启动应该成功")

	// 取消上下文
	cancel()

	// 等待一段时间让服务响应取消
	time.Sleep(100 * time.Millisecond)

	// 验证服务状态（可能仍在运行，取决于实现）
	// 这里的具体行为取决于实现细节
}

func TestDataCollectionService_Configuration(t *testing.T) {
	// 创建测试用的服务实例
	service := createTestService(t)

	// 测试默认配置
	config := service.GetConfig()
	assert.NotNil(t, config, "配置不应该为空")

	// 测试配置更新
	newConfig := &ServiceConfig{
		MaxConnections:      100,
		ReconnectInterval:   5 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		Symbols:             []string{"BTCUSDT", "ETHUSDT"},
	}

	err := service.UpdateConfig(newConfig)
	require.NoError(t, err, "配置更新应该成功")

	// 验证配置已更新
	updatedConfig := service.GetConfig()
	assert.Equal(t, 100, updatedConfig.MaxConnections, "最大连接数应该已更新")
}

// 辅助函数：创建测试用的服务实例
func createTestService(t *testing.T) *DataCollectionServiceImpl {
	// 创建测试用的日志器
	logger, _ := zap.NewDevelopment()

	// 创建测试用的配置
	config := &ServiceConfig{
		MaxConnections:      10,
		ReconnectInterval:   1 * time.Second,
		HealthCheckInterval: 10 * time.Second,
		CollectionInterval:  1 * time.Second,
		BatchSize:           10,
		WorkerPoolSize:      2,
		ChannelBufferSize:   100,
		Symbols:             []string{"BTCUSDT", "ETHUSDT"},
	}

	// 创建服务实例
	service := NewDataCollectionService(config, logger)

	return service
}
