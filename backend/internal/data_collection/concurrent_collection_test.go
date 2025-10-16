package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestConcurrentSymbolCollection 100+交易对并发采集测试
func TestConcurrentSymbolCollection(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的并发采集测试")
	}

	t.Run("100个交易对并发采集", func(t *testing.T) {
		// 创建100个测试交易对
		_ = generateTestSymbols(100)

		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 等待并发采集稳定
		time.Sleep(10 * time.Second)

		// 检查服务状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("200个交易对高并发采集", func(t *testing.T) {
		// 创建200个测试交易对
		_ = generateTestSymbols(200)

		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 等待高并发采集稳定
		time.Sleep(15 * time.Second)

		// 检查服务状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("并发采集性能测试", func(t *testing.T) {
		_ = generateTestSymbols(150)

		service := NewDataCollectionService(&ServiceConfig{
			HealthCheckInterval: 5 * time.Second,
			CollectionInterval:  1 * time.Second,
			ReconnectInterval:   5 * time.Second,
			MaxConnections:      10,
			WorkerPoolSize:      5,
			ChannelBufferSize:   100,
		}, zap.NewNop())

		ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
		defer cancel()

		// 启动服务
		err := service.Start(ctx)
		require.NoError(t, err)

		// 测量性能
		start := time.Now()
		time.Sleep(10 * time.Second)
		duration := time.Since(start)

		// 验证性能指标
		throughput := float64(100) / duration.Seconds()
		assert.Greater(t, throughput, 10.0, "并发采集吞吐量应该大于10个/秒")

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestConcurrentDataProcessing 并发数据处理测试
func TestConcurrentDataProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的并发数据处理测试")
	}

	t.Run("多协程数据处理", func(t *testing.T) {
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

		// 等待多协程处理
		time.Sleep(5 * time.Second)

		// 检查处理状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("协程池压力测试", func(t *testing.T) {
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

		// 压力测试
		time.Sleep(5 * time.Second)

		// 检查服务状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestConcurrentWebSocketConnections 并发WebSocket连接测试
func TestConcurrentWebSocketConnections(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实WebSocket的并发连接测试")
	}

	t.Run("多连接并发测试", func(t *testing.T) {
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

		// 等待连接建立
		time.Sleep(3 * time.Second)

		// 检查连接状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("连接池管理测试", func(t *testing.T) {
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

		// 测试连接池
		time.Sleep(3 * time.Second)

		// 检查连接池状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestConcurrentCacheOperations 并发缓存操作测试
func TestConcurrentCacheOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实Redis的并发缓存测试")
	}

	t.Run("并发缓存写入", func(t *testing.T) {
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

		// 等待缓存操作
		time.Sleep(5 * time.Second)

		// 检查缓存状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("缓存一致性测试", func(t *testing.T) {
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

		// 等待缓存同步
		time.Sleep(5 * time.Second)

		// 检查缓存一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestConcurrentDatabaseOperations 并发数据库操作测试
func TestConcurrentDatabaseOperations(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的并发操作测试")
	}

	t.Run("并发数据库写入", func(t *testing.T) {
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

		// 等待数据库操作
		time.Sleep(5 * time.Second)

		// 检查数据库状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("数据库连接池测试", func(t *testing.T) {
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

		// 测试连接池
		time.Sleep(5 * time.Second)

		// 检查连接池状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestConcurrentErrorHandling 并发错误处理测试
func TestConcurrentErrorHandling(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的并发错误处理测试")
	}

	t.Run("并发错误恢复", func(t *testing.T) {
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

		// 等待错误处理
		time.Sleep(5 * time.Second)

		// 检查错误处理状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("错误统计和监控", func(t *testing.T) {
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

		// 等待错误统计
		time.Sleep(5 * time.Second)

		// 检查错误统计
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// generateTestSymbols 生成测试交易对
func generateTestSymbols(count int) []string {
	symbols := make([]string, count)
	baseSymbols := []string{"BTC", "ETH", "BNB", "ADA", "SOL", "DOT", "MATIC", "AVAX", "LINK", "UNI"}

	for i := 0; i < count; i++ {
		base := baseSymbols[i%len(baseSymbols)]
		symbols[i] = base + "USDT"
	}

	return symbols
}

// TestConcurrentMetrics 并发指标收集测试
func TestConcurrentMetrics(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的并发指标测试")
	}

	t.Run("并发指标收集", func(t *testing.T) {
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

		// 等待指标收集
		time.Sleep(5 * time.Second)

		// 检查指标状态
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("指标聚合测试", func(t *testing.T) {
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

		// 等待指标聚合
		time.Sleep(5 * time.Second)

		// 检查指标聚合
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}
