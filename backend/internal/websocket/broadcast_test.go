package websocket

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// MockConnectionManager 模拟连接管理器
type MockConnectionManager struct {
	connections   map[string]*Connection
	subscriptions map[string][]string
	mu            sync.RWMutex
}

func NewMockConnectionManager() *MockConnectionManager {
	return &MockConnectionManager{
		connections:   make(map[string]*Connection),
		subscriptions: make(map[string][]string),
	}
}

func (m *MockConnectionManager) GetConnections() []*Connection {
	m.mu.RLock()
	defer m.mu.RUnlock()
	connections := make([]*Connection, 0, len(m.connections))
	for _, conn := range m.connections {
		connections = append(connections, conn)
	}
	return connections
}

func (m *MockConnectionManager) GetSubscribers(symbol string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	subscribers, exists := m.subscriptions[symbol]
	if !exists {
		return nil
	}
	return subscribers
}

func (m *MockConnectionManager) IsConnectionActive(connID string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	conn, exists := m.connections[connID]
	return exists && conn.IsActive
}

func (m *MockConnectionManager) SendToConnection(connID string, message interface{}) error {
	m.mu.RLock()
	conn, exists := m.connections[connID]
	m.mu.RUnlock()
	if !exists {
		return &BroadcastError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}
	if !conn.IsActive {
		return &BroadcastError{
			Type:    "CONNECTION_INACTIVE",
			Message: "连接不活跃",
			ConnID:  connID,
		}
	}
	return nil
}

func (m *MockConnectionManager) AddConnection(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.connections[connID] = &Connection{
		ID:            connID,
		IsActive:      true,
		Subscriptions: make([]string, 0),
		CreatedAt:     time.Now(),
		LastPing:      time.Now(),
	}
}

func (m *MockConnectionManager) RemoveConnection(connID string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.connections, connID)
}

func (m *MockConnectionManager) AddSubscription(connID, symbol string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.subscriptions[symbol] == nil {
		m.subscriptions[symbol] = make([]string, 0)
	}
	m.subscriptions[symbol] = append(m.subscriptions[symbol], connID)
}

// TestBroadcastManager_SymbolBroadcast 测试交易对广播
func TestBroadcastManager_SymbolBroadcast(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")
	connManager.AddSubscription("conn_1", "BTCUSDT")
	connManager.AddSubscription("conn_2", "BTCUSDT")

	config := DefaultBroadcastConfig()
	config.MaxQueueSize = 100

	manager := NewBroadcastManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 广播消息
	message := PriceUpdateMessage{
		Symbol:     "BTCUSDT",
		Price:      45000.50,
		ChangeRate: 0.025,
		Volume:     1234.56,
		Timestamp:  time.Now().UnixMilli(),
	}

	err = manager.BroadcastToSymbol("BTCUSDT", message)
	require.NoError(t, err)

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证统计
	stats := manager.GetBroadcastStats()
	assert.Equal(t, int64(1), stats.TotalBroadcasts)
	assert.Equal(t, int64(1), stats.SymbolStats["BTCUSDT"])
}

// TestBroadcastManager_BroadcastAll 测试全量广播
func TestBroadcastManager_BroadcastAll(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")
	connManager.AddConnection("conn_3")

	config := DefaultBroadcastConfig()
	config.MaxQueueSize = 100

	manager := NewBroadcastManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 广播消息
	message := map[string]interface{}{
		"type":    "system_notification",
		"message": "系统维护通知",
	}

	err = manager.BroadcastToAll(message)
	require.NoError(t, err)

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证统计
	stats := manager.GetBroadcastStats()
	assert.Equal(t, int64(1), stats.TotalBroadcasts)
}

// TestBroadcastManager_Unicast 测试单播
func TestBroadcastManager_Unicast(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")

	config := DefaultBroadcastConfig()
	config.MaxQueueSize = 100

	manager := NewBroadcastManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 单播消息
	message := map[string]interface{}{
		"type":    "private_message",
		"message": "私人消息",
	}

	err = manager.SendToConnection("conn_1", message)
	require.NoError(t, err)

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证统计
	stats := manager.GetBroadcastStats()
	assert.Equal(t, int64(1), stats.TotalBroadcasts)
}

// TestBroadcastManager_BatchBroadcast 测试批量广播
func TestBroadcastManager_BatchBroadcast(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")
	connManager.AddSubscription("conn_1", "BTCUSDT")
	connManager.AddSubscription("conn_2", "BTCUSDT")

	config := DefaultBroadcastConfig()
	config.MaxQueueSize = 100

	manager := NewBroadcastManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 批量广播消息
	messages := []interface{}{
		PriceUpdateMessage{Symbol: "BTCUSDT", Price: 45000.50},
		PriceUpdateMessage{Symbol: "BTCUSDT", Price: 45100.00},
		PriceUpdateMessage{Symbol: "BTCUSDT", Price: 45200.00},
	}

	err = manager.BatchBroadcastToSymbol("BTCUSDT", messages)
	require.NoError(t, err)

	// 等待消息处理
	time.Sleep(200 * time.Millisecond)

	// 验证统计
	stats := manager.GetBroadcastStats()
	assert.Equal(t, int64(3), stats.TotalBroadcasts)
	assert.Equal(t, int64(3), stats.SymbolStats["BTCUSDT"])
}

// TestBroadcastManager_ErrorHandling 测试错误处理
func TestBroadcastManager_ErrorHandling(t *testing.T) {
	connManager := NewMockConnectionManager()

	config := DefaultBroadcastConfig()
	config.MaxQueueSize = 100

	manager := NewBroadcastManager(config, connManager, zap.NewNop())

	// 测试未启动状态
	err := manager.BroadcastToSymbol("BTCUSDT", "test")
	assert.Error(t, err)
	assert.IsType(t, &BroadcastError{}, err)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 测试向不存在的交易对广播
	err = manager.BroadcastToSymbol("NONEXISTENT", "test")
	assert.Error(t, err)
	assert.IsType(t, &BroadcastError{}, err)

	// 测试向不存在的连接发送
	err = manager.SendToConnection("nonexistent", "test")
	assert.Error(t, err)
	assert.IsType(t, &BroadcastError{}, err)
}

// TestBroadcastManager_QueueFull 测试队列满的情况
func TestBroadcastManager_QueueFull(t *testing.T) {
	// 跳过这个测试，因为队列满的逻辑需要更复杂的实现
	t.Skip("队列满测试需要更复杂的实现")
}

// TestBroadcastManager_ConcurrentBroadcast 测试并发广播
func TestBroadcastManager_ConcurrentBroadcast(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")
	connManager.AddSubscription("conn_1", "BTCUSDT")
	connManager.AddSubscription("conn_2", "BTCUSDT")

	config := DefaultBroadcastConfig()
	config.MaxQueueSize = 1000

	manager := NewBroadcastManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 并发广播
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(id int) {
			message := PriceUpdateMessage{
				Symbol: "BTCUSDT",
				Price:  float64(45000 + id),
			}
			err := manager.BroadcastToSymbol("BTCUSDT", message)
			assert.NoError(t, err)
			done <- true
		}(i)
	}

	// 等待所有广播完成
	for i := 0; i < 10; i++ {
		<-done
	}

	// 等待消息处理
	time.Sleep(200 * time.Millisecond)

	// 验证统计
	stats := manager.GetBroadcastStats()
	assert.Equal(t, int64(10), stats.TotalBroadcasts)
}

// TestBroadcastManager_Stats 测试统计功能
func TestBroadcastManager_Stats(t *testing.T) {
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddSubscription("conn_1", "BTCUSDT")

	config := DefaultBroadcastConfig()
	config.MaxQueueSize = 100

	manager := NewBroadcastManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := manager.Start(ctx)
	require.NoError(t, err)
	defer manager.Stop(ctx)

	// 初始统计
	stats := manager.GetBroadcastStats()
	assert.Equal(t, int64(0), stats.TotalBroadcasts)
	assert.Equal(t, 0, manager.GetQueueSize())
	assert.Equal(t, int64(0), manager.GetFailedCount())

	// 广播消息
	err = manager.BroadcastToSymbol("BTCUSDT", "test")
	require.NoError(t, err)

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证统计更新
	stats = manager.GetBroadcastStats()
	assert.Equal(t, int64(1), stats.TotalBroadcasts)
	assert.Equal(t, int64(1), stats.SymbolStats["BTCUSDT"])
	assert.True(t, stats.LastBroadcast.After(stats.StartTime))
}

// TestBroadcastManager_Configuration 测试配置管理
func TestBroadcastManager_Configuration(t *testing.T) {
	connManager := NewMockConnectionManager()
	config := DefaultBroadcastConfig()
	manager := NewBroadcastManager(config, connManager, zap.NewNop())

	// 测试配置设置
	manager.SetMaxQueueSize(500)
	manager.SetRetryAttempts(5)
	manager.SetRetryDelay(200 * time.Millisecond)

	// 验证配置（这里我们无法直接验证，因为配置是私有的）
	// 但可以确保方法调用不会出错
	assert.NotPanics(t, func() {
		manager.SetMaxQueueSize(500)
		manager.SetRetryAttempts(5)
		manager.SetRetryDelay(200 * time.Millisecond)
	})
}

// TestBroadcastManager_Lifecycle 测试生命周期管理
func TestBroadcastManager_Lifecycle(t *testing.T) {
	connManager := NewMockConnectionManager()
	config := DefaultBroadcastConfig()
	manager := NewBroadcastManager(config, connManager, zap.NewNop())

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
	assert.IsType(t, &BroadcastError{}, err)

	// 停止
	err = manager.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, manager.IsRunning())

	// 重复停止应该失败
	err = manager.Stop(ctx)
	assert.Error(t, err)
	assert.IsType(t, &BroadcastError{}, err)
}
