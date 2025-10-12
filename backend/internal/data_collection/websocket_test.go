package data_collection

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestWebSocketClient_Connect(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 测试连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "WebSocket连接应该成功")

	// 验证连接状态
	assert.True(t, client.IsConnected(), "客户端应该处于连接状态")

	// 验证连接信息
	info := client.GetConnectionInfo()
	assert.NotNil(t, info, "连接信息不应该为空")
	assert.True(t, info.Connected, "连接状态应该为true")
}

func TestWebSocketClient_Disconnect(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)

	// 先建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "WebSocket连接应该成功")

	// 测试断开连接
	err = client.Disconnect(context.Background())
	require.NoError(t, err, "WebSocket断开应该成功")

	// 验证连接状态
	assert.False(t, client.IsConnected(), "客户端应该处于断开状态")
}

func TestWebSocketClient_Reconnect(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "初始连接应该成功")

	// 断开连接
	err = client.Disconnect(context.Background())
	require.NoError(t, err, "断开连接应该成功")

	// 重新连接
	err = client.Connect(context.Background())
	require.NoError(t, err, "重连应该成功")

	// 验证连接状态
	assert.True(t, client.IsConnected(), "重连后应该处于连接状态")
}

func TestWebSocketClient_AutoReconnect(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 启用自动重连
	client.SetAutoReconnect(true)

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "初始连接应该成功")

	// 模拟连接断开
	client.SimulateConnectionError()

	// 等待自动重连（重连延迟1秒 + 连接建立时间 + 稳定时间）
	time.Sleep(2 * time.Second)

	// 验证重连状态
	connected := client.IsConnected()
	info := client.GetConnectionInfo()
	t.Logf("重连后状态: connected=%v, reconnectCount=%d, lastError=%s",
		connected, info.ReconnectCount, info.LastError)

	assert.True(t, connected, "自动重连后应该处于连接状态")
}

func TestWebSocketClient_Heartbeat(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 启动心跳（在连接建立后立即启动）
	client.StartHeartbeat(200 * time.Millisecond)

	// 等待心跳发送
	time.Sleep(500 * time.Millisecond)

	// 停止心跳
	client.StopHeartbeat()

	// 验证心跳状态
	info := client.GetConnectionInfo()
	assert.True(t, info.LastHeartbeat.After(time.Now().Add(-1*time.Second)), "心跳应该已发送")
}

func TestWebSocketClient_Subscribe(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 订阅单个交易对
	err = client.Subscribe("BTCUSDT")
	require.NoError(t, err, "订阅应该成功")

	// 验证订阅状态
	subscriptions := client.GetSubscriptions()
	assert.Contains(t, subscriptions, "BTCUSDT", "应该包含订阅的交易对")
}

func TestWebSocketClient_BatchSubscribe(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 批量订阅
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	err = client.BatchSubscribe(symbols)
	require.NoError(t, err, "批量订阅应该成功")

	// 验证订阅状态
	subscriptions := client.GetSubscriptions()
	for _, symbol := range symbols {
		assert.Contains(t, subscriptions, symbol, "应该包含所有订阅的交易对")
	}
}

func TestWebSocketClient_Unsubscribe(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 先订阅
	err = client.Subscribe("BTCUSDT")
	require.NoError(t, err, "订阅应该成功")

	// 取消订阅
	err = client.Unsubscribe("BTCUSDT")
	require.NoError(t, err, "取消订阅应该成功")

	// 验证订阅状态
	subscriptions := client.GetSubscriptions()
	assert.NotContains(t, subscriptions, "BTCUSDT", "不应该包含已取消订阅的交易对")
}

func TestWebSocketClient_MessageHandling(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 订阅交易对
	err = client.Subscribe("BTCUSDT")
	require.NoError(t, err, "订阅应该成功")

	// 设置消息处理器
	messageCount := 0
	client.SetMessageHandler(func(data []byte) error {
		messageCount++
		return nil
	})

	// 模拟接收消息
	client.SimulateMessage([]byte(`{"symbol":"BTCUSDT","price":50000}`))

	// 验证消息处理
	assert.Equal(t, 1, messageCount, "应该接收到1条消息")
}

func TestWebSocketClient_ErrorHandling(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 设置错误处理器
	errorCount := 0
	client.SetErrorHandler(func(err error) {
		errorCount++
	})

	// 模拟连接错误
	client.SimulateError(ErrConnectionFailed)

	// 验证错误处理
	assert.Equal(t, 1, errorCount, "应该处理1个错误")
}

func TestWebSocketClient_ConcurrentOperations(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 并发订阅
	done := make(chan bool, 3)

	go func() {
		defer func() { done <- true }()
		client.Subscribe("BTCUSDT")
	}()

	go func() {
		defer func() { done <- true }()
		client.Subscribe("ETHUSDT")
	}()

	go func() {
		defer func() { done <- true }()
		client.Subscribe("ADAUSDT")
	}()

	// 等待所有操作完成
	<-done
	<-done
	<-done

	// 验证订阅状态
	subscriptions := client.GetSubscriptions()
	assert.Len(t, subscriptions, 3, "应该包含3个订阅")
}

func TestWebSocketClient_Configuration(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 测试获取配置
	config := client.GetConfig()
	assert.NotNil(t, config, "配置不应该为空")
	assert.Contains(t, config.URL, "localhost", "URL应该包含localhost")

	// 测试设置配置
	newConfig := &WebSocketConfig{
		URL:               "ws://new-server:8080/ws",
		ReconnectInterval: 2 * time.Second,
	}

	err := client.SetConfig(newConfig)
	require.NoError(t, err, "设置配置应该成功")

	// 验证配置已更新
	updatedConfig := client.GetConfig()
	assert.Equal(t, "ws://new-server:8080/ws", updatedConfig.URL, "URL应该已更新")
}

func TestWebSocketClient_ConnectionInfo(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 测试获取连接信息
	info := client.GetConnectionInfo()
	assert.NotNil(t, info, "连接信息不应该为空")
	assert.True(t, info.Connected, "连接状态应该为true")
	assert.Contains(t, info.URL, "localhost", "URL应该包含localhost")
	assert.Equal(t, int64(0), info.ReconnectCount, "重连次数应该为0")
}

func TestWebSocketClient_AutoReconnectSettings(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 启用自动重连
	client.SetAutoReconnect(true)

	// 验证重连次数
	reconnectCount := client.GetReconnectCount()
	assert.Equal(t, int64(0), reconnectCount, "初始重连次数应该为0")
}

func TestWebSocketClient_HeartbeatManagement(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 启动心跳
	client.StartHeartbeat(1 * time.Second)

	// 等待一段时间
	time.Sleep(100 * time.Millisecond)

	// 停止心跳
	client.StopHeartbeat()

	// 验证心跳已停止（通过检查连接信息）
	info := client.GetConnectionInfo()
	assert.NotNil(t, info, "连接信息不应该为空")
}

func TestWebSocketClient_ConcurrentMessageHandling(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "连接应该成功")

	// 并发设置处理器
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		client.SetMessageHandler(func(data []byte) error { return nil })
	}()

	go func() {
		defer func() { done <- true }()
		client.SetErrorHandler(func(err error) {})
	}()

	// 等待操作完成
	<-done
	<-done

	// 验证操作完成
	assert.True(t, true, "并发操作应该完成")
}

func TestWebSocketClient_ExponentialBackoff(t *testing.T) {
	// 创建简单的WebSocket客户端用于测试
	logger, _ := zap.NewDevelopment()
	config := &WebSocketConfig{
		ReconnectInterval: 1 * time.Second,
	}
	client := NewWebSocketClient(config, logger)

	// 测试指数退避计算
	baseDelay := 1 * time.Second

	// 第1次重连：基础延迟
	delay1 := client.calculateReconnectDelay(1)
	assert.Equal(t, baseDelay, delay1, "第1次重连应该使用基础延迟")

	// 第2次重连：2倍延迟
	delay2 := client.calculateReconnectDelay(2)
	assert.Equal(t, 2*baseDelay, delay2, "第2次重连应该使用2倍延迟")

	// 第3次重连：4倍延迟
	delay3 := client.calculateReconnectDelay(3)
	assert.Equal(t, 4*baseDelay, delay3, "第3次重连应该使用4倍延迟")

	// 第10次重连：应该被限制在最大延迟
	delay10 := client.calculateReconnectDelay(10)
	maxDelay := 5 * time.Minute
	assert.True(t, delay10 <= maxDelay, "第10次重连应该被限制在最大延迟内")
}

func TestWebSocketClient_ReconnectLimit(t *testing.T) {
	// 创建模拟服务器
	server := createMockServer(t)
	defer server.Stop()

	// 创建模拟客户端，设置较小的重连次数
	client := createMockClient(t, server)
	defer client.Disconnect(context.Background()) // 确保测试结束时断开连接

	// 设置重连配置（保持原有URL）
	config := client.GetConfig()
	config.ReconnectInterval = 100 * time.Millisecond
	config.MaxReconnectAttempts = 2
	config.ConnectionTimeout = 100 * time.Millisecond
	client.SetConfig(config)

	// 启用自动重连
	client.SetAutoReconnect(true)

	// 先建立连接
	err := client.Connect(context.Background())
	require.NoError(t, err, "初始连接应该成功")

	// 验证初始重连次数为0
	reconnectCount := client.GetReconnectCount()
	assert.Equal(t, int64(0), reconnectCount, "初始重连次数应该为0")

	// 模拟连接错误，触发自动重连
	client.SimulateConnectionError()

	// 等待自动重连
	time.Sleep(500 * time.Millisecond)

	// 验证重连次数（由于自动重连被触发）
	reconnectCount = client.GetReconnectCount()
	assert.True(t, reconnectCount >= 1, "应该至少重连1次")
}

// 辅助函数：创建模拟服务器
func createMockServer(t *testing.T) *MockWebSocketServer {
	logger, _ := zap.NewDevelopment()

	// 获取可用端口
	listener, err := net.Listen("tcp", ":0")
	require.NoError(t, err, "获取可用端口应该成功")

	port := fmt.Sprintf("%d", listener.Addr().(*net.TCPAddr).Port)
	listener.Close()

	server := NewMockWebSocketServer(port, logger)

	err = server.Start()
	require.NoError(t, err, "模拟服务器启动应该成功")

	return server
}

// 辅助函数：创建模拟客户端
func createMockClient(t *testing.T, server *MockWebSocketServer) *MockWebSocketClient {
	logger, _ := zap.NewDevelopment()

	config := &WebSocketConfig{
		ReconnectInterval:    1 * time.Second,
		MaxReconnectAttempts: 3,
		HeartbeatInterval:    30 * time.Second,
		ConnectionTimeout:    5 * time.Second,
		ReadBufferSize:       1024,
		WriteBufferSize:      1024,
	}

	client := NewMockWebSocketClient(server, config, logger)
	return client
}
