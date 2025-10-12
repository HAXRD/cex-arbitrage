package data_collection

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// MockWebSocketServer 模拟WebSocket服务器
type MockWebSocketServer struct {
	server   *http.Server
	upgrader websocket.Upgrader
	clients  map[*websocket.Conn]bool
	mu       sync.RWMutex
	writeMu  map[*websocket.Conn]*sync.Mutex // 每个连接的写入锁
	logger   *zap.Logger
	port     string
}

// MockWebSocketClient 模拟WebSocket客户端
type MockWebSocketClient struct {
	*WebSocketClientImpl
	server *MockWebSocketServer
}

// NewMockWebSocketServer 创建模拟WebSocket服务器
func NewMockWebSocketServer(port string, logger *zap.Logger) *MockWebSocketServer {
	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	return &MockWebSocketServer{
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源
			},
		},
		clients: make(map[*websocket.Conn]bool),
		writeMu: make(map[*websocket.Conn]*sync.Mutex),
		logger:  logger,
		port:    port,
	}
}

// Start 启动模拟服务器
func (s *MockWebSocketServer) Start() error {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: mux,
	}

	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("模拟服务器启动失败", zap.Error(err))
		}
	}()

	// 等待服务器启动
	time.Sleep(100 * time.Millisecond)
	return nil
}

// Stop 停止模拟服务器
func (s *MockWebSocketServer) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}

// handleWebSocket 处理WebSocket连接
func (s *MockWebSocketServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket升级失败", zap.Error(err))
		return
	}
	defer conn.Close()

	// 注册客户端
	s.mu.Lock()
	s.clients[conn] = true
	s.writeMu[conn] = &sync.Mutex{}
	s.mu.Unlock()

	// 发送欢迎消息
	welcomeMsg := map[string]interface{}{
		"type":      "welcome",
		"message":   "Connected to mock server",
		"timestamp": time.Now().Unix(),
	}

	// 使用写入锁
	s.writeMu[conn].Lock()
	err = conn.WriteJSON(welcomeMsg)
	s.writeMu[conn].Unlock()
	if err != nil {
		s.logger.Error("发送欢迎消息失败", zap.Error(err))
	}

	// 处理消息
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			s.logger.Debug("客户端断开连接", zap.Error(err))
			break
		}

		// 处理订阅请求
		var request map[string]interface{}
		if err := json.Unmarshal(message, &request); err == nil {
			s.handleSubscriptionRequest(conn, request)
		}
	}

	// 移除客户端
	s.mu.Lock()
	delete(s.clients, conn)
	delete(s.writeMu, conn)
	s.mu.Unlock()
}

// handleSubscriptionRequest 处理订阅请求
func (s *MockWebSocketServer) handleSubscriptionRequest(conn *websocket.Conn, request map[string]interface{}) {
	action, ok := request["action"].(string)
	if !ok {
		return
	}

	symbol, _ := request["symbol"].(string)
	channel, _ := request["channel"].(string)

	response := map[string]interface{}{
		"type":      "subscription",
		"action":    action,
		"symbol":    symbol,
		"channel":   channel,
		"success":   true,
		"timestamp": time.Now().Unix(),
	}

	if err := conn.WriteJSON(response); err != nil {
		s.logger.Error("发送订阅响应失败", zap.Error(err))
	}

	// 如果是订阅请求，开始发送模拟数据
	if action == "subscribe" && symbol != "" {
		go s.sendMockData(conn, symbol)
	}
}

// sendMockData 发送模拟数据
func (s *MockWebSocketServer) sendMockData(conn *websocket.Conn, symbol string) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// 检查连接是否仍然活跃
			s.mu.RLock()
			_, exists := s.clients[conn]
			writeMu, hasMu := s.writeMu[conn]
			s.mu.RUnlock()

			if !exists || !hasMu {
				return
			}

			// 发送模拟价格数据
			priceData := map[string]interface{}{
				"type":      "ticker",
				"symbol":    symbol,
				"price":     50000.0 + float64(time.Now().Unix()%1000),
				"volume":    1000.0,
				"timestamp": time.Now().Unix(),
			}

			// 使用写入锁
			writeMu.Lock()
			err := conn.WriteJSON(priceData)
			writeMu.Unlock()
			if err != nil {
				s.logger.Debug("发送模拟数据失败", zap.Error(err))
				return
			}
		}
	}
}

// BroadcastMessage 广播消息给所有客户端
func (s *MockWebSocketServer) BroadcastMessage(message interface{}) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for conn := range s.clients {
		if err := conn.WriteJSON(message); err != nil {
			s.logger.Error("广播消息失败", zap.Error(err))
		}
	}
}

// GetClientCount 获取客户端数量
func (s *MockWebSocketServer) GetClientCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.clients)
}

// NewMockWebSocketClient 创建模拟WebSocket客户端
func NewMockWebSocketClient(server *MockWebSocketServer, config *WebSocketConfig, logger *zap.Logger) *MockWebSocketClient {
	if config == nil {
		config = DefaultWebSocketConfig()
	}
	config.URL = "ws://localhost:" + server.port + "/ws"

	client := &MockWebSocketClient{
		WebSocketClientImpl: NewWebSocketClient(config, logger),
		server:              server,
	}

	return client
}

// Connect 重写连接方法以使用模拟服务器
func (c *MockWebSocketClient) Connect(ctx context.Context) error {
	// 使用父类的连接逻辑，但URL已经设置为模拟服务器
	return c.WebSocketClientImpl.Connect(ctx)
}

// SimulateConnectionError 模拟连接错误
func (c *MockWebSocketClient) SimulateConnectionError() {
	c.mu.Lock()
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	// 先设置为未连接状态，避免messageLoop重复触发错误处理
	c.connected = false
	c.mu.Unlock()

	// 等待一小段时间，让messageLoop检测到连接关闭
	time.Sleep(10 * time.Millisecond)

	// 触发错误处理器，激活自动重连
	c.handleConnectionError(fmt.Errorf("simulated connection error"))
}

// SimulateMessage 模拟接收消息
func (c *MockWebSocketClient) SimulateMessage(data []byte) {
	c.handleMessage(data)
}

// SimulateError 模拟错误
func (c *MockWebSocketClient) SimulateError(err error) {
	c.handleConnectionError(err)
}
