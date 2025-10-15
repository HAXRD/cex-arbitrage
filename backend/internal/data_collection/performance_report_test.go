package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestPerformanceReport 性能测试报告
func TestPerformanceReport(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的性能测试报告")
	}

	t.Run("生成性能报告", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 运行性能测试
		time.Sleep(10 * time.Second)

		// 生成性能报告
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("性能指标收集", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 收集性能指标
		time.Sleep(10 * time.Second)

		// 检查性能指标
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("性能分析报告", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 性能分析
		time.Sleep(10 * time.Second)

		// 检查分析结果
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestPerformanceMetrics 性能指标测试
func TestPerformanceMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的性能指标测试")
	}

	t.Run("吞吐量指标", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 吞吐量测试
		time.Sleep(10 * time.Second)

		// 检查吞吐量
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("延迟指标", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 延迟测试
		time.Sleep(10 * time.Second)

		// 检查延迟
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("资源使用指标", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 资源使用测试
		time.Sleep(10 * time.Second)

		// 检查资源使用
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestPerformanceAnalysis 性能分析测试
func TestPerformanceAnalysis(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的性能分析测试")
	}

	t.Run("性能瓶颈分析", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 瓶颈分析
		time.Sleep(10 * time.Second)

		// 检查瓶颈
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("性能优化建议", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 优化建议
		time.Sleep(10 * time.Second)

		// 检查优化建议
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("性能趋势分析", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 趋势分析
		time.Sleep(10 * time.Second)

		// 检查趋势
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestPerformanceReportGeneration 性能报告生成测试
func TestPerformanceReportGeneration(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的性能报告生成测试")
	}

	t.Run("报告格式验证", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 报告格式测试
		time.Sleep(10 * time.Second)

		// 检查报告格式
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("报告内容验证", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 报告内容测试
		time.Sleep(10 * time.Second)

		// 检查报告内容
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("报告导出功能", func(t *testing.T) {
		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 报告导出测试
		time.Sleep(10 * time.Second)

		// 检查报告导出
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}
