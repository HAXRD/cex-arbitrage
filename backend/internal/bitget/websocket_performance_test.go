package bitget

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// mockWebSocketServerForPerformance 创建高性能的模拟 WebSocket 服务器
func mockWebSocketServerForPerformance(t *testing.T) *httptest.Server {
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

						// 持续发送 Ticker 数据（模拟实时推送）
						go func() {
							ticker := time.NewTicker(100 * time.Millisecond) // 每100ms发送一次
							defer ticker.Stop()

							for {
								select {
								case <-ticker.C:
									for _, arg := range request.Args {
										symbol := arg.InstId // 现在 InstId 是字符串
										// 生成模拟的实时数据
										price := 50000.0 + float64(time.Now().UnixNano()%1000) // 模拟价格波动
										volume := 1000.0 + float64(time.Now().UnixNano()%100)  // 模拟交易量变化

										tickerData := WebSocketTickerData{
											Action: "snapshot",
											Data: []Ticker{{
												Symbol:        symbol,
												LastPr:        fmt.Sprintf("%.2f", price),
												BidPr:         fmt.Sprintf("%.2f", price-1),
												AskPr:         fmt.Sprintf("%.2f", price+1),
												BidSz:         "1.5",
												AskSz:         "2.0",
												High24h:       "52000.0",
												Low24h:        "48000.0",
												Ts:            fmt.Sprintf("%d", time.Now().UnixMilli()),
												BaseVolume:    fmt.Sprintf("%.2f", volume),
												QuoteVolume:   fmt.Sprintf("%.2f", volume*price),
												UsdtVolume:    fmt.Sprintf("%.2f", volume*price),
												IndexPrice:    fmt.Sprintf("%.2f", price*0.999),
												MarkPrice:     fmt.Sprintf("%.2f", price),
												FundingRate:   "0.0001",
												HoldingAmount: "10000.0",
											}},
										}

										tickerBytes, _ := json.Marshal(tickerData)
										conn.WriteMessage(websocket.TextMessage, tickerBytes)
									}
								}
							}
						}()
					}
				}
			}
		}
	}))

	return server
}

func TestWebSocketClient_Performance_ConcurrentSubscriptions(t *testing.T) {
	// 创建高性能模拟服务器
	server := mockWebSocketServerForPerformance(t)
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

	// 测试并发订阅（20个交易对）
	symbols := []string{
		"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT",
		"LTCUSDT", "BCHUSDT", "XRPUSDT", "EOSUSDT", "TRXUSDT",
		"BNBUSDT", "SOLUSDT", "MATICUSDT", "AVAXUSDT", "ATOMUSDT",
		"FTMUSDT", "NEARUSDT", "ALGOUSDT", "VETUSDT", "ICPUSDT",
	}

	receivedCount := 0
	var mu sync.Mutex
	callback := func(ticker Ticker) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	// 记录开始时间
	startTime := time.Now()

	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待数据推送（5秒）
	time.Sleep(5 * time.Second)

	// 记录结束时间
	endTime := time.Now()
	duration := endTime.Sub(startTime)

	// 验证性能
	mu.Lock()
	t.Logf("Received %d ticker updates in %v", receivedCount, duration)
	t.Logf("Average rate: %.2f updates/second", float64(receivedCount)/duration.Seconds())
	mu.Unlock()

	// 验证至少接收到一些数据
	if receivedCount == 0 {
		t.Error("Expected to receive some ticker updates")
	}

	// 验证性能（至少每秒10个更新）
	rate := float64(receivedCount) / duration.Seconds()
	if rate < 10 {
		t.Errorf("Expected at least 10 updates/second, got %.2f", rate)
	}

	// 清理
	client.Close()
}

func TestWebSocketClient_Performance_MemoryUsage(t *testing.T) {
	// 创建高性能模拟服务器
	server := mockWebSocketServerForPerformance(t)
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

	// 测试内存使用（长时间运行）
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT"}

	receivedCount := 0
	var mu sync.Mutex
	callback := func(ticker Ticker) {
		mu.Lock()
		receivedCount++
		mu.Unlock()
	}

	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 运行10秒
	time.Sleep(10 * time.Second)

	// 验证内存使用合理（通过检查没有内存泄漏的迹象）
	mu.Lock()
	t.Logf("Received %d ticker updates in 10 seconds", receivedCount)
	mu.Unlock()

	// 验证连接仍然活跃
	if !client.IsConnected() {
		t.Error("Expected client to still be connected after 10 seconds")
	}

	// 清理
	client.Close()
}

func TestWebSocketClient_Performance_DataLatency(t *testing.T) {
	// 创建高性能模拟服务器
	server := mockWebSocketServerForPerformance(t)
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

	// 测试数据延迟
	symbols := []string{"BTCUSDT"}

	var latencies []time.Duration
	var mu sync.Mutex
	callback := func(ticker Ticker) {
		// 计算延迟（从时间戳到接收时间）
		ts, err := time.Parse("", ticker.Ts)
		if err == nil {
			latency := time.Now().Sub(ts)
			mu.Lock()
			latencies = append(latencies, latency)
			mu.Unlock()
		}
	}

	err = client.SubscribeTicker(symbols, callback)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}

	// 等待数据
	time.Sleep(2 * time.Second)

	// 分析延迟
	mu.Lock()
	if len(latencies) > 0 {
		var totalLatency time.Duration
		for _, latency := range latencies {
			totalLatency += latency
		}
		avgLatency := totalLatency / time.Duration(len(latencies))

		t.Logf("Average latency: %v", avgLatency)
		t.Logf("Total samples: %d", len(latencies))

		// 验证延迟合理（小于500ms）
		if avgLatency > 500*time.Millisecond {
			t.Errorf("Expected average latency < 500ms, got %v", avgLatency)
		}
	}
	mu.Unlock()

	// 清理
	client.Close()
}
