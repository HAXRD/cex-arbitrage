package bitget

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// mockWebSocketServer 创建模拟 WebSocket 服务器
func mockWebSocketServer(t *testing.T) *httptest.Server {
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

		// 模拟 WebSocket 消息处理
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					t.Errorf("WebSocket error: %v", err)
				}
				break
			}

			// 解析订阅消息
			var wsRequest WebSocketRequest
			if err := json.Unmarshal(message, &wsRequest); err != nil {
				t.Errorf("Failed to unmarshal message: %v", err)
				continue
			}

			// 发送订阅成功响应
			if wsRequest.Op == "subscribe" {
				response := WebSocketMessage{
					Event: "subscribe",
					Code:  "0",
					Msg:   "success",
				}
				responseBytes, _ := json.Marshal(response)
				conn.WriteMessage(websocket.TextMessage, responseBytes)
			}

			// 模拟发送 Ticker 数据
			if wsRequest.Op == "subscribe" {
				// 等待一下再发送数据
				time.Sleep(100 * time.Millisecond)

				// 发送模拟的 Ticker 数据
				tickerData := WebSocketTickerData{
					Action: "snapshot",
					Data: []Ticker{{
						Symbol:        "BTCUSDT",
						LastPr:        "50000.0",
						BidPr:         "49999.0",
						AskPr:         "50001.0",
						BidSz:         "1.5",
						AskSz:         "2.0",
						High24h:       "52000.0",
						Low24h:        "48000.0",
						Ts:            "1678886400000",
						BaseVolume:    "1000.0",
						QuoteVolume:   "50000000.0",
						UsdtVolume:    "50000000.0",
						IndexPrice:    "49950.0",
						MarkPrice:     "50000.0",
						FundingRate:   "0.0001",
						HoldingAmount: "10000.0",
					}},
				}

				tickerBytes, _ := json.Marshal(tickerData)
				conn.WriteMessage(websocket.TextMessage, tickerBytes)
			}
		}
	}))

	return server
}

func TestNewWebSocketClient(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	config := DefaultWebSocketConfig()

	client := NewWebSocketClient(config, logger)

	if client == nil {
		t.Fatal("Expected client to be created")
	}

	if client.url != config.URL {
		t.Errorf("Expected URL = %s, got %s", config.URL, client.url)
	}

	if client.logger == nil {
		t.Error("Expected logger to be set")
	}

	if client.subscribers == nil {
		t.Error("Expected subscribers map to be initialized")
	}
}

func TestWebSocketClient_Connect(t *testing.T) {
	// 创建模拟 WebSocket 服务器
	server := mockWebSocketServer(t)
	defer server.Close()

	// 将 HTTP 服务器 URL 转换为 WebSocket URL
	wsURL := "ws" + server.URL[4:] // 将 http:// 替换为 ws://

	logger, _ := zap.NewDevelopment()
	config := &WebSocketConfig{
		URL:             wsURL,
		PingInterval:    30 * time.Second,
		PongWait:        60 * time.Second,
		WriteWait:       10 * time.Second,
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	client := NewWebSocketClient(config, logger)

	// 测试连接
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := client.Connect(ctx)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	// 验证连接状态
	if !client.IsConnected() {
		t.Error("Expected client to be connected")
	}

	// 清理
	client.Close()
}

func TestWebSocketClient_Close(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := NewWebSocketClient(nil, logger)

	// 测试关闭未连接的客户端
	err := client.Close()
	if err != nil {
		t.Errorf("Expected no error when closing unconnected client, got: %v", err)
	}
}

func TestWebSocketClient_SubscribeTicker(t *testing.T) {
	// 创建模拟 WebSocket 服务器
	server := mockWebSocketServer(t)
	defer server.Close()

	// 将 HTTP 服务器 URL 转换为 WebSocket URL
	wsURL := "ws" + server.URL[4:]

	logger, _ := zap.NewDevelopment()
	config := &WebSocketConfig{
		URL:             wsURL,
		PingInterval:    30 * time.Second,
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

	// 启动消息处理器
	client.StartMessageHandler()

	// 测试订阅
	var receivedTicker Ticker
	callback := func(ticker Ticker) {
		receivedTicker = ticker
	}

	symbols := []string{"BTCUSDT"}
	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待数据
	time.Sleep(500 * time.Millisecond)

	// 验证接收到的数据
	if receivedTicker.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol = BTCUSDT, got %s", receivedTicker.Symbol)
	}

	// 清理
	client.Close()
}

func TestWebSocketClient_Unsubscribe(t *testing.T) {
	// 创建模拟 WebSocket 服务器
	server := mockWebSocketServer(t)
	defer server.Close()

	// 将 HTTP 服务器 URL 转换为 WebSocket URL
	wsURL := "ws" + server.URL[4:]

	logger, _ := zap.NewDevelopment()
	config := &WebSocketConfig{
		URL:             wsURL,
		PingInterval:    30 * time.Second,
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

	// 测试取消订阅
	symbols := []string{"BTCUSDT"}
	err = client.Unsubscribe(symbols)
	if err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// 清理
	client.Close()
}

func TestWebSocketClient_IsConnected(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := NewWebSocketClient(nil, logger)

	// 初始状态应该是未连接
	if client.IsConnected() {
		t.Error("Expected client to be disconnected initially")
	}
}

func TestDefaultWebSocketConfig(t *testing.T) {
	config := DefaultWebSocketConfig()

	if config.URL != "wss://ws.bitget.com/spot/v1/stream" {
		t.Errorf("Expected default URL, got %s", config.URL)
	}

	if config.PingInterval != 30*time.Second {
		t.Errorf("Expected default ping interval, got %v", config.PingInterval)
	}

	if config.PongWait != 60*time.Second {
		t.Errorf("Expected default pong wait, got %v", config.PongWait)
	}
}
