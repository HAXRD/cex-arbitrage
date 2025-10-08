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

// mockWebSocketServerForTicker 创建支持 Ticker 数据推送的模拟 WebSocket 服务器
func mockWebSocketServerForTicker(t *testing.T) *httptest.Server {
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
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// 处理 ping 消息
			if messageType == websocket.PingMessage {
				conn.WriteMessage(websocket.PongMessage, nil)
			}

			// 处理文本消息（订阅请求）
			if messageType == websocket.TextMessage {
				var request WebSocketRequest
				if err := json.Unmarshal(message, &request); err == nil {
					if request.Op == "subscribe" {
						// 发送订阅成功响应
						response := WebSocketMessage{
							Event: "subscribe",
							Code:  "0",
							Msg:   "success",
						}
						responseBytes, _ := json.Marshal(response)
						conn.WriteMessage(websocket.TextMessage, responseBytes)

						// 模拟发送 Ticker 数据
						for _, arg := range request.Args {
							symbol := arg.InstId // 现在 InstId 是字符串
							// 等待一下再发送数据
							time.Sleep(100 * time.Millisecond)

							// 发送模拟的 Ticker 数据
							tickerData := WebSocketTickerData{
								Action: "snapshot",
								Data: []Ticker{{
									Symbol:        symbol,
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
					} else if request.Op == "unsubscribe" {
						// 发送取消订阅成功响应
						response := WebSocketMessage{
							Event: "unsubscribe",
							Code:  "0",
							Msg:   "success",
						}
						responseBytes, _ := json.Marshal(response)
						conn.WriteMessage(websocket.TextMessage, responseBytes)
					}
				}
			}
		}
	}))

	return server
}

func TestWebSocketClient_SubscribeTicker_Single(t *testing.T) {
	// 创建模拟服务器
	server := mockWebSocketServerForTicker(t)
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

	// 测试单个订阅
	var receivedTicker Ticker
	var mu sync.Mutex
	callback := func(ticker Ticker) {
		mu.Lock()
		receivedTicker = ticker
		mu.Unlock()
	}

	symbols := []string{"BTCUSDT"}
	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待数据
	time.Sleep(500 * time.Millisecond)

	// 验证接收到的数据
	mu.Lock()
	if receivedTicker.Symbol != "BTCUSDT" {
		t.Errorf("Expected symbol = BTCUSDT, got %s", receivedTicker.Symbol)
	}
	if receivedTicker.LastPr != "50000.0" {
		t.Errorf("Expected LastPr = 50000.0, got %s", receivedTicker.LastPr)
	}
	if receivedTicker.BaseVolume != "1000.0" {
		t.Errorf("Expected BaseVolume = 1000.0, got %s", receivedTicker.BaseVolume)
	}
	mu.Unlock()

	// 清理
	client.Close()
}

func TestWebSocketClient_SubscribeTicker_Batch(t *testing.T) {
	// 创建模拟服务器
	server := mockWebSocketServerForTicker(t)
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

	// 测试批量订阅（10个交易对）
	symbols := []string{
		"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT",
		"LTCUSDT", "BCHUSDT", "XRPUSDT", "EOSUSDT", "TRXUSDT",
	}

	receivedTickers := make(map[string]Ticker)
	var mu sync.Mutex
	callback := func(ticker Ticker) {
		mu.Lock()
		receivedTickers[ticker.Symbol] = ticker
		mu.Unlock()
	}

	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待所有数据
	time.Sleep(2 * time.Second)

	// 验证接收到的数据
	mu.Lock()
	if len(receivedTickers) != len(symbols) {
		t.Errorf("Expected %d tickers, got %d", len(symbols), len(receivedTickers))
	}

	for _, symbol := range symbols {
		if ticker, exists := receivedTickers[symbol]; !exists {
			t.Errorf("Expected ticker for %s, but not received", symbol)
		} else if ticker.Symbol != symbol {
			t.Errorf("Expected symbol = %s, got %s", symbol, ticker.Symbol)
		}
	}
	mu.Unlock()

	// 清理
	client.Close()
}

func TestWebSocketClient_UnsubscribeTicker(t *testing.T) {
	// 创建模拟服务器
	server := mockWebSocketServerForTicker(t)
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

	// 先订阅
	symbols := []string{"BTCUSDT", "ETHUSDT"}
	callback := func(ticker Ticker) {
		// 测试回调
	}

	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待订阅完成
	time.Sleep(200 * time.Millisecond)

	// 取消订阅
	err = client.Unsubscribe(symbols)
	if err != nil {
		t.Fatalf("Failed to unsubscribe: %v", err)
	}

	// 验证取消订阅成功
	time.Sleep(200 * time.Millisecond)

	// 清理
	client.Close()
}

func TestWebSocketClient_TickerDataParsing(t *testing.T) {
	// 创建模拟服务器
	server := mockWebSocketServerForTicker(t)
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

	// 测试数据解析
	var receivedTicker Ticker
	var mu sync.Mutex
	callback := func(ticker Ticker) {
		mu.Lock()
		receivedTicker = ticker
		mu.Unlock()
	}

	symbols := []string{"BTCUSDT"}
	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待数据
	time.Sleep(500 * time.Millisecond)

	// 验证数据解析
	mu.Lock()
	if receivedTicker.Symbol != "BTCUSDT" {
		t.Errorf("Expected Symbol = BTCUSDT, got %s", receivedTicker.Symbol)
	}
	if receivedTicker.LastPr != "50000.0" {
		t.Errorf("Expected LastPr = 50000.0, got %s", receivedTicker.LastPr)
	}
	if receivedTicker.BidPr != "49999.0" {
		t.Errorf("Expected BidPr = 49999.0, got %s", receivedTicker.BidPr)
	}
	if receivedTicker.AskPr != "50001.0" {
		t.Errorf("Expected AskPr = 50001.0, got %s", receivedTicker.AskPr)
	}
	if receivedTicker.BaseVolume != "1000.0" {
		t.Errorf("Expected BaseVolume = 1000.0, got %s", receivedTicker.BaseVolume)
	}
	if receivedTicker.QuoteVolume != "50000000.0" {
		t.Errorf("Expected QuoteVolume = 50000000.0, got %s", receivedTicker.QuoteVolume)
	}
	if receivedTicker.IndexPrice != "49950.0" {
		t.Errorf("Expected IndexPrice = 49950.0, got %s", receivedTicker.IndexPrice)
	}
	if receivedTicker.MarkPrice != "50000.0" {
		t.Errorf("Expected MarkPrice = 50000.0, got %s", receivedTicker.MarkPrice)
	}
	if receivedTicker.FundingRate != "0.0001" {
		t.Errorf("Expected FundingRate = 0.0001, got %s", receivedTicker.FundingRate)
	}
	mu.Unlock()

	// 清理
	client.Close()
}

func TestWebSocketClient_SubscriptionResponse(t *testing.T) {
	// 创建模拟服务器
	server := mockWebSocketServerForTicker(t)
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

	// 测试订阅响应处理
	callback := func(ticker Ticker) {
		// 测试回调
	}

	symbols := []string{"BTCUSDT"}
	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待订阅响应
	time.Sleep(300 * time.Millisecond)

	// 验证订阅成功（通过日志或状态）
	// 这里主要验证没有错误发生
	if !client.IsConnected() {
		t.Error("Expected client to still be connected after subscription")
	}

	// 清理
	client.Close()
}
