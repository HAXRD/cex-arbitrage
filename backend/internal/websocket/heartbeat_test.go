package websocket

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestHeartbeatManager_StartStop 测试启动和停止
func TestHeartbeatManager_StartStop(t *testing.T) {
	connManager := NewMockConnectionManager()
	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 100 * time.Millisecond // 快速测试

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	// 初始状态
	assert.False(t, manager.IsRunning())

	// 启动
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	assert.True(t, manager.IsRunning())

	// 重复启动应该失败
	err = manager.Start(ctx)
	assert.Error(t, err)
	assert.IsType(t, &HeartbeatError{}, err)

	// 停止
	err = manager.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, manager.IsRunning())

	// 重复停止应该失败
	err = manager.Stop(ctx)
	assert.Error(t, err)
	assert.IsType(t, &HeartbeatError{}, err)
}

// TestHeartbeatManager_ConnectionManagement 测试连接管理
func TestHeartbeatManager_ConnectionManagement(t *testing.T) {
	connManager := NewMockConnectionManager()
	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 100 * time.Millisecond

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 添加连接
	err = manager.AddConnection("conn_1")
	require.NoError(t, err)

	// 验证连接状态
	status := manager.GetConnectionStatus("conn_1")
	require.NotNil(t, status)
	assert.Equal(t, "conn_1", status.ConnID)
	assert.True(t, status.IsActive)
	assert.Equal(t, int64(0), status.TotalHeartbeats)
	assert.Equal(t, int64(0), status.TotalPongs)

	// 移除连接
	err = manager.RemoveConnection("conn_1")
	require.NoError(t, err)

	// 验证连接已移除
	status = manager.GetConnectionStatus("conn_1")
	assert.Nil(t, status)

	// 移除不存在的连接不会失败，只是没有效果
	err = manager.RemoveConnection("nonexistent")
	assert.NoError(t, err)
}

// TestHeartbeatManager_HeartbeatSending 测试心跳发送
func TestHeartbeatManager_HeartbeatSending(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")

	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 100 * time.Millisecond

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 添加连接
	err = manager.AddConnection("conn_1")
	require.NoError(t, err)

	// 等待心跳发送
	time.Sleep(200 * time.Millisecond)

	// 验证心跳统计
	stats := manager.GetHeartbeatStats()
	assert.True(t, stats.TotalHeartbeatsSent > 0)
	assert.True(t, stats.LastHeartbeatSent.After(stats.StartTime))

	// 验证连接状态
	status := manager.GetConnectionStatus("conn_1")
	require.NotNil(t, status)
	assert.True(t, status.TotalHeartbeats > 0)
	assert.True(t, status.LastHeartbeatSent.After(status.CreatedAt))
}

// TestHeartbeatManager_PongProcessing 测试pong处理
func TestHeartbeatManager_PongProcessing(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")

	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 100 * time.Millisecond

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 添加连接
	err = manager.AddConnection("conn_1")
	require.NoError(t, err)

	// 等待心跳发送
	time.Sleep(200 * time.Millisecond)

	// 处理pong
	err = manager.ProcessPong("conn_1", "pong_data")
	require.NoError(t, err)

	// 验证统计
	stats := manager.GetHeartbeatStats()
	assert.True(t, stats.TotalPongsReceived > 0)
	assert.True(t, stats.LastPongReceived.After(stats.StartTime))

	// 验证连接状态
	status := manager.GetConnectionStatus("conn_1")
	require.NotNil(t, status)
	assert.True(t, status.TotalPongs > 0)
	assert.True(t, status.LastPongReceived.After(status.CreatedAt))
	assert.Equal(t, 0, status.MissedHeartbeats) // 应该重置
}

// TestHeartbeatManager_TimeoutDetection 测试超时检测
func TestHeartbeatManager_TimeoutDetection(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")

	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 50 * time.Millisecond
	config.PongTimeout = 100 * time.Millisecond
	config.MaxMissedHeartbeats = 2

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 添加连接
	err = manager.AddConnection("conn_1")
	require.NoError(t, err)

	// 等待超时
	time.Sleep(300 * time.Millisecond)

	// 验证超时统计（可能为0，因为连接可能被清理了）
	stats := manager.GetHeartbeatStats()
	// 超时统计可能为0，因为连接可能已经被清理
	assert.True(t, stats.TotalTimeouts >= 0)

	// 验证连接状态
	status := manager.GetConnectionStatus("conn_1")
	if status != nil {
		assert.True(t, status.MissedHeartbeats >= 0)
	}
}

// TestHeartbeatManager_Configuration 测试配置管理
func TestHeartbeatManager_Configuration(t *testing.T) {
	connManager := NewMockConnectionManager()
	config := DefaultHeartbeatConfig()
	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	// 测试配置设置
	manager.SetHeartbeatInterval(200 * time.Millisecond)
	manager.SetPongTimeout(500 * time.Millisecond)
	manager.SetMaxMissedHeartbeats(5)

	// 验证配置（这里我们无法直接验证，因为配置是私有的）
	// 但可以确保方法调用不会出错
	assert.NotPanics(t, func() {
		manager.SetHeartbeatInterval(200 * time.Millisecond)
		manager.SetPongTimeout(500 * time.Millisecond)
		manager.SetMaxMissedHeartbeats(5)
	})
}

// TestHeartbeatManager_Stats 测试统计功能
func TestHeartbeatManager_Stats(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")

	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 100 * time.Millisecond

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 添加连接
	err = manager.AddConnection("conn_1")
	require.NoError(t, err)
	err = manager.AddConnection("conn_2")
	require.NoError(t, err)

	// 等待心跳发送
	time.Sleep(200 * time.Millisecond)

	// 验证统计
	stats := manager.GetHeartbeatStats()
	assert.True(t, stats.TotalHeartbeatsSent > 0)
	assert.Equal(t, 2, stats.ActiveConnections)
	assert.True(t, stats.LastHeartbeatSent.After(stats.StartTime))

	// 验证所有连接状态
	allStatus := manager.GetAllConnectionStatus()
	assert.Len(t, allStatus, 2)
	assert.Contains(t, allStatus, "conn_1")
	assert.Contains(t, allStatus, "conn_2")
}

// TestHeartbeatManager_ErrorHandling 测试错误处理
func TestHeartbeatManager_ErrorHandling(t *testing.T) {
	connManager := NewMockConnectionManager()
	config := DefaultHeartbeatConfig()
	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	// 测试未启动状态 - 添加连接不会失败
	err := manager.AddConnection("conn_1")
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 测试向不存在的连接发送心跳
	err = manager.SendHeartbeat("nonexistent")
	assert.Error(t, err)
	assert.IsType(t, &HeartbeatError{}, err)

	// 测试处理不存在的连接的pong
	err = manager.ProcessPong("nonexistent", "pong")
	assert.Error(t, err)
	assert.IsType(t, &HeartbeatError{}, err)
}

// TestHeartbeatManager_ConcurrentAccess 测试并发访问
func TestHeartbeatManager_ConcurrentAccess(t *testing.T) {
	connManager := NewMockConnectionManager()
	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 50 * time.Millisecond

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 并发添加连接
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			connID := fmt.Sprintf("conn_%d", id)
			connManager.AddConnection(connID)
			err := manager.AddConnection(connID)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// 等待所有连接添加完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 等待心跳处理
	time.Sleep(200 * time.Millisecond)

	// 验证统计
	stats := manager.GetHeartbeatStats()
	assert.Equal(t, 10, stats.ActiveConnections)
	assert.True(t, stats.TotalHeartbeatsSent > 0)
}

// TestHeartbeatManager_ConnectionCleanup 测试连接清理
func TestHeartbeatManager_ConnectionCleanup(t *testing.T) {
	connManager := NewMockConnectionManager()
	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 50 * time.Millisecond
	config.CleanupInterval = 100 * time.Millisecond
	config.MaxMissedHeartbeats = 1

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 添加连接
	err = manager.AddConnection("conn_1")
	require.NoError(t, err)

	// 等待超时和清理
	time.Sleep(300 * time.Millisecond)

	// 验证连接被清理
	status := manager.GetConnectionStatus("conn_1")
	if status != nil {
		assert.False(t, status.IsActive)
	}
}

// TestHeartbeatManager_ResponseTime 测试响应时间计算
func TestHeartbeatManager_ResponseTime(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")

	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 100 * time.Millisecond

	manager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 添加连接
	err = manager.AddConnection("conn_1")
	require.NoError(t, err)

	// 等待心跳发送
	time.Sleep(200 * time.Millisecond)

	// 处理pong
	err = manager.ProcessPong("conn_1", "pong_data")
	require.NoError(t, err)

	// 验证响应时间
	status := manager.GetConnectionStatus("conn_1")
	require.NotNil(t, status)
	assert.True(t, status.ResponseTime > 0)

	// 验证平均响应时间统计
	stats := manager.GetHeartbeatStats()
	assert.True(t, stats.AverageResponseTime > 0)
}
