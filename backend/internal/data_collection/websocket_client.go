package data_collection

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketClientImpl WebSocket客户端实现
type WebSocketClientImpl struct {
	// 配置
	config *WebSocketConfig

	// 日志
	logger *zap.Logger

	// 连接管理
	conn        *websocket.Conn
	connected   bool
	connectedAt time.Time

	// 并发控制
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	// 重连管理
	autoReconnect        bool
	reconnectCount       int64
	maxReconnectAttempts int

	// 心跳管理
	heartbeatTicker *time.Ticker
	heartbeatCtx    context.Context
	heartbeatCancel context.CancelFunc

	// 订阅管理
	subscriptions map[string]bool
	subsMu        sync.RWMutex

	// 消息处理
	messageHandler MessageHandler
	errorHandler   ErrorHandler

	// 统计信息
	messageCount  int64
	errorCount    int64
	lastError     error
	lastHeartbeat time.Time
	latency       time.Duration
}

// NewWebSocketClient 创建新的WebSocket客户端
func NewWebSocketClient(config *WebSocketConfig, logger *zap.Logger) *WebSocketClientImpl {
	if config == nil {
		config = DefaultWebSocketConfig()
	}

	if logger == nil {
		logger, _ = zap.NewDevelopment()
	}

	return &WebSocketClientImpl{
		config:               config,
		logger:               logger,
		subscriptions:        make(map[string]bool),
		maxReconnectAttempts: config.MaxReconnectAttempts,
	}
}

// Connect 建立连接
func (c *WebSocketClientImpl) Connect(ctx context.Context) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// 检查是否已连接
	if c.connected {
		return ErrServiceAlreadyRunning
	}

	// 如果有旧的连接或上下文，先清理
	if c.cancel != nil {
		c.cancel()
	}
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	// 等待旧的goroutine退出
	if c.cancel != nil {
		c.mu.Unlock()
		c.wg.Wait()
		c.mu.Lock()
	}

	// 创建上下文
	c.ctx, c.cancel = context.WithCancel(ctx)

	// 建立WebSocket连接
	conn, _, err := websocket.DefaultDialer.DialContext(ctx, c.config.URL, nil)
	if err != nil {
		c.lastError = err
		atomic.AddInt64(&c.errorCount, 1)
		return fmt.Errorf("WebSocket连接失败: %w", err)
	}

	c.conn = conn
	c.connected = true
	c.connectedAt = time.Now()

	// 启动消息处理
	c.wg.Add(1)
	go c.messageLoop()

	// 启动心跳（如果启用）
	if c.config.EnablePingPong {
		c.startHeartbeat()
	}

	c.logger.Info("WebSocket连接成功",
		zap.String("url", c.config.URL),
		zap.Time("connected_at", c.connectedAt),
	)

	return nil
}

// Disconnect 断开连接
func (c *WebSocketClientImpl) Disconnect(ctx context.Context) error {
	// 检查连接状态
	c.mu.RLock()
	if !c.connected {
		c.mu.RUnlock()
		return ErrServiceNotRunning
	}
	connectedAt := c.connectedAt
	c.mu.RUnlock()

	// 停止心跳（不持有锁）
	c.StopHeartbeat()

	// 取消上下文并关闭连接
	c.mu.Lock()
	if c.cancel != nil {
		c.cancel()
	}

	// 关闭连接
	if c.conn != nil {
		// 发送关闭帧
		c.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		c.conn.Close()
		c.conn = nil
	}

	// 设置连接状态为false，让消息循环退出
	c.connected = false
	c.mu.Unlock()

	// 等待消息处理完成（带超时）
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 正常完成
	case <-time.After(2 * time.Second):
		// 超时，强制继续
		c.logger.Warn("等待消息处理完成超时")
	}

	c.logger.Info("WebSocket连接已断开",
		zap.Duration("uptime", time.Since(connectedAt)),
	)

	return nil
}

// IsConnected 检查连接状态
func (c *WebSocketClientImpl) IsConnected() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.connected
}

// Reconnect 重新连接
func (c *WebSocketClientImpl) Reconnect(ctx context.Context) error {
	// 先断开现有连接
	if c.connected {
		c.Disconnect(ctx)
	}

	// 重新连接
	return c.Connect(ctx)
}

// Subscribe 订阅单个交易对
func (c *WebSocketClientImpl) Subscribe(symbol string) error {
	if !c.connected {
		return ErrServiceNotRunning
	}

	c.subsMu.Lock()
	defer c.subsMu.Unlock()

	// 发送订阅消息
	request := SubscriptionRequest{
		Symbol:    symbol,
		Channel:   "ticker",
		Action:    ActionSubscribe,
		Timestamp: time.Now().Unix(),
	}

	if err := c.sendSubscriptionRequest(request); err != nil {
		return err
	}

	// 记录订阅
	c.subscriptions[symbol] = true

	c.logger.Info("订阅交易对成功",
		zap.String("symbol", symbol),
	)

	return nil
}

// Unsubscribe 取消订阅
func (c *WebSocketClientImpl) Unsubscribe(symbol string) error {
	if !c.connected {
		return ErrServiceNotRunning
	}

	c.subsMu.Lock()
	defer c.subsMu.Unlock()

	// 发送取消订阅消息
	request := SubscriptionRequest{
		Symbol:    symbol,
		Channel:   "ticker",
		Action:    ActionUnsubscribe,
		Timestamp: time.Now().Unix(),
	}

	if err := c.sendSubscriptionRequest(request); err != nil {
		return err
	}

	// 删除订阅记录
	delete(c.subscriptions, symbol)

	c.logger.Info("取消订阅交易对成功",
		zap.String("symbol", symbol),
	)

	return nil
}

// BatchSubscribe 批量订阅
func (c *WebSocketClientImpl) BatchSubscribe(symbols []string) error {
	if !c.connected {
		return ErrServiceNotRunning
	}

	c.subsMu.Lock()
	defer c.subsMu.Unlock()

	// 批量发送订阅请求
	for _, symbol := range symbols {
		request := SubscriptionRequest{
			Symbol:    symbol,
			Channel:   "ticker",
			Action:    ActionSubscribe,
			Timestamp: time.Now().Unix(),
		}

		if err := c.sendSubscriptionRequest(request); err != nil {
			return err
		}

		c.subscriptions[symbol] = true
	}

	c.logger.Info("批量订阅成功",
		zap.Int("count", len(symbols)),
		zap.Strings("symbols", symbols),
	)

	return nil
}

// GetSubscriptions 获取订阅列表
func (c *WebSocketClientImpl) GetSubscriptions() []string {
	c.subsMu.RLock()
	defer c.subsMu.RUnlock()

	subscriptions := make([]string, 0, len(c.subscriptions))
	for symbol := range c.subscriptions {
		subscriptions = append(subscriptions, symbol)
	}

	return subscriptions
}

// SetMessageHandler 设置消息处理器
func (c *WebSocketClientImpl) SetMessageHandler(handler MessageHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messageHandler = handler
}

// SetErrorHandler 设置错误处理器
func (c *WebSocketClientImpl) SetErrorHandler(handler ErrorHandler) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.errorHandler = handler
}

// SendMessage 发送消息
func (c *WebSocketClientImpl) SendMessage(data []byte) error {
	if !c.connected {
		return ErrServiceNotRunning
	}

	c.mu.RLock()
	conn := c.conn
	c.mu.RUnlock()

	if conn == nil {
		return ErrConnectionFailed
	}

	return conn.WriteMessage(websocket.TextMessage, data)
}

// StartHeartbeat 启动心跳
func (c *WebSocketClientImpl) StartHeartbeat(interval time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.heartbeatTicker != nil {
		return // 心跳已在运行
	}

	c.heartbeatCtx, c.heartbeatCancel = context.WithCancel(c.ctx)
	c.heartbeatTicker = time.NewTicker(interval)

	c.wg.Add(1)
	go c.heartbeatLoop()
}

// StopHeartbeat 停止心跳
func (c *WebSocketClientImpl) StopHeartbeat() {
	c.mu.Lock()

	if c.heartbeatTicker != nil {
		c.heartbeatTicker.Stop()
		c.heartbeatTicker = nil
	}

	if c.heartbeatCancel != nil {
		c.heartbeatCancel()
		c.heartbeatCancel = nil
	}

	c.mu.Unlock()

	// 等待心跳循环完成
	done := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// 正常完成
	case <-time.After(1 * time.Second):
		// 超时，强制继续
		c.logger.Warn("等待心跳循环完成超时")
	}
}

// SetAutoReconnect 设置自动重连
func (c *WebSocketClientImpl) SetAutoReconnect(enabled bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.autoReconnect = enabled
}

// GetReconnectCount 获取重连次数
func (c *WebSocketClientImpl) GetReconnectCount() int64 {
	return atomic.LoadInt64(&c.reconnectCount)
}

// GetConnectionInfo 获取连接信息
func (c *WebSocketClientImpl) GetConnectionInfo() *ConnectionInfo {
	c.mu.RLock()
	defer c.mu.RUnlock()

	info := &ConnectionInfo{
		Connected:      c.connected,
		URL:            c.config.URL,
		ConnectedAt:    c.connectedAt,
		LastHeartbeat:  c.lastHeartbeat,
		ReconnectCount: atomic.LoadInt64(&c.reconnectCount),
		Latency:        c.latency,
		MessageCount:   atomic.LoadInt64(&c.messageCount),
		ErrorCount:     atomic.LoadInt64(&c.errorCount),
	}

	if c.lastError != nil {
		info.LastError = c.lastError.Error()
	}

	return info
}

// GetConfig 获取配置
func (c *WebSocketClientImpl) GetConfig() *WebSocketConfig {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.config
}

// SetConfig 设置配置
func (c *WebSocketClientImpl) SetConfig(config *WebSocketConfig) error {
	if config == nil {
		return ErrInvalidConfig
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.config = config
	c.maxReconnectAttempts = config.MaxReconnectAttempts

	return nil
}

// SetConnectionTimeout 设置连接超时（测试辅助方法）
func (c *WebSocketClientImpl) SetConnectionTimeout(timeout time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.config == nil {
		c.config = DefaultWebSocketConfig()
	}
	c.config.ConnectionTimeout = timeout
}

// 私有方法

// messageLoop 消息处理循环
func (c *WebSocketClientImpl) messageLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.ctx.Done():
			c.logger.Debug("消息循环收到停止信号")
			return
		default:
		}

		// 获取连接（加锁）
		c.mu.RLock()
		conn := c.conn
		connected := c.connected
		c.mu.RUnlock()

		// 检查连接是否仍然有效
		if conn == nil || !connected {
			c.logger.Debug("连接已关闭，退出消息循环")
			return
		}

		// 使用defer恢复panic，并使用标志来决定是否退出循环
		shouldExit := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					c.logger.Debug("消息循环发生panic，退出", zap.Any("panic", r))
					// 不再修改连接状态，避免影响新连接的状态
					// 设置退出标志
					shouldExit = true
					return
				}
			}()

			// 设置读取超时
			conn.SetReadDeadline(time.Now().Add(1 * time.Second))

			// 读取消息
			_, message, err := conn.ReadMessage()
			if err != nil {
				// 检查是否是超时错误
				if netErr, ok := err.(net.Error); ok && netErr.Timeout() {
					// 超时，继续循环
					return
				}

				// 检查连接状态，如果已经被标记为未连接，说明是主动断开，不触发错误处理
				c.mu.RLock()
				isConnected := c.connected
				c.mu.RUnlock()

				// 检查是否是连接关闭错误
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					c.logger.Debug("连接正常关闭")
					shouldExit = true
					return
				}

				// 检查是否是 "use of closed network connection" 错误
				if err.Error() == "use of closed network connection" || err.Error() == "repeated read on failed websocket connection" {
					c.logger.Debug("连接已关闭")
					shouldExit = true
					return
				}

				// 其他错误，只有在连接仍然标记为连接状态时才处理（避免重复触发）
				if isConnected {
					c.handleConnectionError(err)
				}
				shouldExit = true
				return
			}

			// 处理消息
			c.handleMessage(message)
		}()

		// 如果需要退出，则退出循环
		if shouldExit {
			return
		}
	}
}

// heartbeatLoop 心跳循环
func (c *WebSocketClientImpl) heartbeatLoop() {
	defer c.wg.Done()

	for {
		select {
		case <-c.heartbeatCtx.Done():
			return
		case <-c.heartbeatTicker.C:
			// 检查连接是否仍然有效
			c.mu.RLock()
			conn := c.conn
			connected := c.connected
			c.mu.RUnlock()

			if conn == nil || !connected {
				c.logger.Debug("连接已关闭，停止心跳")
				return
			}

			c.sendHeartbeat()
		}
	}
}

// handleMessage 处理消息
func (c *WebSocketClientImpl) handleMessage(data []byte) {
	atomic.AddInt64(&c.messageCount, 1)

	// 调用消息处理器
	if c.messageHandler != nil {
		if err := c.messageHandler(data); err != nil {
			c.logger.Error("消息处理失败",
				zap.Error(err),
				zap.String("data", string(data)),
			)
		}
	}
}

// handleConnectionError 处理连接错误
func (c *WebSocketClientImpl) handleConnectionError(err error) {
	c.lastError = err
	atomic.AddInt64(&c.errorCount, 1)

	// 调用错误处理器
	if c.errorHandler != nil {
		c.errorHandler(err)
	}

	// 自动重连
	if c.autoReconnect && atomic.LoadInt64(&c.reconnectCount) < int64(c.maxReconnectAttempts) {
		c.attemptReconnect()
	}
}

// attemptReconnect 尝试重连
func (c *WebSocketClientImpl) attemptReconnect() {
	attempt := atomic.AddInt64(&c.reconnectCount, 1)

	c.logger.Info("尝试自动重连",
		zap.Int64("attempt", attempt),
		zap.Int("max_attempts", c.maxReconnectAttempts),
	)

	// 计算指数退避延迟
	delay := c.calculateReconnectDelay(attempt)

	c.logger.Debug("重连延迟",
		zap.Duration("delay", delay),
		zap.Int64("attempt", attempt),
	)

	// 等待重连间隔
	time.Sleep(delay)

	// 重新连接（使用新的上下文）
	if err := c.Connect(context.Background()); err != nil {
		c.logger.Error("自动重连失败",
			zap.Error(err),
			zap.Int64("attempt", attempt),
		)
	} else {
		c.logger.Info("自动重连成功",
			zap.Int64("attempt", attempt),
		)
		// 重新订阅
		c.resubscribe()
	}
}

// calculateReconnectDelay 计算重连延迟（指数退避策略）
func (c *WebSocketClientImpl) calculateReconnectDelay(attempt int64) time.Duration {
	// 基础延迟
	baseDelay := c.config.ReconnectInterval

	// 指数退避：delay = baseDelay * 2^(attempt-1)
	// 但限制最大延迟为5分钟
	maxDelay := 5 * time.Minute

	delay := baseDelay
	for i := int64(1); i < attempt; i++ {
		delay *= 2
		if delay > maxDelay {
			delay = maxDelay
			break
		}
	}

	return delay
}

// resubscribe 重新订阅
func (c *WebSocketClientImpl) resubscribe() {
	c.subsMu.RLock()
	symbols := make([]string, 0, len(c.subscriptions))
	for symbol := range c.subscriptions {
		symbols = append(symbols, symbol)
	}
	c.subsMu.RUnlock()

	if len(symbols) > 0 {
		c.BatchSubscribe(symbols)
	}
}

// sendHeartbeat 发送心跳
func (c *WebSocketClientImpl) sendHeartbeat() {
	heartbeat := HeartbeatMessage{
		Type:      MessageTypeHeartbeat,
		Timestamp: time.Now(),
		Ping:      true,
	}

	data, err := json.Marshal(heartbeat)
	if err != nil {
		c.logger.Error("心跳消息序列化失败", zap.Error(err))
		return
	}

	if err := c.SendMessage(data); err != nil {
		c.logger.Error("发送心跳失败", zap.Error(err))
		return
	}

	c.lastHeartbeat = time.Now()
}

// sendSubscriptionRequest 发送订阅请求
func (c *WebSocketClientImpl) sendSubscriptionRequest(request SubscriptionRequest) error {
	data, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("订阅请求序列化失败: %w", err)
	}

	return c.SendMessage(data)
}

// startHeartbeat 启动心跳
func (c *WebSocketClientImpl) startHeartbeat() {
	if c.config.HeartbeatInterval > 0 {
		c.StartHeartbeat(c.config.HeartbeatInterval)
	}
}

// stopHeartbeat 停止心跳
func (c *WebSocketClientImpl) stopHeartbeat() {
	c.StopHeartbeat()
}

// 测试辅助方法

// simulateDisconnect 模拟连接断开（仅用于测试）
func (c *WebSocketClientImpl) simulateDisconnect() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}
	c.connected = false
}

// simulateMessage 模拟接收消息（仅用于测试）
func (c *WebSocketClientImpl) simulateMessage(data []byte) {
	c.handleMessage(data)
}

// simulateError 模拟错误（仅用于测试）
func (c *WebSocketClientImpl) simulateError(err error) {
	c.handleConnectionError(err)
}
