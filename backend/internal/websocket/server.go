package websocket

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// WebSocketServerImpl WebSocket服务器实现
type WebSocketServerImpl struct {
	config        *ServerConfig
	logger        *zap.Logger
	server        *http.Server
	upgrader      websocket.Upgrader
	connections   map[string]*Connection
	subscriptions map[string][]string // symbol -> connection IDs
	mu            sync.RWMutex
	running       bool
	startTime     time.Time
	lastActivity  time.Time
}

// NewWebSocketServer 创建WebSocket服务器
func NewWebSocketServer(config *ServerConfig, logger *zap.Logger) WebSocketServer {
	if config == nil {
		config = DefaultServerConfig()
	}
	if logger == nil {
		logger = zap.NewNop()
	}

	return &WebSocketServerImpl{
		config:        config,
		logger:        logger,
		connections:   make(map[string]*Connection),
		subscriptions: make(map[string][]string),
		upgrader: websocket.Upgrader{
			ReadBufferSize:  config.ReadBufferSize,
			WriteBufferSize: config.WriteBufferSize,
			CheckOrigin: func(r *http.Request) bool {
				return true // 允许所有来源，生产环境需要限制
			},
		},
	}
}

// Start 启动WebSocket服务器
func (s *WebSocketServerImpl) Start(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return fmt.Errorf("服务器已在运行")
	}

	// 创建HTTP服务器
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWebSocket)

	s.server = &http.Server{
		Addr:         fmt.Sprintf("%s:%d", s.config.Host, s.config.Port),
		Handler:      mux,
		ReadTimeout:  s.config.HandshakeTimeout,
		WriteTimeout: s.config.WriteWait,
	}

	// 启动服务器
	go func() {
		s.logger.Info("启动WebSocket服务器", zap.String("addr", s.server.Addr))
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			s.logger.Error("WebSocket服务器启动失败", zap.Error(err))
		}
	}()

	s.running = true
	s.startTime = time.Now()
	s.lastActivity = time.Now()

	// 启动心跳检测
	go s.startHeartbeat(ctx)

	return nil
}

// Stop 停止WebSocket服务器
func (s *WebSocketServerImpl) Stop(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return fmt.Errorf("服务器未运行")
	}

	// 关闭所有连接
	for _, conn := range s.connections {
		conn.Conn.Close()
	}

	// 停止HTTP服务器
	if s.server != nil {
		s.server.Shutdown(ctx)
	}

	s.running = false
	s.logger.Info("WebSocket服务器已停止")

	return nil
}

// IsRunning 检查服务器是否运行
func (s *WebSocketServerImpl) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.running
}

// GetConnectionCount 获取连接数量
func (s *WebSocketServerImpl) GetConnectionCount() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.connections)
}

// GetConnections 获取所有连接
func (s *WebSocketServerImpl) GetConnections() []*Connection {
	s.mu.RLock()
	defer s.mu.RUnlock()

	connections := make([]*Connection, 0, len(s.connections))
	for _, conn := range s.connections {
		connections = append(connections, conn)
	}
	return connections
}

// handleWebSocket 处理WebSocket连接
func (s *WebSocketServerImpl) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// 检查连接数限制
	if len(s.connections) >= s.config.MaxConnections {
		http.Error(w, "连接数已达上限", http.StatusServiceUnavailable)
		return
	}

	// 升级到WebSocket连接
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Error("WebSocket升级失败", zap.Error(err))
		return
	}

	// 创建连接对象
	connectionID := generateConnectionID()
	connection := &Connection{
		ID:            connectionID,
		Conn:          conn,
		Subscriptions: make([]string, 0),
		LastPing:      time.Now(),
		CreatedAt:     time.Now(),
		IsActive:      true,
	}

	// 添加连接到管理器
	s.addConnection(connection)

	// 设置连接参数
	conn.SetReadLimit(int64(s.config.MaxMessageSize))
	conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
	conn.SetPongHandler(func(string) error {
		connection.LastPing = time.Now()
		conn.SetReadDeadline(time.Now().Add(s.config.PongWait))
		return nil
	})

	s.logger.Info("新WebSocket连接建立", zap.String("conn_id", connectionID))

	// 处理消息循环
	s.handleMessages(connection)

	// 清理连接
	s.removeConnection(connectionID)
	s.logger.Info("WebSocket连接关闭", zap.String("conn_id", connectionID))
}

// handleMessages 处理消息循环
func (s *WebSocketServerImpl) handleMessages(conn *Connection) {
	defer conn.Conn.Close()

	for {
		var msg Message
		err := conn.Conn.ReadJSON(&msg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				s.logger.Error("WebSocket读取错误", zap.String("conn_id", conn.ID), zap.Error(err))
			}
			break
		}

		s.lastActivity = time.Now()
		s.processMessage(conn, &msg)
	}
}

// processMessage 处理消息
func (s *WebSocketServerImpl) processMessage(conn *Connection, msg *Message) {
	switch msg.Type {
	case "subscribe":
		s.handleSubscribe(conn, msg.Symbols)
	case "unsubscribe":
		s.handleUnsubscribe(conn, msg.Symbols)
	case "ping":
		s.handlePing(conn)
	default:
		s.sendError(conn, "INVALID_MESSAGE_TYPE", "无效的消息类型")
	}
}

// handleSubscribe 处理订阅
func (s *WebSocketServerImpl) handleSubscribe(conn *Connection, symbols []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 添加订阅
	for _, symbol := range symbols {
		if !contains(conn.Subscriptions, symbol) {
			conn.Subscriptions = append(conn.Subscriptions, symbol)
		}
		if !contains(s.subscriptions[symbol], conn.ID) {
			s.subscriptions[symbol] = append(s.subscriptions[symbol], conn.ID)
		}
	}

	// 发送确认消息
	response := Message{
		Type:      "subscribe_success",
		Symbols:   symbols,
		Timestamp: time.Now().UnixMilli(),
	}
	s.sendMessage(conn, response)

	s.logger.Info("客户端订阅成功", zap.String("conn_id", conn.ID), zap.Strings("symbols", symbols))
}

// handleUnsubscribe 处理取消订阅
func (s *WebSocketServerImpl) handleUnsubscribe(conn *Connection, symbols []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 移除订阅
	for _, symbol := range symbols {
		conn.Subscriptions = removeFromSlice(conn.Subscriptions, symbol)
		s.subscriptions[symbol] = removeFromSlice(s.subscriptions[symbol], conn.ID)
	}

	// 发送确认消息
	response := Message{
		Type:      "unsubscribe_success",
		Symbols:   symbols,
		Timestamp: time.Now().UnixMilli(),
	}
	s.sendMessage(conn, response)

	s.logger.Info("客户端取消订阅", zap.String("conn_id", conn.ID), zap.Strings("symbols", symbols))
}

// handlePing 处理心跳
func (s *WebSocketServerImpl) handlePing(conn *Connection) {
	response := Message{
		Type:      "pong",
		Timestamp: time.Now().UnixMilli(),
	}
	s.sendMessage(conn, response)
}

// BroadcastToSymbol 向特定交易对的所有订阅者广播消息
func (s *WebSocketServerImpl) BroadcastToSymbol(symbol string, message interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	connIDs, exists := s.subscriptions[symbol]
	if !exists {
		return fmt.Errorf("交易对 %s 没有订阅者", symbol)
	}

	for _, connID := range connIDs {
		if conn, exists := s.connections[connID]; exists && conn.IsActive {
			s.sendMessage(conn, message)
		}
	}

	return nil
}

// BroadcastToAll 向所有连接广播消息
func (s *WebSocketServerImpl) BroadcastToAll(message interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, conn := range s.connections {
		if conn.IsActive {
			s.sendMessage(conn, message)
		}
	}

	return nil
}

// SendToConnection 向特定连接发送消息
func (s *WebSocketServerImpl) SendToConnection(connID string, message interface{}) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[connID]
	if !exists {
		return fmt.Errorf("连接 %s 不存在", connID)
	}

	if !conn.IsActive {
		return fmt.Errorf("连接 %s 不活跃", connID)
	}

	return s.sendMessage(conn, message)
}

// Subscribe 订阅交易对
func (s *WebSocketServerImpl) Subscribe(connID string, symbols []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connections[connID]
	if !exists {
		return fmt.Errorf("连接 %s 不存在", connID)
	}

	// 添加订阅
	for _, symbol := range symbols {
		if !contains(conn.Subscriptions, symbol) {
			conn.Subscriptions = append(conn.Subscriptions, symbol)
		}
		if !contains(s.subscriptions[symbol], connID) {
			s.subscriptions[symbol] = append(s.subscriptions[symbol], connID)
		}
	}

	return nil
}

// Unsubscribe 取消订阅交易对
func (s *WebSocketServerImpl) Unsubscribe(connID string, symbols []string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	conn, exists := s.connections[connID]
	if !exists {
		return fmt.Errorf("连接 %s 不存在", connID)
	}

	// 移除订阅
	for _, symbol := range symbols {
		conn.Subscriptions = removeFromSlice(conn.Subscriptions, symbol)
		s.subscriptions[symbol] = removeFromSlice(s.subscriptions[symbol], connID)
	}

	return nil
}

// GetSubscriptions 获取连接的订阅列表
func (s *WebSocketServerImpl) GetSubscriptions(connID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	conn, exists := s.connections[connID]
	if !exists {
		return nil
	}

	return conn.Subscriptions
}

// GetSubscribers 获取交易对的订阅者列表
func (s *WebSocketServerImpl) GetSubscribers(symbol string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	subscribers, exists := s.subscriptions[symbol]
	if !exists {
		return nil
	}

	// 返回副本，避免外部修改
	result := make([]string, len(subscribers))
	copy(result, subscribers)
	return result
}

// 辅助方法

func (s *WebSocketServerImpl) addConnection(conn *Connection) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.connections[conn.ID] = conn
}

func (s *WebSocketServerImpl) removeConnection(connID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn, exists := s.connections[connID]; exists {
		// 清理订阅
		for _, symbol := range conn.Subscriptions {
			s.subscriptions[symbol] = removeFromSlice(s.subscriptions[symbol], connID)
		}
		delete(s.connections, connID)
	}
}

func (s *WebSocketServerImpl) sendMessage(conn *Connection, message interface{}) error {
	conn.Conn.SetWriteDeadline(time.Now().Add(s.config.WriteWait))
	return conn.Conn.WriteJSON(message)
}

func (s *WebSocketServerImpl) sendError(conn *Connection, code, message string) {
	errorMsg := ErrorMessage{
		Code:    code,
		Message: message,
	}
	s.sendMessage(conn, errorMsg)
}

func (s *WebSocketServerImpl) startHeartbeat(ctx context.Context) {
	ticker := time.NewTicker(s.config.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.sendPingToAll()
		}
	}
}

func (s *WebSocketServerImpl) sendPingToAll() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pingMsg := Message{
		Type:      "ping",
		Timestamp: time.Now().UnixMilli(),
	}

	for _, conn := range s.connections {
		if conn.IsActive {
			s.sendMessage(conn, pingMsg)
		}
	}
}

// 工具函数

func generateConnectionID() string {
	return fmt.Sprintf("conn_%d", time.Now().UnixNano())
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func removeFromSlice(slice []string, item string) []string {
	for i, s := range slice {
		if s == item {
			return append(slice[:i], slice[i+1:]...)
		}
	}
	return slice
}
