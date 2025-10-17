package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketClient WebSocket客户端
type WebSocketClient struct {
	conn         *websocket.Conn
	url          string
	subscribed   map[string]bool
	mu           sync.RWMutex
	done         chan struct{}
	interrupt    chan os.Signal
	reconnect    bool
	maxRetries   int
	retryDelay   time.Duration
	onMessage    func([]byte)
	onError      func(error)
	onConnect    func()
	onDisconnect func()
}

// ClientConfig 客户端配置
type ClientConfig struct {
	URL        string        `json:"url" yaml:"url"`
	Reconnect  bool          `json:"reconnect" yaml:"reconnect"`
	MaxRetries int           `json:"max_retries" yaml:"max_retries"`
	RetryDelay time.Duration `json:"retry_delay" yaml:"retry_delay"`
	PingPeriod time.Duration `json:"ping_period" yaml:"ping_period"`
	PongWait   time.Duration `json:"pong_wait" yaml:"pong_wait"`
	WriteWait  time.Duration `json:"write_wait" yaml:"write_wait"`
	ReadLimit  int64         `json:"read_limit" yaml:"read_limit"`
}

// DefaultClientConfig 默认客户端配置
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		URL:        "ws://localhost:8080/ws",
		Reconnect:  true,
		MaxRetries: 5,
		RetryDelay: 5 * time.Second,
		PingPeriod: 54 * time.Second,
		PongWait:   60 * time.Second,
		WriteWait:  10 * time.Second,
		ReadLimit:  512,
	}
}

// NewWebSocketClient 创建WebSocket客户端
func NewWebSocketClient(config *ClientConfig) *WebSocketClient {
	if config == nil {
		config = DefaultClientConfig()
	}

	return &WebSocketClient{
		url:        config.URL,
		subscribed: make(map[string]bool),
		done:       make(chan struct{}),
		interrupt:  make(chan os.Signal, 1),
		reconnect:  config.Reconnect,
		maxRetries: config.MaxRetries,
		retryDelay: config.RetryDelay,
	}
}

// SetMessageHandler 设置消息处理器
func (c *WebSocketClient) SetMessageHandler(handler func([]byte)) {
	c.onMessage = handler
}

// SetErrorHandler 设置错误处理器
func (c *WebSocketClient) SetErrorHandler(handler func(error)) {
	c.onError = handler
}

// SetConnectHandler 设置连接处理器
func (c *WebSocketClient) SetConnectHandler(handler func()) {
	c.onConnect = handler
}

// SetDisconnectHandler 设置断开连接处理器
func (c *WebSocketClient) SetDisconnectHandler(handler func()) {
	c.onDisconnect = handler
}

// Connect 连接到WebSocket服务器
func (c *WebSocketClient) Connect() error {
	u, err := url.Parse(c.url)
	if err != nil {
		return fmt.Errorf("解析URL失败: %v", err)
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return fmt.Errorf("连接失败: %v", err)
	}

	c.conn = conn

	// 设置连接参数
	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	// 触发连接事件
	if c.onConnect != nil {
		c.onConnect()
	}

	return nil
}

// Disconnect 断开连接
func (c *WebSocketClient) Disconnect() error {
	close(c.done)

	if c.conn != nil {
		err := c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			return fmt.Errorf("发送关闭消息失败: %v", err)
		}

		err = c.conn.Close()
		if err != nil {
			return fmt.Errorf("关闭连接失败: %v", err)
		}
	}

	// 触发断开连接事件
	if c.onDisconnect != nil {
		c.onDisconnect()
	}

	return nil
}

// Subscribe 订阅交易对
func (c *WebSocketClient) Subscribe(symbols []string) error {
	if c.conn == nil {
		return fmt.Errorf("连接未建立")
	}

	message := Message{
		Type:      "subscribe",
		Symbols:   symbols,
		Timestamp: time.Now().UnixMilli(),
	}

	err := c.conn.WriteJSON(message)
	if err != nil {
		return fmt.Errorf("发送订阅消息失败: %v", err)
	}

	// 更新订阅状态
	c.mu.Lock()
	for _, symbol := range symbols {
		c.subscribed[symbol] = true
	}
	c.mu.Unlock()

	return nil
}

// Unsubscribe 取消订阅交易对
func (c *WebSocketClient) Unsubscribe(symbols []string) error {
	if c.conn == nil {
		return fmt.Errorf("连接未建立")
	}

	message := Message{
		Type:      "unsubscribe",
		Symbols:   symbols,
		Timestamp: time.Now().UnixMilli(),
	}

	err := c.conn.WriteJSON(message)
	if err != nil {
		return fmt.Errorf("发送取消订阅消息失败: %v", err)
	}

	// 更新订阅状态
	c.mu.Lock()
	for _, symbol := range symbols {
		delete(c.subscribed, symbol)
	}
	c.mu.Unlock()

	return nil
}

// GetSubscriptions 获取当前订阅
func (c *WebSocketClient) GetSubscriptions() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	symbols := make([]string, 0, len(c.subscribed))
	for symbol := range c.subscribed {
		symbols = append(symbols, symbol)
	}

	return symbols
}

// SendPing 发送心跳
func (c *WebSocketClient) SendPing() error {
	if c.conn == nil {
		return fmt.Errorf("连接未建立")
	}

	message := Message{
		Type:      "ping",
		Timestamp: time.Now().UnixMilli(),
	}

	err := c.conn.WriteJSON(message)
	if err != nil {
		return fmt.Errorf("发送心跳失败: %v", err)
	}

	return nil
}

// Start 启动客户端
func (c *WebSocketClient) Start() error {
	// 连接信号处理
	signal.Notify(c.interrupt, os.Interrupt)

	// 启动消息读取循环
	go c.readPump()

	// 启动心跳循环
	go c.pingPump()

	// 等待中断信号
	select {
	case <-c.interrupt:
		log.Println("收到中断信号，正在关闭...")
		return c.Disconnect()
	case <-c.done:
		return nil
	}
}

// readPump 读取消息循环
func (c *WebSocketClient) readPump() {
	defer func() {
		if c.reconnect {
			c.reconnectWithRetry()
		}
	}()

	for {
		select {
		case <-c.done:
			return
		default:
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket错误: %v", err)
					if c.onError != nil {
						c.onError(err)
					}
				}
				return
			}

			// 处理消息
			if c.onMessage != nil {
				c.onMessage(message)
			}
		}
	}
}

// pingPump 心跳循环
func (c *WebSocketClient) pingPump() {
	ticker := time.NewTicker(54 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-c.done:
			return
		case <-ticker.C:
			if err := c.SendPing(); err != nil {
				log.Printf("发送心跳失败: %v", err)
				return
			}
		}
	}
}

// reconnectWithRetry 重连重试
func (c *WebSocketClient) reconnectWithRetry() {
	for i := 0; i < c.maxRetries; i++ {
		log.Printf("尝试重连 (%d/%d)...", i+1, c.maxRetries)

		time.Sleep(c.retryDelay)

		err := c.Connect()
		if err != nil {
			log.Printf("重连失败: %v", err)
			continue
		}

		// 重连成功，恢复订阅
		c.restoreSubscriptions()

		// 重新启动消息循环
		go c.readPump()
		go c.pingPump()

		log.Println("重连成功")
		return
	}

	log.Printf("重连失败，已达到最大重试次数: %d", c.maxRetries)
}

// restoreSubscriptions 恢复订阅
func (c *WebSocketClient) restoreSubscriptions() {
	c.mu.RLock()
	symbols := make([]string, 0, len(c.subscribed))
	for symbol := range c.subscribed {
		symbols = append(symbols, symbol)
	}
	c.mu.RUnlock()

	if len(symbols) > 0 {
		err := c.Subscribe(symbols)
		if err != nil {
			log.Printf("恢复订阅失败: %v", err)
		}
	}
}

// IsConnected 检查是否已连接
func (c *WebSocketClient) IsConnected() bool {
	return c.conn != nil
}

// ExampleWebSocketClient 示例WebSocket客户端
func ExampleWebSocketClient() {
	// 创建客户端配置
	config := &ClientConfig{
		URL:        "ws://localhost:8080/ws",
		Reconnect:  true,
		MaxRetries: 5,
		RetryDelay: 5 * time.Second,
	}

	// 创建客户端
	client := NewWebSocketClient(config)

	// 设置消息处理器
	client.SetMessageHandler(func(data []byte) {
		var message Message
		if err := json.Unmarshal(data, &message); err != nil {
			log.Printf("解析消息失败: %v", err)
			return
		}

		log.Printf("收到消息: %+v", message)

		// 处理不同类型的消息
		switch message.Type {
		case "price_update":
			log.Printf("价格更新 - %s: %+v", message.Symbol, message.Data)
		case "subscription_confirmed":
			log.Printf("订阅确认: %v", message.Symbols)
		case "pong":
			log.Printf("收到心跳响应")
		default:
			log.Printf("未知消息类型: %s", message.Type)
		}
	})

	// 设置错误处理器
	client.SetErrorHandler(func(err error) {
		log.Printf("WebSocket错误: %v", err)
	})

	// 设置连接处理器
	client.SetConnectHandler(func() {
		log.Println("已连接到WebSocket服务器")
	})

	// 设置断开连接处理器
	client.SetDisconnectHandler(func() {
		log.Println("已断开WebSocket连接")
	})

	// 连接到服务器
	err := client.Connect()
	if err != nil {
		log.Fatalf("连接失败: %v", err)
	}

	// 订阅交易对
	symbols := []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"}
	err = client.Subscribe(symbols)
	if err != nil {
		log.Fatalf("订阅失败: %v", err)
	}

	log.Printf("已订阅交易对: %v", symbols)

	// 启动客户端
	err = client.Start()
	if err != nil {
		log.Fatalf("客户端启动失败: %v", err)
	}
}

// ExampleWebSocketClientWithReconnect 带重连功能的示例客户端
func ExampleWebSocketClientWithReconnect() {
	config := &ClientConfig{
		URL:        "ws://localhost:8080/ws",
		Reconnect:  true,
		MaxRetries: 10,
		RetryDelay: 3 * time.Second,
	}

	client := NewWebSocketClient(config)

	// 设置消息处理器
	client.SetMessageHandler(func(data []byte) {
		var message Message
		if err := json.Unmarshal(data, &message); err != nil {
			return
		}

		if message.Type == "price_update" {
			log.Printf("价格更新: %s = %+v", message.Symbol, message.Data)
		}
	})

	// 设置错误处理器
	client.SetErrorHandler(func(err error) {
		log.Printf("连接错误: %v", err)
	})

	// 设置连接处理器
	client.SetConnectHandler(func() {
		log.Println("连接成功，开始订阅...")

		// 订阅交易对
		symbols := []string{"BTCUSDT", "ETHUSDT"}
		if err := client.Subscribe(symbols); err != nil {
			log.Printf("订阅失败: %v", err)
		}
	})

	// 连接到服务器
	if err := client.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}

	// 启动客户端
	if err := client.Start(); err != nil {
		log.Fatalf("客户端启动失败: %v", err)
	}
}

// ExampleWebSocketClientBatch 批量订阅示例客户端
func ExampleWebSocketClientBatch() {
	client := NewWebSocketClient(nil)

	// 设置消息处理器
	client.SetMessageHandler(func(data []byte) {
		var message Message
		if err := json.Unmarshal(data, &message); err != nil {
			return
		}

		log.Printf("收到消息: %s", string(data))
	})

	// 连接到服务器
	if err := client.Connect(); err != nil {
		log.Fatalf("连接失败: %v", err)
	}

	// 批量订阅
	allSymbols := []string{
		"BTCUSDT", "ETHUSDT", "ADAUSDT", "DOTUSDT", "LINKUSDT",
		"UNIUSDT", "LTCUSDT", "BCHUSDT", "XRPUSDT", "EOSUSDT",
	}

	// 分批订阅
	batchSize := 3
	for i := 0; i < len(allSymbols); i += batchSize {
		end := i + batchSize
		if end > len(allSymbols) {
			end = len(allSymbols)
		}

		batch := allSymbols[i:end]
		if err := client.Subscribe(batch); err != nil {
			log.Printf("批量订阅失败: %v", err)
		}

		log.Printf("已订阅批次: %v", batch)
		time.Sleep(1 * time.Second) // 避免过于频繁的订阅
	}

	// 启动客户端
	if err := client.Start(); err != nil {
		log.Fatalf("客户端启动失败: %v", err)
	}
}
