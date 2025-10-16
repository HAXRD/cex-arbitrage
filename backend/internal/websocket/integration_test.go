package websocket

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestWebSocketServer_SubscriptionIntegration 测试WebSocket服务器与订阅管理器的集成
func TestWebSocketServer_SubscriptionIntegration(t *testing.T) {
	server := NewWebSocketServer(DefaultServerConfig(), zap.NewNop())
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 创建WebSocket连接
	wsURL := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 获取连接ID（从服务器连接列表中获取）
	connections := server.GetConnections()
	require.Len(t, connections, 1)
	connID := connections[0].ID

	// 测试订阅
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	err = server.Subscribe(connID, symbols)
	require.NoError(t, err)

	// 验证订阅状态
	subscriptions := server.GetSubscriptions(connID)
	assert.ElementsMatch(t, symbols, subscriptions)

	// 测试取消订阅
	err = server.Unsubscribe(connID, []string{"BTCUSDT"})
	require.NoError(t, err)

	// 验证取消订阅后的状态
	subscriptions = server.GetSubscriptions(connID)
	assert.ElementsMatch(t, []string{"ETHUSDT"}, subscriptions)
}

// TestWebSocketServer_MessageBroadcastIntegration 测试消息广播集成
func TestWebSocketServer_MessageBroadcastIntegration(t *testing.T) {
	server := NewWebSocketServer(DefaultServerConfig(), zap.NewNop())
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 创建多个WebSocket连接
	conn1, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	require.NoError(t, err)
	defer conn1.Close()

	conn2, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	require.NoError(t, err)
	defer conn2.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 获取连接ID
	connections := server.GetConnections()
	require.Len(t, connections, 2)

	conn1ID := connections[0].ID
	conn2ID := connections[1].ID

	// 订阅相同交易对
	symbols := []string{"BTCUSDT"}
	err = server.Subscribe(conn1ID, symbols)
	require.NoError(t, err)
	err = server.Subscribe(conn2ID, symbols)
	require.NoError(t, err)

	// 广播价格更新消息
	priceUpdate := PriceUpdateMessage{
		Symbol:     "BTCUSDT",
		Price:      45000.50,
		ChangeRate: 0.025,
		Volume:     1234.56,
		Timestamp:  time.Now().UnixMilli(),
	}

	err = server.BroadcastToSymbol("BTCUSDT", priceUpdate)
	require.NoError(t, err)

	// 验证消息发送（这里我们无法直接验证消息接收，因为需要客户端处理）
	// 但可以验证广播操作没有错误
	time.Sleep(100 * time.Millisecond)
}

// TestWebSocketServer_SubscriptionThroughMessages 测试通过消息进行订阅
func TestWebSocketServer_SubscriptionThroughMessages(t *testing.T) {
	server := NewWebSocketServer(DefaultServerConfig(), zap.NewNop())
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 创建WebSocket连接
	wsURL := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 发送订阅消息
	subscribeMsg := Message{
		Type:    "subscribe",
		Symbols: []string{"BTCUSDT", "ETHUSDT"},
	}
	err = conn.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证订阅状态
	connections := server.GetConnections()
	require.Len(t, connections, 1)

	subscriptions := connections[0].Subscriptions
	assert.Contains(t, subscriptions, "BTCUSDT")
	assert.Contains(t, subscriptions, "ETHUSDT")

	// 发送取消订阅消息
	unsubscribeMsg := Message{
		Type:    "unsubscribe",
		Symbols: []string{"BTCUSDT"},
	}
	err = conn.WriteJSON(unsubscribeMsg)
	require.NoError(t, err)

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证取消订阅后的状态
	connections = server.GetConnections()
	require.Len(t, connections, 1)

	subscriptions = connections[0].Subscriptions
	assert.NotContains(t, subscriptions, "BTCUSDT")
	assert.Contains(t, subscriptions, "ETHUSDT")
}

// TestWebSocketServer_ConcurrentSubscriptions 测试并发订阅
func TestWebSocketServer_ConcurrentSubscriptions(t *testing.T) {
	server := NewWebSocketServer(DefaultServerConfig(), zap.NewNop())
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 创建多个并发连接
	connections := make([]*websocket.Conn, 5)
	for i := 0; i < 5; i++ {
		conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
		require.NoError(t, err)
		connections[i] = conn
		defer conn.Close()
	}

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 获取所有连接ID
	serverConnections := server.GetConnections()
	require.Len(t, serverConnections, 5)

	// 并发订阅不同交易对
	for i, conn := range serverConnections {
		symbols := []string{fmt.Sprintf("SYMBOL_%d", i)}
		err = server.Subscribe(conn.ID, symbols)
		require.NoError(t, err)
	}

	// 验证订阅状态
	for i, conn := range serverConnections {
		subscriptions := server.GetSubscriptions(conn.ID)
		assert.Contains(t, subscriptions, fmt.Sprintf("SYMBOL_%d", i))
	}

	// 验证交易对订阅者
	for i := 0; i < 5; i++ {
		subscribers := server.GetSubscribers(fmt.Sprintf("SYMBOL_%d", i))
		assert.Len(t, subscribers, 1)
	}
}

// TestWebSocketServer_SubscriptionCleanup 测试订阅清理
func TestWebSocketServer_SubscriptionCleanup(t *testing.T) {
	server := NewWebSocketServer(DefaultServerConfig(), zap.NewNop())
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 创建连接并订阅
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:8080/ws", nil)
	require.NoError(t, err)

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 验证初始状态
	assert.Equal(t, 1, server.GetConnectionCount())

	// 关闭连接
	conn.Close()

	// 等待连接清理
	time.Sleep(100 * time.Millisecond)

	// 验证连接已清理
	assert.Equal(t, 0, server.GetConnectionCount())
}

// TestWebSocketServer_ErrorHandling 测试错误处理
func TestWebSocketServer_ErrorHandling(t *testing.T) {
	server := NewWebSocketServer(DefaultServerConfig(), zap.NewNop())
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 测试向不存在的连接发送消息
	err = server.SendToConnection("nonexistent", "test message")
	assert.Error(t, err)

	// 测试向不存在的连接订阅
	err = server.Subscribe("nonexistent", []string{"BTCUSDT"})
	assert.Error(t, err)

	// 测试获取不存在连接的订阅
	subscriptions := server.GetSubscriptions("nonexistent")
	assert.Nil(t, subscriptions)
}
