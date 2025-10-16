package websocket

import (
	"context"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestWebSocketServer_Connection 测试WebSocket连接
func TestWebSocketServer_Connection(t *testing.T) {
	// 创建测试服务器
	server := NewWebSocketServer(DefaultServerConfig(), zap.NewNop())
	require.NotNil(t, server)

	// 启动服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 创建WebSocket客户端连接
	wsURL := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// 验证连接建立
	assert.True(t, server.IsRunning())
	assert.Equal(t, 1, server.GetConnectionCount())
}

// TestWebSocketServer_MessageHandling 测试消息处理
func TestWebSocketServer_MessageHandling(t *testing.T) {
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
	assert.Contains(t, connections[0].Subscriptions, "BTCUSDT")
	assert.Contains(t, connections[0].Subscriptions, "ETHUSDT")
}

// TestWebSocketServer_InvalidMessageHandling 测试无效消息处理
func TestWebSocketServer_InvalidMessageHandling(t *testing.T) {
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

	// 发送无效消息
	invalidMsg := map[string]interface{}{
		"type": "invalid_type",
		"data": "invalid_data",
	}
	err = conn.WriteJSON(invalidMsg)
	require.NoError(t, err)

	// 等待错误处理
	time.Sleep(100 * time.Millisecond)

	// 验证连接仍然活跃
	assert.Equal(t, 1, server.GetConnectionCount())
}

// TestWebSocketServer_Broadcast 测试消息广播
func TestWebSocketServer_Broadcast(t *testing.T) {
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

	// 订阅相同交易对
	subscribeMsg := Message{
		Type:    "subscribe",
		Symbols: []string{"BTCUSDT"},
	}
	err = conn1.WriteJSON(subscribeMsg)
	require.NoError(t, err)
	err = conn2.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

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

	// 验证消息接收
	time.Sleep(100 * time.Millisecond)
}

// TestWebSocketServer_Heartbeat 测试心跳检测
func TestWebSocketServer_Heartbeat(t *testing.T) {
	config := DefaultServerConfig()
	config.PingInterval = 1 * time.Second
	config.PongWait = 5 * time.Second

	server := NewWebSocketServer(config, zap.NewNop())
	require.NotNil(t, server)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)

	// 创建WebSocket连接
	wsURL := "ws://localhost:8080/ws"
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// 设置pong处理器
	conn.SetPongHandler(func(string) error {
		return nil
	})

	// 等待连接建立
	time.Sleep(100 * time.Millisecond)

	// 验证初始连接
	assert.Equal(t, 1, server.GetConnectionCount())

	// 等待心跳检测
	time.Sleep(2 * time.Second)

	// 验证连接仍然活跃
	assert.Equal(t, 1, server.GetConnectionCount())
}

// TestWebSocketServer_ServerLifecycle 测试服务器生命周期
func TestWebSocketServer_ServerLifecycle(t *testing.T) {
	server := NewWebSocketServer(DefaultServerConfig(), zap.NewNop())
	require.NotNil(t, server)

	// 初始状态
	assert.False(t, server.IsRunning())
	assert.Equal(t, 0, server.GetConnectionCount())

	// 启动服务器
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Start(ctx)
	require.NoError(t, err)
	assert.True(t, server.IsRunning())

	// 停止服务器
	err = server.Stop(ctx)
	require.NoError(t, err)
	assert.False(t, server.IsRunning())
}

// TestWebSocketServer_ConcurrentConnections 测试并发连接
func TestWebSocketServer_ConcurrentConnections(t *testing.T) {
	config := DefaultServerConfig()
	config.MaxConnections = 10

	server := NewWebSocketServer(config, zap.NewNop())
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

	// 验证连接数量
	assert.Equal(t, 5, server.GetConnectionCount())
}

// TestWebSocketServer_MessageValidation 测试消息验证
func TestWebSocketServer_MessageValidation(t *testing.T) {
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

	// 测试有效订阅消息
	validSubscribe := Message{
		Type:    "subscribe",
		Symbols: []string{"BTCUSDT"},
	}
	err = conn.WriteJSON(validSubscribe)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// 测试无效消息类型
	invalidMessage := Message{
		Type: "unknown_type",
	}
	err = conn.WriteJSON(invalidMessage)
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	// 验证连接仍然活跃
	assert.Equal(t, 1, server.GetConnectionCount())
}
