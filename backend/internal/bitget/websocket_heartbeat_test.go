package bitget

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// mockWebSocketServerWithHeartbeat 创建支持心跳的模拟 WebSocket 服务器
func mockWebSocketServerWithHeartbeat(t *testing.T) *httptest.Server {
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// 设置 pong 处理器
		conn.SetPongHandler(func(string) error {
			// 模拟服务器响应 pong
			return nil
		})

		// 处理消息
		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					t.Errorf("WebSocket error: %v", err)
				}
				break
			}

			// 处理 ping 消息
			if messageType == websocket.PingMessage {
				// 发送 pong 响应
				conn.WriteMessage(websocket.PongMessage, nil)
			}
		}
	}))

	return server
}

func TestWebSocketClient_Heartbeat(t *testing.T) {
	// 创建支持心跳的模拟服务器
	server := mockWebSocketServerWithHeartbeat(t)
	defer server.Close()

	// 将 HTTP 服务器 URL 转换为 WebSocket URL
	wsURL := "ws" + server.URL[4:]

	logger, _ := zap.NewDevelopment()
	config := &WebSocketConfig{
		URL:             wsURL,
		PingInterval:    1 * time.Second, // 1秒发送一次ping，便于测试
		PongWait:        5 * time.Second, // 5秒超时
		WriteWait:       10 * time.Second,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 验证连接状态
	if !client.IsConnected() {
		t.Error("Expected client to be connected")
	}

	// 等待心跳发送和响应
	time.Sleep(3 * time.Second)

	// 验证连接仍然活跃
	if !client.IsConnected() {
		t.Error("Expected client to still be connected after heartbeat")
	}

	// 清理
	client.Close()
}

func TestWebSocketClient_HeartbeatTimeout(t *testing.T) {
	// 创建一个不响应 pong 的服务器
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// 不设置 pong 处理器，模拟服务器不响应 pong
		for {
			_, _, err := conn.ReadMessage()
			if err != nil {
				break
			}
		}
	}))
	defer server.Close()

	// 将 HTTP 服务器 URL 转换为 WebSocket URL
	wsURL := "ws" + server.URL[4:]

	logger, _ := zap.NewDevelopment()
	config := &WebSocketConfig{
		URL:             wsURL,
		PingInterval:    1 * time.Second, // 1秒发送一次ping
		PongWait:        2 * time.Second, // 2秒超时，便于测试
		WriteWait:       10 * time.Second,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 验证初始连接状态
	if !client.IsConnected() {
		t.Error("Expected client to be connected initially")
	}

	// 等待心跳超时（需要等待超时检测器运行）
	time.Sleep(7 * time.Second)

	// 验证连接已断开
	if client.IsConnected() {
		t.Error("Expected client to be disconnected due to heartbeat timeout")
	}

	// 清理
	client.Close()
}

func TestWebSocketClient_SendPing(t *testing.T) {
	// 创建支持心跳的模拟服务器
	server := mockWebSocketServerWithHeartbeat(t)
	defer server.Close()

	// 将 HTTP 服务器 URL 转换为 WebSocket URL
	wsURL := "ws" + server.URL[4:]

	logger, _ := zap.NewDevelopment()
	config := &WebSocketConfig{
		URL:             wsURL,
		PingInterval:    30 * time.Second, // 不自动发送ping
		PongWait:        60 * time.Second,
		WriteWait:       10 * time.Second,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 手动发送 ping
	client.sendPing()

	// 等待一下
	time.Sleep(100 * time.Millisecond)

	// 验证连接仍然活跃
	if !client.IsConnected() {
		t.Error("Expected client to still be connected after manual ping")
	}

	// 清理
	client.Close()
}

func TestWebSocketClient_HeartbeatWithSubscription(t *testing.T) {
	// 创建支持心跳和订阅的模拟服务器
	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			t.Errorf("WebSocket upgrade failed: %v", err)
			return
		}
		defer conn.Close()

		// 设置 pong 处理器
		conn.SetPongHandler(func(string) error {
			return nil
		})

		// 处理消息
		for {
			messageType, _, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 处理 ping 消息
			if messageType == websocket.PingMessage {
				conn.WriteMessage(websocket.PongMessage, nil)
			}

			// 处理文本消息（订阅请求）
			if messageType == websocket.TextMessage {
				// 发送订阅成功响应
				response := `{"event":"subscribe","code":"0","msg":"success"}`
				conn.WriteMessage(websocket.TextMessage, []byte(response))
			}
		}
	}))
	defer server.Close()

	// 将 HTTP 服务器 URL 转换为 WebSocket URL
	wsURL := "ws" + server.URL[4:]

	logger, _ := zap.NewDevelopment()
	config := &WebSocketConfig{
		URL:             wsURL,
		PingInterval:    1 * time.Second, // 1秒发送一次ping
		PongWait:        5 * time.Second, // 5秒超时
		WriteWait:       10 * time.Second,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	client := NewWebSocketClient(config, logger)

	// 连接
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 启动消息处理器
	client.StartMessageHandler()

	// 订阅
	callback := func(ticker Ticker) {
		// 测试回调
	}

	err = client.SubscribeTicker([]string{"BTCUSDT"}, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待心跳和订阅处理
	time.Sleep(3 * time.Second)

	// 验证连接仍然活跃
	if !client.IsConnected() {
		t.Error("Expected client to still be connected after heartbeat and subscription")
	}

	// 清理
	client.Close()
}
