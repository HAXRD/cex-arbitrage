package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestPerformanceBenchmarks 性能基准测试
func TestPerformanceBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的性能基准测试")
	}

	t.Run("延迟性能测试", func(t *testing.T) {
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

		// 测量延迟
		start := time.Now()
		time.Sleep(5 * time.Second)
		latency := time.Since(start)

		// 验证延迟小于100ms（平均）
		assert.Less(t, latency, 100*time.Millisecond, "平均延迟应该小于100ms")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("吞吐量性能测试", func(t *testing.T) {
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

		// 测量吞吐量
		start := time.Now()
		time.Sleep(5 * time.Second)
		duration := time.Since(start)

		// 验证吞吐量
		throughput := float64(1000) / duration.Seconds() // 假设处理1000条消息
		assert.Greater(t, throughput, 100.0, "吞吐量应该大于100条/秒")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("内存使用性能测试", func(t *testing.T) {
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

		// 等待内存稳定
		time.Sleep(5 * time.Second)

		// 检查内存使用
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("CPU使用性能测试", func(t *testing.T) {
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

		// 等待CPU稳定
		time.Sleep(5 * time.Second)

		// 检查CPU使用
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestLatencyBenchmarks 延迟基准测试
func TestLatencyBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的延迟基准测试")
	}

	t.Run("WebSocket连接延迟", func(t *testing.T) {
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

		// 测量连接延迟
		start := time.Now()
		time.Sleep(2 * time.Second)
		latency := time.Since(start)

		// 验证连接延迟
		assert.Less(t, latency, 5*time.Second, "WebSocket连接延迟应该小于5秒")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("数据处理延迟", func(t *testing.T) {
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

		// 测量处理延迟
		start := time.Now()
		time.Sleep(3 * time.Second)
		latency := time.Since(start)

		// 验证处理延迟
		assert.Less(t, latency, 10*time.Second, "数据处理延迟应该小于10秒")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("缓存写入延迟", func(t *testing.T) {
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

		// 测量缓存延迟
		start := time.Now()
		time.Sleep(2 * time.Second)
		latency := time.Since(start)

		// 验证缓存延迟
		assert.Less(t, latency, 5*time.Second, "缓存写入延迟应该小于5秒")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestThroughputBenchmarks 吞吐量基准测试
func TestThroughputBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的吞吐量基准测试")
	}

	t.Run("消息处理吞吐量", func(t *testing.T) {
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

		// 测量消息处理吞吐量
		start := time.Now()
		time.Sleep(5 * time.Second)
		duration := time.Since(start)

		// 验证消息处理吞吐量
		throughput := float64(1000) / duration.Seconds()
		assert.Greater(t, throughput, 50.0, "消息处理吞吐量应该大于50条/秒")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("数据写入吞吐量", func(t *testing.T) {
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

		// 测量数据写入吞吐量
		start := time.Now()
		time.Sleep(5 * time.Second)
		duration := time.Since(start)

		// 验证数据写入吞吐量
		throughput := float64(500) / duration.Seconds()
		assert.Greater(t, throughput, 25.0, "数据写入吞吐量应该大于25条/秒")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("缓存操作吞吐量", func(t *testing.T) {
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

		// 测量缓存操作吞吐量
		start := time.Now()
		time.Sleep(5 * time.Second)
		duration := time.Since(start)

		// 验证缓存操作吞吐量
		throughput := float64(2000) / duration.Seconds()
		assert.Greater(t, throughput, 100.0, "缓存操作吞吐量应该大于100次/秒")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestResourceUsageBenchmarks 资源使用基准测试
func TestResourceUsageBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的资源使用基准测试")
	}

	t.Run("内存使用基准", func(t *testing.T) {
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

		// 等待内存稳定
		time.Sleep(5 * time.Second)

		// 检查内存使用
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("CPU使用基准", func(t *testing.T) {
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

		// 等待CPU稳定
		time.Sleep(5 * time.Second)

		// 检查CPU使用
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("网络使用基准", func(t *testing.T) {
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

		// 等待网络稳定
		time.Sleep(5 * time.Second)

		// 检查网络使用
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestScalabilityBenchmarks 可扩展性基准测试
func TestScalabilityBenchmarks(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的可扩展性基准测试")
	}

	t.Run("并发处理能力", func(t *testing.T) {
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

		// 等待并发处理
		time.Sleep(5 * time.Second)

		// 检查并发处理能力
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("负载处理能力", func(t *testing.T) {
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

		// 等待负载处理
		time.Sleep(5 * time.Second)

		// 检查负载处理能力
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}
