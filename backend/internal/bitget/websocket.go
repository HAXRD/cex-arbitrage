package bitget

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketClient WebSocket 客户端
type WebSocketClient struct {
	conn        *websocket.Conn
	url         string
	logger      *zap.Logger
	mu          sync.RWMutex
	subscribers map[string]TickerCallback
	done        chan struct{}
	readPump    chan []byte
	writePump   chan []byte
	connected   bool
	reconnect   bool
	// 心跳相关字段
	lastPong    time.Time
	pongTimeout time.Duration
	// 自动重连相关字段
	reconnectInterval    time.Duration // 重连间隔
	maxReconnectAttempts int           // 最大重连次数
	reconnectAttempts    int           // 当前重连次数
	reconnectBackoff     time.Duration // 重连退避时间
	reconnectTimer       *time.Timer   // 重连定时器
	reconnectMutex       sync.Mutex    // 重连互斥锁
}

// TickerCallback 已在 types.go 中定义

// WebSocketConfig WebSocket 配置
type WebSocketConfig struct {
	URL                  string        `yaml:"url"`
	PingInterval         time.Duration `yaml:"ping_interval"`
	PongWait             time.Duration `yaml:"pong_wait"`
	WriteWait            time.Duration `yaml:"write_wait"`
	ReadBufferSize       int           `yaml:"read_buffer_size"`
	WriteBufferSize      int           `yaml:"write_buffer_size"`
	ReconnectInterval    time.Duration `yaml:"reconnect_interval"`
	MaxReconnectAttempts int           `yaml:"max_reconnect_attempts"`
	ReconnectBackoff     time.Duration `yaml:"reconnect_backoff"`
}

// DefaultWebSocketConfig 默认 WebSocket 配置
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		URL:                  "wss://ws.bitget.com/spot/v1/stream",
		PingInterval:         30 * time.Second,
		PongWait:             60 * time.Second,
		WriteWait:            10 * time.Second,
		ReadBufferSize:       1024,
		WriteBufferSize:      1024,
		ReconnectInterval:    5 * time.Second,
		MaxReconnectAttempts: 10,
		ReconnectBackoff:     2 * time.Second,
	}
}

// NewWebSocketClient 创建新的 WebSocket 客户端
func NewWebSocketClient(config *WebSocketConfig, logger *zap.Logger) *WebSocketClient {
	if config == nil {
		config = DefaultWebSocketConfig()
	}

	return &WebSocketClient{
		url:                  config.URL,
		logger:               logger,
		subscribers:          make(map[string]TickerCallback),
		done:                 make(chan struct{}),
		readPump:             make(chan []byte, 256),
		writePump:            make(chan []byte, 256),
		connected:            false,
		reconnect:            true,
		lastPong:             time.Now(),
		pongTimeout:          config.PongWait,
		reconnectInterval:    config.ReconnectInterval,
		maxReconnectAttempts: config.MaxReconnectAttempts,
		reconnectAttempts:    0,
		reconnectBackoff:     config.ReconnectBackoff,
	}
}

// Connect 连接到 WebSocket 服务器
func (c *WebSocketClient) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.connected {
		return fmt.Errorf("WebSocket already connected")
	}

	// 解析 WebSocket URL
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("invalid WebSocket URL: %w", err)
	}

	// 建立 WebSocket 连接
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, u.String(), nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	c.conn = conn
	c.connected = true

	// 启动消息处理 goroutine
	go c.readPumpGoroutine()
	go c.writePumpGoroutine()
	c.StartMessageHandler()

	c.logger.Info("WebSocket connected successfully", zap.String("url", c.url))
	return nil
}

// Close 关闭 WebSocket 连接
func (c *WebSocketClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return nil
	}

	c.reconnect = false
	c.connected = false

	// 关闭 done channel（如果还没有关闭）
	select {
	case <-c.done:
		// 已经关闭
	default:
		close(c.done)
	}

	if c.conn != nil {
		err := c.conn.Close()
		if err != nil {
			c.logger.Error("failed to close WebSocket connection", zap.Error(err))
			return err
		}
	}

	c.logger.Info("WebSocket connection closed")
	return nil
}

// IsConnected 检查连接状态
func (c *WebSocketClient) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// readPumpGoroutine 读取 WebSocket 消息的 goroutine
func (c *WebSocketClient) readPumpGoroutine() {
	defer func() {
		c.mu.Lock()
		c.connected = false
		reconnect := c.reconnect
		c.mu.Unlock()

		// 启动自动重连
		if reconnect {
			go c.startReconnect()
		}
	}()

	// 设置初始读取超时
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))

	// 设置 pong 处理器
	c.conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.lastPong = time.Now()
		c.mu.Unlock()

		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		c.logger.Debug("received pong from WebSocket server")
		return nil
	})

	// 启动心跳超时检测
	go c.heartbeatTimeoutChecker()

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					c.logger.Error("WebSocket read error", zap.Error(err))
				}
				return
			}

			select {
			case c.readPump <- message:
			case <-c.done:
				return
			}
		}
	}
}

// heartbeatTimeoutChecker 心跳超时检测器
func (c *WebSocketClient) heartbeatTimeoutChecker() {
	ticker := time.NewTicker(5 * time.Second) // 每5秒检查一次
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			c.mu.RLock()
			connected := c.connected
			lastPong := c.lastPong
			timeout := c.pongTimeout
			c.mu.RUnlock()

			if !connected {
				return
			}

			// 检查是否超时
			if time.Since(lastPong) > timeout {
				c.logger.Error("WebSocket heartbeat timeout",
					zap.Duration("timeout", timeout),
					zap.Time("lastPong", lastPong),
					zap.Duration("elapsed", time.Since(lastPong)),
				)

				// 关闭连接
				c.Close()
				return
			}
		}
	}
}

// writePumpGoroutine 写入 WebSocket 消息的 goroutine
func (c *WebSocketClient) writePumpGoroutine() {
	pingTicker := time.NewTicker(30 * time.Second)
	defer pingTicker.Stop()

	for {
		select {
		case <-c.done:
			return
		case message := <-c.writePump:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.TextMessage, message); err != nil {
				c.logger.Error("WebSocket write error", zap.Error(err))
				return
			}
		case <-pingTicker.C:
			// 发送心跳 ping
			c.sendPing()
		}
	}
}

// sendPing 发送心跳 ping
func (c *WebSocketClient) sendPing() {
	c.mu.RLock()
	connected := c.connected
	c.mu.RUnlock()

	if !connected {
		return
	}

	c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
		c.logger.Error("WebSocket ping error", zap.Error(err))
		return
	}

	c.logger.Debug("sent ping to WebSocket server")
}

// SubscribeTicker 订阅 Ticker 数据
func (c *WebSocketClient) SubscribeTicker(symbols []string, callback TickerCallback) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("WebSocket not connected")
	}

	// 为每个 symbol 创建单独的订阅
	subscriptions := make([]WebSocketSubscription, 0, len(symbols))
	for _, symbol := range symbols {
		subscriptions = append(subscriptions, WebSocketSubscription{
			InstType: "USDT-FUTURES",
			Channel:  "ticker",
			InstId:   symbol, // 单个 symbol，而不是数组
		})
	}

	request := WebSocketRequest{
		Op:   "subscribe",
		Args: subscriptions,
	}

	// 发送订阅消息
	msgBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal subscription message: %w", err)
	}

	select {
	case c.writePump <- msgBytes:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending subscription message")
	}

	// 存储回调函数
	for _, symbol := range symbols {
		c.subscribers[symbol] = callback
	}

	c.logger.Info("subscribed to ticker data", zap.Strings("symbols", symbols))
	return nil
}

// Unsubscribe 取消订阅
func (c *WebSocketClient) Unsubscribe(symbols []string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if !c.connected {
		return fmt.Errorf("WebSocket not connected")
	}

	// 为每个 symbol 创建单独的取消订阅
	subscriptions := make([]WebSocketSubscription, 0, len(symbols))
	for _, symbol := range symbols {
		subscriptions = append(subscriptions, WebSocketSubscription{
			InstType: "USDT-FUTURES",
			Channel:  "ticker",
			InstId:   symbol, // 单个 symbol，而不是数组
		})
	}

	request := WebSocketRequest{
		Op:   "unsubscribe",
		Args: subscriptions,
	}

	// 发送取消订阅消息
	msgBytes, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("failed to marshal unsubscription message: %w", err)
	}

	select {
	case c.writePump <- msgBytes:
	case <-time.After(5 * time.Second):
		return fmt.Errorf("timeout sending unsubscription message")
	}

	// 移除回调函数
	for _, symbol := range symbols {
		delete(c.subscribers, symbol)
	}

	c.logger.Info("unsubscribed from ticker data", zap.Strings("symbols", symbols))
	return nil
}

// StartMessageHandler 启动消息处理器
func (c *WebSocketClient) StartMessageHandler() {
	go func() {
		for {
			select {
			case <-c.done:
				return
			case message := <-c.readPump:
				c.handleMessage(message)
			}
		}
	}()
}

// handleMessage 处理接收到的消息
func (c *WebSocketClient) handleMessage(data []byte) {
	// 首先尝试解析为订阅响应
	var response WebSocketMessage
	if err := json.Unmarshal(data, &response); err == nil && response.Event != "" {
		// 处理订阅响应
		if response.Event == "subscribe" {
			c.logger.Info("subscription successful")
			return
		} else if response.Event == "error" {
			c.logger.Error("subscription failed", zap.String("code", response.Code), zap.String("msg", response.Msg))
			return
		}
	}

	// 尝试解析为 Ticker 数据
	var tickerData WebSocketTickerData
	if err := json.Unmarshal(data, &tickerData); err == nil && len(tickerData.Data) > 0 {
		// 分发到回调函数
		c.mu.RLock()
		// 遍历数据数组
		for _, ticker := range tickerData.Data {
			callback, exists := c.subscribers[ticker.Symbol]
			if exists && callback != nil {
				callback(ticker)
			}
		}
		c.mu.RUnlock()
		return
	}

	// 如果都不匹配，记录原始消息（DEBUG级别）
	c.logger.Debug("received unknown message", zap.String("data", string(data)))
}

// startReconnect 启动自动重连
func (c *WebSocketClient) startReconnect() {
	c.reconnectMutex.Lock()
	defer c.reconnectMutex.Unlock()

	// 检查是否已经达到最大重连次数
	if c.reconnectAttempts >= c.maxReconnectAttempts {
		c.logger.Error("max reconnect attempts reached, giving up",
			zap.Int("attempts", c.reconnectAttempts),
			zap.Int("maxAttempts", c.maxReconnectAttempts),
		)
		return
	}

	// 计算重连延迟（指数退避）
	delay := c.reconnectInterval + time.Duration(c.reconnectAttempts)*c.reconnectBackoff
	attempt := c.reconnectAttempts + 1

	c.logger.Info("scheduling reconnect",
		zap.Duration("delay", delay),
		zap.Int("attempt", attempt),
		zap.Int("maxAttempts", c.maxReconnectAttempts),
	)

	// 设置重连定时器
	c.reconnectTimer = time.AfterFunc(delay, func() {
		c.performReconnect()
	})
}

// performReconnect 执行重连
func (c *WebSocketClient) performReconnect() {
	c.reconnectMutex.Lock()
	defer c.reconnectMutex.Unlock()

	// 检查是否已经连接
	if c.connected {
		return
	}

	// 递增重连次数
	c.reconnectAttempts++

	c.logger.Info("attempting to reconnect",
		zap.Int("attempt", c.reconnectAttempts),
		zap.Int("maxAttempts", c.maxReconnectAttempts),
	)

	// 创建新的 context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 尝试重连
	err := c.Connect(ctx)
	if err != nil {
		c.logger.Error("reconnect failed", zap.Error(err))
		// 继续重连
		go c.startReconnect()
		return
	}

	// 重连成功，重置重连次数
	c.reconnectAttempts = 0
	c.logger.Info("reconnected successfully")

	// 重新订阅所有之前的订阅
	c.resubscribeAll()
}

// resubscribeAll 重新订阅所有之前的订阅
func (c *WebSocketClient) resubscribeAll() {
	c.mu.RLock()
	symbols := make([]string, 0, len(c.subscribers))
	for symbol := range c.subscribers {
		symbols = append(symbols, symbol)
	}
	c.mu.RUnlock()

	if len(symbols) == 0 {
		return
	}

	// 为每个 symbol 创建单独的订阅
	subscriptions := make([]WebSocketSubscription, 0, len(symbols))
	for _, symbol := range symbols {
		subscriptions = append(subscriptions, WebSocketSubscription{
			InstType: "USDT-FUTURES",
			Channel:  "ticker",
			InstId:   symbol, // 单个 symbol，而不是数组
		})
	}

	request := WebSocketRequest{
		Op:   "subscribe",
		Args: subscriptions,
	}

	msgBytes, err := json.Marshal(request)
	if err != nil {
		c.logger.Error("failed to marshal resubscription message", zap.Error(err))
		return
	}

	select {
	case c.writePump <- msgBytes:
		c.logger.Info("resubscribed to ticker data", zap.Strings("symbols", symbols))
	case <-time.After(5 * time.Second):
		c.logger.Error("timeout sending resubscription message")
	}
}

// SetReconnectEnabled 设置是否启用自动重连
func (c *WebSocketClient) SetReconnectEnabled(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.reconnect = enabled
}

// GetReconnectStatus 获取重连状态
func (c *WebSocketClient) GetReconnectStatus() (attempts int, maxAttempts int, enabled bool) {
	c.reconnectMutex.Lock()
	defer c.reconnectMutex.Unlock()

	c.mu.RLock()
	enabled = c.reconnect
	c.mu.RUnlock()

	return c.reconnectAttempts, c.maxReconnectAttempts, enabled
}
