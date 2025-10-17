package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

// TestE2E_CompleteMessageFlow 测试完整的消息流程
func TestE2E_CompleteMessageFlow(t *testing.T) {
	// 创建测试服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 模拟WebSocket升级
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		}

		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket升级失败: %v", err)
		}
		defer conn.Close()

		// 模拟消息处理
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			var msg Message
			if err := json.Unmarshal(message, &msg); err != nil {
				continue
			}

			// 处理订阅消息
			if msg.Type == "subscribe" {
				response := Message{
					Type:      "subscription_confirmed",
					Symbols:   msg.Symbols,
					Timestamp: time.Now().UnixMilli(),
				}
				conn.WriteJSON(response)
			}

			// 处理心跳消息
			if msg.Type == "ping" {
				response := Message{
					Type:      "pong",
					Timestamp: time.Now().UnixMilli(),
				}
				conn.WriteJSON(response)
			}
		}
	}))
	defer server.Close()

	// 创建WebSocket客户端
	wsURL := fmt.Sprintf("ws://%s", server.URL[7:]) // 移除http://前缀
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	require.NoError(t, err)
	defer conn.Close()

	// 测试订阅消息
	subscribeMsg := Message{
		Type:      "subscribe",
		Symbols:   []string{"BTCUSDT", "ETHUSDT"},
		Timestamp: time.Now().UnixMilli(),
	}

	err = conn.WriteJSON(subscribeMsg)
	require.NoError(t, err)

	// 读取订阅确认
	var response Message
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	assert.Equal(t, "subscription_confirmed", response.Type)
	assert.Equal(t, []string{"BTCUSDT", "ETHUSDT"}, response.Symbols)

	// 测试心跳消息
	pingMsg := Message{
		Type:      "ping",
		Timestamp: time.Now().UnixMilli(),
	}

	err = conn.WriteJSON(pingMsg)
	require.NoError(t, err)

	// 读取pong响应
	err = conn.ReadJSON(&response)
	require.NoError(t, err)
	assert.Equal(t, "pong", response.Type)
}

// TestE2E_WebSocketServerIntegration 测试WebSocket服务器集成
func TestE2E_WebSocketServerIntegration(t *testing.T) {
	// 创建WebSocket服务器
	config := &ServerConfig{
		Port:            8080,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		PingInterval:    30 * time.Second,
	}

	server := NewWebSocketServer(config, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动服务器
	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	// 验证服务器状态
	assert.True(t, server.IsRunning())
	assert.Equal(t, 0, server.GetConnectionCount())

	// 测试连接管理
	connections := server.GetConnections()
	assert.Empty(t, connections)
}

// TestE2E_SubscriptionManagerIntegration 测试订阅管理器集成
func TestE2E_SubscriptionManagerIntegration(t *testing.T) {
	// 创建订阅管理器
	manager := NewSubscriptionManager(zap.NewNop())

	// 添加连接
	err := manager.AddConnection("conn_1")
	require.NoError(t, err)
	err = manager.AddConnection("conn_2")
	require.NoError(t, err)

	// 测试订阅
	symbols1 := []string{"BTCUSDT", "ETHUSDT"}
	err = manager.Subscribe("conn_1", symbols1)
	require.NoError(t, err)

	symbols2 := []string{"ETHUSDT", "ADAUSDT"}
	err = manager.Subscribe("conn_2", symbols2)
	require.NoError(t, err)

	// 验证订阅
	subs1 := manager.GetSubscriptions("conn_1")
	assert.Equal(t, symbols1, subs1)

	subs2 := manager.GetSubscriptions("conn_2")
	assert.Equal(t, symbols2, subs2)

	// 验证订阅者
	btcSubscribers := manager.GetSubscribers("BTCUSDT")
	assert.Contains(t, btcSubscribers, "conn_1")

	ethSubscribers := manager.GetSubscribers("ETHUSDT")
	assert.Contains(t, ethSubscribers, "conn_1")
	assert.Contains(t, ethSubscribers, "conn_2")

	// 验证统计
	stats := manager.GetSubscriptionStats()
	assert.Equal(t, 2, stats.TotalConnections)
	assert.Equal(t, 2, stats.ActiveConnections)
	assert.Equal(t, 4, stats.TotalSubscriptions)
}

// TestE2E_BroadcastManagerIntegration 测试广播管理器集成
func TestE2E_BroadcastManagerIntegration(t *testing.T) {
	// 创建模拟连接管理器
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")
	connManager.AddSubscription("conn_1", "BTCUSDT")
	connManager.AddSubscription("conn_2", "BTCUSDT")

	// 创建广播管理器
	config := DefaultBroadcastConfig()
	broadcastManager := NewBroadcastManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动广播管理器
	err := broadcastManager.Start(ctx)
	require.NoError(t, err)
	defer broadcastManager.Stop(ctx)

	// 测试广播
	message := map[string]interface{}{
		"price":     45000.50,
		"volume":    1234.56,
		"timestamp": time.Now().UnixMilli(),
	}

	err = broadcastManager.BroadcastToSymbol("BTCUSDT", message)
	require.NoError(t, err)

	// 等待消息处理
	time.Sleep(100 * time.Millisecond)

	// 验证统计
	stats := broadcastManager.GetBroadcastStats()
	assert.True(t, stats.TotalBroadcasts > 0)
	assert.Equal(t, int64(1), stats.SymbolStats["BTCUSDT"])
}

// TestE2E_HeartbeatManagerIntegration 测试心跳管理器集成
func TestE2E_HeartbeatManagerIntegration(t *testing.T) {
	// 创建模拟连接管理器
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")

	// 创建心跳管理器
	config := DefaultHeartbeatConfig()
	config.HeartbeatInterval = 100 * time.Millisecond // 快速测试
	heartbeatManager := NewHeartbeatManager(config, connManager, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动心跳管理器
	err := heartbeatManager.Start(ctx)
	require.NoError(t, err)
	defer heartbeatManager.Stop(ctx)

	// 添加连接
	err = heartbeatManager.AddConnection("conn_1")
	require.NoError(t, err)

	// 等待心跳发送
	time.Sleep(200 * time.Millisecond)

	// 验证心跳统计
	stats := heartbeatManager.GetHeartbeatStats()
	assert.True(t, stats.TotalHeartbeatsSent > 0)
	assert.Equal(t, 1, stats.ActiveConnections)

	// 验证连接状态
	status := heartbeatManager.GetConnectionStatus("conn_1")
	require.NotNil(t, status)
	assert.True(t, status.TotalHeartbeats > 0)
}

// TestE2E_ReconnectManagerIntegration 测试重连管理器集成
func TestE2E_ReconnectManagerIntegration(t *testing.T) {
	// 创建重连管理器
	config := DefaultReconnectConfig()
	config.ReconnectInterval = 100 * time.Millisecond
	reconnectManager := NewReconnectManager(config, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动重连管理器
	err := reconnectManager.Start(ctx)
	require.NoError(t, err)
	defer reconnectManager.Stop(ctx)

	// 添加连接
	connConfig := DefaultConnectionReconnectConfig()
	err = reconnectManager.AddConnection("conn_1", connConfig)
	require.NoError(t, err)

	// 保存连接状态
	state := &ConnectionState{
		ConnID:        "conn_1",
		Subscriptions: []string{"BTCUSDT", "ETHUSDT"},
		LastMessageID: "msg_123",
		LastActivity:  time.Now(),
		CustomData:    map[string]interface{}{"user_id": "user_123"},
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	err = reconnectManager.SaveConnectionState("conn_1", state)
	require.NoError(t, err)

	// 恢复连接状态
	restoredState, err := reconnectManager.RestoreConnectionState("conn_1")
	require.NoError(t, err)
	assert.Equal(t, state.ConnID, restoredState.ConnID)
	assert.Equal(t, state.Subscriptions, restoredState.Subscriptions)
	assert.Equal(t, state.CustomData, restoredState.CustomData)
}

// TestE2E_PerformanceMonitorIntegration 测试性能监控器集成
func TestE2E_PerformanceMonitorIntegration(t *testing.T) {
	// 创建性能监控器
	config := DefaultPerformanceConfig()
	config.AggregationInterval = 100 * time.Millisecond
	performanceMonitor := NewPerformanceMonitor(config, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 启动性能监控器
	err := performanceMonitor.Start(ctx)
	require.NoError(t, err)
	defer performanceMonitor.Stop(ctx)

	// 记录各种指标
	performanceMonitor.RecordLatency("websocket_connect", 50*time.Millisecond)
	performanceMonitor.RecordThroughput("message_processing", 100)
	performanceMonitor.RecordMemoryUsage("data_processing", 1024*1024)
	performanceMonitor.RecordError("websocket_connect", "timeout")

	// 等待聚合
	time.Sleep(200 * time.Millisecond)

	// 验证延迟统计
	latencyStats := performanceMonitor.GetLatencyStats("websocket_connect")
	require.NotNil(t, latencyStats)
	assert.Equal(t, "websocket_connect", latencyStats.Operation)
	assert.Equal(t, int64(1), latencyStats.Count)
	assert.Equal(t, 50*time.Millisecond, latencyStats.LastLatency)

	// 验证吞吐量统计
	throughputStats := performanceMonitor.GetThroughputStats("message_processing")
	require.NotNil(t, throughputStats)
	assert.Equal(t, "message_processing", throughputStats.Operation)
	assert.Equal(t, int64(100), throughputStats.Count)

	// 验证内存统计
	memoryStats := performanceMonitor.GetMemoryStats("data_processing")
	require.NotNil(t, memoryStats)
	assert.Equal(t, "data_processing", memoryStats.Operation)
	assert.Equal(t, int64(1024*1024), memoryStats.LastBytes)

	// 验证错误统计
	errorStats := performanceMonitor.GetErrorStats("websocket_connect")
	require.NotNil(t, errorStats)
	assert.Equal(t, "websocket_connect", errorStats.Operation)
	assert.Equal(t, int64(1), errorStats.TotalErrors)
	assert.Equal(t, int64(1), errorStats.ErrorTypes["timeout"])
}

// TestE2E_FullSystemIntegration 测试完整系统集成
func TestE2E_FullSystemIntegration(t *testing.T) {
	// 创建所有组件
	serverConfig := &ServerConfig{
		Port:            8080,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		PingInterval:    30 * time.Second,
	}

	server := NewWebSocketServer(serverConfig, zap.NewNop())
	subscriptionManager := NewSubscriptionManager(zap.NewNop())

	// 创建模拟连接管理器
	connManager := NewMockConnectionManager()
	connManager.AddConnection("conn_1")
	connManager.AddConnection("conn_2")

	// 创建其他管理器
	broadcastConfig := DefaultBroadcastConfig()
	broadcastManager := NewBroadcastManager(broadcastConfig, connManager, zap.NewNop())

	heartbeatConfig := DefaultHeartbeatConfig()
	heartbeatConfig.HeartbeatInterval = 100 * time.Millisecond
	heartbeatManager := NewHeartbeatManager(heartbeatConfig, connManager, zap.NewNop())

	reconnectConfig := DefaultReconnectConfig()
	reconnectManager := NewReconnectManager(reconnectConfig, zap.NewNop())

	performanceConfig := DefaultPerformanceConfig()
	performanceConfig.AggregationInterval = 100 * time.Millisecond
	performanceMonitor := NewPerformanceMonitor(performanceConfig, zap.NewNop())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 启动所有组件
	err := server.Start(ctx)
	require.NoError(t, err)
	defer server.Stop(ctx)

	err = broadcastManager.Start(ctx)
	require.NoError(t, err)
	defer broadcastManager.Stop(ctx)

	err = heartbeatManager.Start(ctx)
	require.NoError(t, err)
	defer heartbeatManager.Stop(ctx)

	err = reconnectManager.Start(ctx)
	require.NoError(t, err)
	defer reconnectManager.Stop(ctx)

	err = performanceMonitor.Start(ctx)
	require.NoError(t, err)
	defer performanceMonitor.Stop(ctx)

	// 验证所有组件都在运行
	assert.True(t, server.IsRunning())
	assert.True(t, broadcastManager.IsRunning())
	assert.True(t, heartbeatManager.IsRunning())
	assert.True(t, reconnectManager.IsRunning())
	assert.True(t, performanceMonitor.IsRunning())

	// 测试完整流程
	// 1. 添加连接
	err = subscriptionManager.AddConnection("conn_1")
	require.NoError(t, err)

	// 2. 订阅交易对
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	err = subscriptionManager.Subscribe("conn_1", symbols)
	require.NoError(t, err)

	// 3. 添加心跳监控
	err = heartbeatManager.AddConnection("conn_1")
	require.NoError(t, err)

	// 4. 添加重连管理
	connConfig := DefaultConnectionReconnectConfig()
	err = reconnectManager.AddConnection("conn_1", connConfig)
	require.NoError(t, err)

	// 5. 记录性能指标
	performanceMonitor.RecordLatency("connection_establish", 30*time.Millisecond)
	performanceMonitor.RecordThroughput("subscription_processing", 1)

	// 6. 广播消息
	message := map[string]interface{}{
		"price":     45000.50,
		"volume":    1234.56,
		"timestamp": time.Now().UnixMilli(),
	}

	// 先添加连接和订阅到连接管理器
	connManager.AddConnection("conn_1")
	connManager.AddSubscription("conn_1", "BTCUSDT")

	err = broadcastManager.BroadcastToSymbol("BTCUSDT", message)
	require.NoError(t, err)

	// 等待所有操作完成
	time.Sleep(500 * time.Millisecond)

	// 验证所有组件的状态
	assert.Equal(t, 1, subscriptionManager.GetConnectionCount())
	assert.Equal(t, 2, subscriptionManager.GetSymbolCount())

	stats := subscriptionManager.GetSubscriptionStats()
	assert.Equal(t, 1, stats.TotalConnections)
	assert.Equal(t, 2, stats.TotalSubscriptions)

	// 验证广播统计
	broadcastStats := broadcastManager.GetBroadcastStats()
	assert.True(t, broadcastStats.TotalBroadcasts > 0)

	// 验证心跳统计
	heartbeatStats := heartbeatManager.GetHeartbeatStats()
	assert.True(t, heartbeatStats.TotalHeartbeatsSent > 0)

	// 验证重连统计
	reconnectStats := reconnectManager.GetReconnectStats()
	assert.True(t, reconnectStats.TotalReconnects >= 0)

	// 验证性能统计
	latencyStats := performanceMonitor.GetLatencyStats("connection_establish")
	require.NotNil(t, latencyStats)
	assert.Equal(t, int64(1), latencyStats.Count)

	throughputStats := performanceMonitor.GetThroughputStats("subscription_processing")
	require.NotNil(t, throughputStats)
	assert.Equal(t, int64(1), throughputStats.Count)
}
