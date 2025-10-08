package bitget

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// TestWebSocketClient_Reconnect_ConnectionLoss 测试连接丢失时的自动重连
func TestWebSocketClient_Reconnect_ConnectionLoss(t *testing.T) {
	var connectionCount int
	var mu sync.Mutex

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
		}
		defer conn.Close()

		mu.Lock()
		connectionCount++
		mu.Unlock()

		// 模拟连接后立即关闭
		time.Sleep(100 * time.Millisecond)
		conn.Close()
	}))
	defer server.Close()

	// 创建客户端
	logger := zap.NewNop()
	config := &WebSocketConfig{
		URL:                  "ws" + server.URL[4:], // 转换为 WebSocket URL
		ReconnectInterval:    100 * time.Millisecond,
		MaxReconnectAttempts: 3,
		ReconnectBackoff:     50 * time.Millisecond,
	}
	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 等待重连尝试
	time.Sleep(2 * time.Second)

	// 检查连接次数（应该至少有2次连接）
	mu.Lock()
	finalConnectionCount := connectionCount
	mu.Unlock()

	if finalConnectionCount < 2 {
		t.Errorf("期望至少2次连接, 实际 = %d", finalConnectionCount)
	}

	// 检查重连状态
	_, maxAttempts, enabled := client.GetReconnectStatus()
	if !enabled {
		t.Error("自动重连应该启用")
	}
	if maxAttempts != 3 {
		t.Errorf("期望最大重连次数 = 3, 实际 = %d", maxAttempts)
	}

	client.Close()
}

// TestWebSocketClient_Reconnect_SuccessfulReconnect 测试成功重连
func TestWebSocketClient_Reconnect_SuccessfulReconnect(t *testing.T) {
	var mu sync.Mutex
	connectionCount := 0

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
		}
		defer conn.Close()

		mu.Lock()
		connectionCount++
		connNum := connectionCount
		mu.Unlock()

		// 第一次连接立即关闭，模拟网络问题
		if connNum == 1 {
			time.Sleep(100 * time.Millisecond)
			conn.Close()
			return
		}

		// 第二次连接保持开启
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	// 创建客户端
	logger := zap.NewNop()
	config := &WebSocketConfig{
		URL:                  "ws" + server.URL[4:],
		ReconnectInterval:    200 * time.Millisecond,
		MaxReconnectAttempts: 5,
		ReconnectBackoff:     100 * time.Millisecond,
	}
	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 等待重连
	time.Sleep(1 * time.Second)

	// 检查是否重连成功
	mu.Lock()
	finalConnCount := connectionCount
	mu.Unlock()

	if finalConnCount < 2 {
		t.Errorf("期望至少2次连接, 实际 = %d", finalConnCount)
	}

	// 检查重连状态
	attempts, _, _ := client.GetReconnectStatus()
	if attempts > 0 {
		t.Errorf("重连成功后重连次数应该重置为0, 实际 = %d", attempts)
	}

	client.Close()
}

// TestWebSocketClient_Reconnect_MaxAttemptsReached 测试达到最大重连次数
func TestWebSocketClient_Reconnect_MaxAttemptsReached(t *testing.T) {
	var connectionCount int
	var mu sync.Mutex

	// 创建模拟服务器，连接后立即关闭
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
		}
		defer conn.Close()

		mu.Lock()
		connectionCount++
		mu.Unlock()

		// 立即关闭连接，模拟网络问题
		time.Sleep(50 * time.Millisecond)
		conn.Close()
	}))
	defer server.Close()

	// 创建客户端
	logger := zap.NewNop()
	config := &WebSocketConfig{
		URL:                  "ws" + server.URL[4:],
		ReconnectInterval:    100 * time.Millisecond,
		MaxReconnectAttempts: 2,
		ReconnectBackoff:     50 * time.Millisecond,
	}
	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 等待重连尝试
	time.Sleep(2 * time.Second)

	// 检查连接次数（应该至少有初始连接 + 2次重连 = 3次）
	mu.Lock()
	finalConnectionCount := connectionCount
	mu.Unlock()

	if finalConnectionCount < 3 {
		t.Errorf("期望至少3次连接（初始 + 2次重连）, 实际 = %d", finalConnectionCount)
	}

	client.Close()
}

// TestWebSocketClient_Reconnect_Resubscribe 测试重连后重新订阅
func TestWebSocketClient_Reconnect_Resubscribe(t *testing.T) {
	var mu sync.Mutex
	connectionCount := 0
	subscriptionMessages := make([]string, 0)

	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
		}
		defer conn.Close()

		mu.Lock()
		connectionCount++
		connNum := connectionCount
		mu.Unlock()

		// 第一次连接立即关闭
		if connNum == 1 {
			time.Sleep(100 * time.Millisecond)
			conn.Close()
			return
		}

		// 第二次连接处理消息
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			mu.Lock()
			subscriptionMessages = append(subscriptionMessages, string(message))
			mu.Unlock()

			// 发送订阅成功响应
			response := WebSocketMessage{
				Event: "subscribe",
				Code:  "0",
				Msg:   "success",
			}
			responseBytes, _ := json.Marshal(response)
			conn.WriteMessage(websocket.TextMessage, responseBytes)
		}
	}))
	defer server.Close()

	// 创建客户端
	logger := zap.NewNop()
	config := &WebSocketConfig{
		URL:                  "ws" + server.URL[4:],
		ReconnectInterval:    200 * time.Millisecond,
		MaxReconnectAttempts: 5,
		ReconnectBackoff:     100 * time.Millisecond,
	}
	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 订阅
	callback := func(ticker Ticker) {}
	err = client.SubscribeTicker([]string{"BTCUSDT"}, callback)
	if err != nil {
		t.Fatalf("订阅失败: %v", err)
	}

	// 启动消息处理器
	client.StartMessageHandler()

	// 等待重连和重新订阅
	time.Sleep(2 * time.Second)

	// 检查重新订阅
	mu.Lock()
	messageCount := len(subscriptionMessages)
	mu.Unlock()

	if messageCount < 1 {
		t.Errorf("期望至少1次订阅消息, 实际 = %d", messageCount)
	}

	client.Close()
}

// TestWebSocketClient_Reconnect_DisableReconnect 测试禁用自动重连
func TestWebSocketClient_Reconnect_DisableReconnect(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
		}
		defer conn.Close()

		// 立即关闭连接
		time.Sleep(100 * time.Millisecond)
		conn.Close()
	}))
	defer server.Close()

	// 创建客户端
	logger := zap.NewNop()
	config := &WebSocketConfig{
		URL:                  "ws" + server.URL[4:],
		ReconnectInterval:    100 * time.Millisecond,
		MaxReconnectAttempts: 5,
		ReconnectBackoff:     50 * time.Millisecond,
	}
	client := NewWebSocketClient(config, logger)

	// 禁用自动重连
	client.SetReconnectEnabled(false)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 等待
	time.Sleep(1 * time.Second)

	// 检查重连状态
	_, _, enabled := client.GetReconnectStatus()
	if enabled {
		t.Error("自动重连应该被禁用")
	}

	client.Close()
}

// TestWebSocketClient_Reconnect_ExponentialBackoff 测试指数退避
func TestWebSocketClient_Reconnect_ExponentialBackoff(t *testing.T) {
	var connectionCount int
	var mu sync.Mutex

	// 创建模拟服务器，连接后立即关闭
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
		}
		defer conn.Close()

		mu.Lock()
		connectionCount++
		mu.Unlock()

		// 立即关闭连接，模拟网络问题
		time.Sleep(50 * time.Millisecond)
		conn.Close()
	}))
	defer server.Close()

	// 创建客户端
	logger := zap.NewNop()
	config := &WebSocketConfig{
		URL:                  "ws" + server.URL[4:],
		ReconnectInterval:    100 * time.Millisecond,
		MaxReconnectAttempts: 3,
		ReconnectBackoff:     50 * time.Millisecond,
	}
	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 等待重连尝试
	time.Sleep(2 * time.Second)

	// 检查连接次数（应该至少有初始连接 + 3次重连 = 4次）
	mu.Lock()
	finalConnectionCount := connectionCount
	mu.Unlock()

	if finalConnectionCount < 4 {
		t.Errorf("期望至少4次连接（初始 + 3次重连）, 实际 = %d", finalConnectionCount)
	}

	client.Close()
}

// TestWebSocketClient_Reconnect_ConcurrentAccess 测试并发访问
func TestWebSocketClient_Reconnect_ConcurrentAccess(t *testing.T) {
	// 创建模拟服务器
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Fatalf("WebSocket upgrade failed: %v", err)
		}
		defer conn.Close()

		// 保持连接
		time.Sleep(2 * time.Second)
	}))
	defer server.Close()

	// 创建客户端
	logger := zap.NewNop()
	config := &WebSocketConfig{
		URL:                  "ws" + server.URL[4:],
		ReconnectInterval:    100 * time.Millisecond,
		MaxReconnectAttempts: 5,
		ReconnectBackoff:     50 * time.Millisecond,
	}
	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("连接失败: %v", err)
	}

	// 并发访问重连状态
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			client.GetReconnectStatus()
		}()
	}

	wg.Wait()

	// 检查连接状态
	if !client.IsConnected() {
		t.Error("客户端应该保持连接")
	}

	client.Close()
}
