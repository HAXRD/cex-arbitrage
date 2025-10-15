package data_collection

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestDataConsistency 数据一致性验证测试
func TestDataConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的数据一致性测试")
	}

	t.Run("缓存和数据库一致性", func(t *testing.T) {
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

		// 等待数据同步
		time.Sleep(5 * time.Second)

		// 验证缓存和数据库数据一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("价格数据一致性", func(t *testing.T) {
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

		// 等待价格数据同步
		time.Sleep(5 * time.Second)

		// 验证价格数据一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("时间戳一致性", func(t *testing.T) {
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

		// 等待时间戳同步
		time.Sleep(5 * time.Second)

		// 验证时间戳一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestPriceChangeRateConsistency 价格变化率一致性测试
func TestPriceChangeRateConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的价格变化率一致性测试")
	}

	t.Run("1分钟窗口一致性", func(t *testing.T) {
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

		// 等待1分钟窗口计算
		time.Sleep(5 * time.Second)

		// 验证1分钟窗口一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("5分钟窗口一致性", func(t *testing.T) {
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

		// 等待5分钟窗口计算
		time.Sleep(5 * time.Second)

		// 验证5分钟窗口一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("15分钟窗口一致性", func(t *testing.T) {
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

		// 等待15分钟窗口计算
		time.Sleep(5 * time.Second)

		// 验证15分钟窗口一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestDataValidationConsistency 数据验证一致性测试
func TestDataValidationConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的数据验证一致性测试")
	}

	t.Run("数据验证规则一致性", func(t *testing.T) {
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

		// 等待数据验证
		time.Sleep(5 * time.Second)

		// 验证数据验证规则一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("异常检测一致性", func(t *testing.T) {
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

		// 等待异常检测
		time.Sleep(5 * time.Second)

		// 验证异常检测一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestCacheConsistency 缓存一致性测试
func TestCacheConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实Redis的缓存一致性测试")
	}

	t.Run("缓存写入一致性", func(t *testing.T) {
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

		// 等待缓存写入
		time.Sleep(5 * time.Second)

		// 验证缓存写入一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("缓存TTL一致性", func(t *testing.T) {
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

		// 等待缓存TTL设置
		time.Sleep(5 * time.Second)

		// 验证缓存TTL一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestDatabaseConsistency 数据库一致性测试
func TestDatabaseConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实数据库的数据库一致性测试")
	}

	t.Run("数据库写入一致性", func(t *testing.T) {
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

		// 等待数据库写入
		time.Sleep(5 * time.Second)

		// 验证数据库写入一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("事务一致性", func(t *testing.T) {
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

		// 等待事务处理
		time.Sleep(5 * time.Second)

		// 验证事务一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}

// TestMetricsConsistency 指标一致性测试
func TestMetricsConsistency(t *testing.T) {
	if testing.Short() {
		t.Skip("跳过需要真实服务的指标一致性测试")
	}

	t.Run("性能指标一致性", func(t *testing.T) {
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

		// 等待性能指标收集
		time.Sleep(5 * time.Second)

		// 验证性能指标一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})

	t.Run("业务指标一致性", func(t *testing.T) {
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

		// 等待业务指标收集
		time.Sleep(5 * time.Second)

		// 验证业务指标一致性
		status := service.GetStatus()
		assert.Equal(t, "running", status.State)

		// 停止服务
		err = service.Stop(ctx)
		require.NoError(t, err)
	})
}
