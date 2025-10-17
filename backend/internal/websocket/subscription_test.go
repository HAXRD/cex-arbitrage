package websocket

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestSubscriptionManager_Subscribe 测试订阅功能
func TestSubscriptionManager_Subscribe(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 添加连接
	connID := "conn_1"
	err := manager.AddConnection(connID)
	require.NoError(t, err)

	// 订阅交易对
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	err = manager.Subscribe(connID, symbols)
	require.NoError(t, err)

	// 验证订阅
	subscriptions := manager.GetSubscriptions(connID)
	assert.ElementsMatch(t, symbols, subscriptions)

	// 验证交易对的订阅者
	subscribers := manager.GetSubscribers("BTCUSDT")
	assert.Contains(t, subscribers, connID)

	subscribers = manager.GetSubscribers("ETHUSDT")
	assert.Contains(t, subscribers, connID)
}

// TestSubscriptionManager_Unsubscribe 测试取消订阅功能
func TestSubscriptionManager_Unsubscribe(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 添加连接并订阅
	connID := "conn_1"
	err := manager.AddConnection(connID)
	require.NoError(t, err)

	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	err = manager.Subscribe(connID, symbols)
	require.NoError(t, err)

	// 取消部分订阅
	unsubscribeSymbols := []string{"BTCUSDT", "ETHUSDT"}
	err = manager.Unsubscribe(connID, unsubscribeSymbols)
	require.NoError(t, err)

	// 验证剩余订阅
	subscriptions := manager.GetSubscriptions(connID)
	assert.ElementsMatch(t, []string{"ADAUSDT"}, subscriptions)

	// 验证交易对订阅者
	subscribers := manager.GetSubscribers("BTCUSDT")
	assert.NotContains(t, subscribers, connID)

	subscribers = manager.GetSubscribers("ETHUSDT")
	assert.NotContains(t, subscribers, connID)

	subscribers = manager.GetSubscribers("ADAUSDT")
	assert.Contains(t, subscribers, connID)
}

// TestSubscriptionManager_MultipleConnections 测试多连接订阅
func TestSubscriptionManager_MultipleConnections(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 添加多个连接
	conn1 := "conn_1"
	conn2 := "conn_2"
	conn3 := "conn_3"

	err := manager.AddConnection(conn1)
	require.NoError(t, err)
	err = manager.AddConnection(conn2)
	require.NoError(t, err)
	err = manager.AddConnection(conn3)
	require.NoError(t, err)

	// 订阅不同交易对
	err = manager.Subscribe(conn1, []string{"BTCUSDT", "ETHUSDT"})
	require.NoError(t, err)

	err = manager.Subscribe(conn2, []string{"BTCUSDT", "ADAUSDT"})
	require.NoError(t, err)

	err = manager.Subscribe(conn3, []string{"ETHUSDT", "ADAUSDT"})
	require.NoError(t, err)

	// 验证BTCUSDT的订阅者
	subscribers := manager.GetSubscribers("BTCUSDT")
	assert.ElementsMatch(t, []string{conn1, conn2}, subscribers)

	// 验证ETHUSDT的订阅者
	subscribers = manager.GetSubscribers("ETHUSDT")
	assert.ElementsMatch(t, []string{conn1, conn3}, subscribers)

	// 验证ADAUSDT的订阅者
	subscribers = manager.GetSubscribers("ADAUSDT")
	assert.ElementsMatch(t, []string{conn2, conn3}, subscribers)
}

// TestSubscriptionManager_ConnectionManagement 测试连接管理
func TestSubscriptionManager_ConnectionManagement(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 测试添加连接
	connID := "conn_1"
	err := manager.AddConnection(connID)
	require.NoError(t, err)

	// 验证连接存在
	assert.True(t, manager.IsConnectionActive(connID))
	assert.Equal(t, 1, manager.GetConnectionCount())

	// 测试重复添加连接
	err = manager.AddConnection(connID)
	assert.Error(t, err)
	assert.IsType(t, &SubscriptionError{}, err)

	// 测试移除连接
	err = manager.RemoveConnection(connID)
	require.NoError(t, err)

	// 验证连接已移除
	assert.False(t, manager.IsConnectionActive(connID))
	assert.Equal(t, 0, manager.GetConnectionCount())

	// 测试移除不存在的连接
	err = manager.RemoveConnection("nonexistent")
	assert.Error(t, err)
	assert.IsType(t, &SubscriptionError{}, err)
}

// TestSubscriptionManager_SubscriptionStats 测试订阅统计
func TestSubscriptionManager_SubscriptionStats(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 添加连接和订阅
	conn1 := "conn_1"
	conn2 := "conn_2"

	err := manager.AddConnection(conn1)
	require.NoError(t, err)
	err = manager.AddConnection(conn2)
	require.NoError(t, err)

	err = manager.Subscribe(conn1, []string{"BTCUSDT", "ETHUSDT"})
	require.NoError(t, err)
	err = manager.Subscribe(conn2, []string{"BTCUSDT", "ADAUSDT"})
	require.NoError(t, err)

	// 获取统计信息
	stats := manager.GetSubscriptionStats()

	assert.Equal(t, 2, stats.TotalConnections)
	assert.Equal(t, 2, stats.ActiveConnections)
	assert.Equal(t, 4, stats.TotalSubscriptions)
	assert.Equal(t, 2, stats.SymbolSubscriptions["BTCUSDT"])
	assert.Equal(t, 1, stats.SymbolSubscriptions["ETHUSDT"])
	assert.Equal(t, 1, stats.SymbolSubscriptions["ADAUSDT"])
	assert.Equal(t, 2, stats.ConnectionSubscriptions[conn1])
	assert.Equal(t, 2, stats.ConnectionSubscriptions[conn2])
}

// TestSubscriptionManager_ErrorHandling 测试错误处理
func TestSubscriptionManager_ErrorHandling(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 测试向不存在的连接订阅
	err := manager.Subscribe("nonexistent", []string{"BTCUSDT"})
	assert.Error(t, err)
	assert.IsType(t, &SubscriptionError{}, err)

	// 测试从不存在的连接取消订阅
	err = manager.Unsubscribe("nonexistent", []string{"BTCUSDT"})
	assert.Error(t, err)
	assert.IsType(t, &SubscriptionError{}, err)

	// 测试获取不存在连接的订阅
	subscriptions := manager.GetSubscriptions("nonexistent")
	assert.Nil(t, subscriptions)
}

// TestSubscriptionManager_Cleanup 测试清理功能
func TestSubscriptionManager_Cleanup(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 添加连接
	connID := "conn_1"
	err := manager.AddConnection(connID)
	require.NoError(t, err)

	// 订阅交易对
	err = manager.Subscribe(connID, []string{"BTCUSDT", "ETHUSDT"})
	require.NoError(t, err)

	// 模拟连接变为非活跃状态
	// 注意：这里我们无法直接修改LastActivity，因为它是私有的
	// 在实际实现中，这通常通过心跳检测或其他机制来处理

	// 移除连接（模拟清理）
	err = manager.RemoveConnection(connID)
	require.NoError(t, err)

	// 验证清理结果
	assert.Equal(t, 0, manager.GetConnectionCount())
	assert.Equal(t, 0, manager.GetSymbolCount())

	// 验证交易对订阅者已清理
	subscribers := manager.GetSubscribers("BTCUSDT")
	assert.Empty(t, subscribers)

	subscribers = manager.GetSubscribers("ETHUSDT")
	assert.Empty(t, subscribers)
}

// TestSubscriptionManager_ConcurrentAccess 测试并发访问
func TestSubscriptionManager_ConcurrentAccess(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 并发添加连接
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			connID := fmt.Sprintf("conn_%d", id)
			err := manager.AddConnection(connID)
			if err == nil {
				manager.Subscribe(connID, []string{"BTCUSDT"})
			}
			done <- true
		}(i)
	}

	// 等待所有goroutine完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 验证结果
	assert.Equal(t, 10, manager.GetConnectionCount())
	subscribers := manager.GetSubscribers("BTCUSDT")
	assert.Len(t, subscribers, 10)
}

// TestSubscriptionManager_GetAllSubscriptions 测试获取所有订阅
func TestSubscriptionManager_GetAllSubscriptions(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	// 添加连接和订阅
	conn1 := "conn_1"
	conn2 := "conn_2"

	err := manager.AddConnection(conn1)
	require.NoError(t, err)
	err = manager.AddConnection(conn2)
	require.NoError(t, err)

	err = manager.Subscribe(conn1, []string{"BTCUSDT", "ETHUSDT"})
	require.NoError(t, err)
	err = manager.Subscribe(conn2, []string{"BTCUSDT", "ADAUSDT"})
	require.NoError(t, err)

	// 获取所有订阅
	allSubscriptions := manager.GetAllSubscriptions()

	// 验证结果
	assert.Contains(t, allSubscriptions, "BTCUSDT")
	assert.Contains(t, allSubscriptions, "ETHUSDT")
	assert.Contains(t, allSubscriptions, "ADAUSDT")

	assert.ElementsMatch(t, []string{conn1, conn2}, allSubscriptions["BTCUSDT"])
	assert.ElementsMatch(t, []string{conn1}, allSubscriptions["ETHUSDT"])
	assert.ElementsMatch(t, []string{conn2}, allSubscriptions["ADAUSDT"])
}

// TestSubscriptionManager_DuplicateSubscriptions 测试重复订阅
func TestSubscriptionManager_DuplicateSubscriptions(t *testing.T) {
	manager := NewSubscriptionManager(zap.NewNop())

	connID := "conn_1"
	err := manager.AddConnection(connID)
	require.NoError(t, err)

	// 重复订阅相同交易对
	symbols := []string{"BTCUSDT", "BTCUSDT", "ETHUSDT"}
	err = manager.Subscribe(connID, symbols)
	require.NoError(t, err)

	// 验证没有重复
	subscriptions := manager.GetSubscriptions(connID)
	assert.ElementsMatch(t, []string{"BTCUSDT", "ETHUSDT"}, subscriptions)

	// 验证交易对订阅者没有重复
	subscribers := manager.GetSubscribers("BTCUSDT")
	assert.Len(t, subscribers, 1)
	assert.Contains(t, subscribers, connID)
}
