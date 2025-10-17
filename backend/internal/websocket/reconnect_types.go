package websocket

import (
	"context"
	"sync"
	"time"
)

// ReconnectManager 重连管理器接口
type ReconnectManager interface {
	// 重连管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 连接管理
	AddConnection(connID string, config *ConnectionReconnectConfig) error
	RemoveConnection(connID string) error
	UpdateConnectionStatus(connID string, isConnected bool) error

	// 重连控制
	TriggerReconnect(connID string) error
	CancelReconnect(connID string) error

	// 状态恢复
	SaveConnectionState(connID string, state *ConnectionState) error
	RestoreConnectionState(connID string) (*ConnectionState, error)

	// 配置管理
	SetReconnectStrategy(strategy ReconnectStrategy)
	SetMaxReconnectAttempts(max int)
	SetReconnectInterval(interval time.Duration)

	// 统计和监控
	GetReconnectStats() *ReconnectStats
	GetConnectionReconnectStatus(connID string) *ConnectionReconnectStatus
	GetAllReconnectStatus() map[string]*ConnectionReconnectStatus
}

// ConnectionReconnectConfig 连接重连配置
type ConnectionReconnectConfig struct {
	MaxAttempts       int           `json:"max_attempts" yaml:"max_attempts"`
	BaseInterval      time.Duration `json:"base_interval" yaml:"base_interval"`
	MaxInterval       time.Duration `json:"max_interval" yaml:"max_interval"`
	BackoffMultiplier float64       `json:"backoff_multiplier" yaml:"backoff_multiplier"`
	JitterEnabled     bool          `json:"jitter_enabled" yaml:"jitter_enabled"`
	AutoReconnect     bool          `json:"auto_reconnect" yaml:"auto_reconnect"`
	StateRecovery     bool          `json:"state_recovery" yaml:"state_recovery"`
	ReconnectTimeout  time.Duration `json:"reconnect_timeout" yaml:"reconnect_timeout"`
}

// DefaultConnectionReconnectConfig 默认连接重连配置
func DefaultConnectionReconnectConfig() *ConnectionReconnectConfig {
	return &ConnectionReconnectConfig{
		MaxAttempts:       5,
		BaseInterval:      1 * time.Second,
		MaxInterval:       30 * time.Second,
		BackoffMultiplier: 2.0,
		JitterEnabled:     true,
		AutoReconnect:     true,
		StateRecovery:     true,
		ReconnectTimeout:  10 * time.Second,
	}
}

// ConnectionState 连接状态
type ConnectionState struct {
	ConnID        string                 `json:"conn_id"`
	Subscriptions []string               `json:"subscriptions"`
	LastMessageID string                 `json:"last_message_id"`
	LastActivity  time.Time              `json:"last_activity"`
	CustomData    map[string]interface{} `json:"custom_data"`
	CreatedAt     time.Time              `json:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at"`
}

// ConnectionReconnectStatus 连接重连状态
type ConnectionReconnectStatus struct {
	ConnID               string        `json:"conn_id"`
	IsConnected          bool          `json:"is_connected"`
	IsReconnecting       bool          `json:"is_reconnecting"`
	ReconnectAttempts    int           `json:"reconnect_attempts"`
	LastReconnectTime    time.Time     `json:"last_reconnect_time"`
	NextReconnectTime    time.Time     `json:"next_reconnect_time"`
	ReconnectInterval    time.Duration `json:"reconnect_interval"`
	TotalReconnects      int64         `json:"total_reconnects"`
	SuccessfulReconnects int64         `json:"successful_reconnects"`
	FailedReconnects     int64         `json:"failed_reconnects"`
	LastDisconnectTime   time.Time     `json:"last_disconnect_time"`
	LastConnectTime      time.Time     `json:"last_connect_time"`
	CreatedAt            time.Time     `json:"created_at"`
}

// ReconnectStats 重连统计信息
type ReconnectStats struct {
	TotalReconnects      int64            `json:"total_reconnects"`
	SuccessfulReconnects int64            `json:"successful_reconnects"`
	FailedReconnects     int64            `json:"failed_reconnects"`
	ActiveReconnects     int              `json:"active_reconnects"`
	AverageReconnectTime time.Duration    `json:"average_reconnect_time"`
	LastReconnectTime    time.Time        `json:"last_reconnect_time"`
	StartTime            time.Time        `json:"start_time"`
	ConnectionStats      map[string]int64 `json:"connection_stats"`
}

// ReconnectStrategy 重连策略
type ReconnectStrategy int

const (
	ReconnectStrategyLinear ReconnectStrategy = iota
	ReconnectStrategyExponential
	ReconnectStrategyFixed
)

// ReconnectManagerImpl 重连管理器实现
type ReconnectManagerImpl struct {
	// 配置
	config *ReconnectConfig

	// 连接状态
	connections map[string]*ConnectionReconnectStatus
	connMu      sync.RWMutex

	// 连接状态存储
	connectionStates map[string]*ConnectionState
	stateMu          sync.RWMutex

	// 控制
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.RWMutex

	// 统计
	stats   *ReconnectStats
	statsMu sync.RWMutex

	// 重连队列
	reconnectQueue chan string

	// 日志记录器
	logger interface{}
}

// ReconnectConfig 重连配置
type ReconnectConfig struct {
	MaxReconnectAttempts int               `json:"max_reconnect_attempts" yaml:"max_reconnect_attempts"`
	ReconnectInterval    time.Duration     `json:"reconnect_interval" yaml:"reconnect_interval"`
	ReconnectTimeout     time.Duration     `json:"reconnect_timeout" yaml:"reconnect_timeout"`
	StateRecoveryEnabled bool              `json:"state_recovery_enabled" yaml:"state_recovery_enabled"`
	CleanupInterval      time.Duration     `json:"cleanup_interval" yaml:"cleanup_interval"`
	Strategy             ReconnectStrategy `json:"strategy" yaml:"strategy"`
}

// DefaultReconnectConfig 默认重连配置
func DefaultReconnectConfig() *ReconnectConfig {
	return &ReconnectConfig{
		MaxReconnectAttempts: 5,
		ReconnectInterval:    5 * time.Second,
		ReconnectTimeout:     10 * time.Second,
		StateRecoveryEnabled: true,
		CleanupInterval:      60 * time.Second,
		Strategy:             ReconnectStrategyExponential,
	}
}

// NewReconnectManager 创建重连管理器
func NewReconnectManager(config *ReconnectConfig, logger interface{}) ReconnectManager {
	if config == nil {
		config = DefaultReconnectConfig()
	}

	return &ReconnectManagerImpl{
		config:           config,
		connections:      make(map[string]*ConnectionReconnectStatus),
		connectionStates: make(map[string]*ConnectionState),
		reconnectQueue:   make(chan string, 100),
		stats: &ReconnectStats{
			ConnectionStats: make(map[string]int64),
			StartTime:       time.Now(),
		},
		logger: logger,
	}
}

// Start 启动重连管理器
func (rm *ReconnectManagerImpl) Start(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if rm.running {
		return &ReconnectError{
			Type:    "ALREADY_RUNNING",
			Message: "重连管理器已在运行",
		}
	}

	rm.ctx, rm.cancel = context.WithCancel(ctx)
	rm.running = true

	// 启动重连处理协程
	go rm.reconnectLoop()

	// 启动清理协程
	go rm.cleanupLoop()

	return nil
}

// Stop 停止重连管理器
func (rm *ReconnectManagerImpl) Stop(ctx context.Context) error {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if !rm.running {
		return &ReconnectError{
			Type:    "NOT_RUNNING",
			Message: "重连管理器未运行",
		}
	}

	rm.running = false
	rm.cancel()

	// 关闭重连队列
	close(rm.reconnectQueue)

	return nil
}

// IsRunning 检查是否运行
func (rm *ReconnectManagerImpl) IsRunning() bool {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.running
}

// AddConnection 添加连接
func (rm *ReconnectManagerImpl) AddConnection(connID string, config *ConnectionReconnectConfig) error {
	rm.connMu.Lock()
	defer rm.connMu.Unlock()

	if config == nil {
		config = DefaultConnectionReconnectConfig()
	}

	rm.connections[connID] = &ConnectionReconnectStatus{
		ConnID:            connID,
		IsConnected:       true,
		IsReconnecting:    false,
		ReconnectAttempts: 0,
		ReconnectInterval: config.BaseInterval,
		CreatedAt:         time.Now(),
		LastConnectTime:   time.Now(),
	}

	rm.updateStats("connection_added", connID)
	return nil
}

// RemoveConnection 移除连接
func (rm *ReconnectManagerImpl) RemoveConnection(connID string) error {
	rm.connMu.Lock()
	defer rm.connMu.Unlock()

	delete(rm.connections, connID)
	rm.updateStats("connection_removed", connID)
	return nil
}

// UpdateConnectionStatus 更新连接状态
func (rm *ReconnectManagerImpl) UpdateConnectionStatus(connID string, isConnected bool) error {
	rm.connMu.Lock()
	defer rm.connMu.Unlock()

	conn, exists := rm.connections[connID]
	if !exists {
		return &ReconnectError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	conn.IsConnected = isConnected

	if isConnected {
		conn.IsReconnecting = false
		conn.ReconnectAttempts = 0
		conn.LastConnectTime = time.Now()
		rm.updateStats("connection_connected", connID)
	} else {
		conn.LastDisconnectTime = time.Now()
		rm.updateStats("connection_disconnected", connID)

		// 如果启用自动重连，触发重连
		if conn.TotalReconnects < int64(rm.config.MaxReconnectAttempts) {
			rm.triggerReconnect(connID)
		}
	}

	return nil
}

// TriggerReconnect 触发重连
func (rm *ReconnectManagerImpl) TriggerReconnect(connID string) error {
	rm.connMu.RLock()
	conn, exists := rm.connections[connID]
	rm.connMu.RUnlock()

	if !exists {
		return &ReconnectError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	if conn.IsReconnecting {
		return &ReconnectError{
			Type:    "ALREADY_RECONNECTING",
			Message: "连接正在重连中",
			ConnID:  connID,
		}
	}

	rm.triggerReconnect(connID)
	return nil
}

// CancelReconnect 取消重连
func (rm *ReconnectManagerImpl) CancelReconnect(connID string) error {
	rm.connMu.Lock()
	defer rm.connMu.Unlock()

	conn, exists := rm.connections[connID]
	if !exists {
		return &ReconnectError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	conn.IsReconnecting = false
	rm.updateStats("reconnect_cancelled", connID)
	return nil
}

// SaveConnectionState 保存连接状态
func (rm *ReconnectManagerImpl) SaveConnectionState(connID string, state *ConnectionState) error {
	rm.stateMu.Lock()
	defer rm.stateMu.Unlock()

	state.ConnID = connID
	state.UpdatedAt = time.Now()

	rm.connectionStates[connID] = state
	rm.updateStats("state_saved", connID)
	return nil
}

// RestoreConnectionState 恢复连接状态
func (rm *ReconnectManagerImpl) RestoreConnectionState(connID string) (*ConnectionState, error) {
	rm.stateMu.RLock()
	defer rm.stateMu.RUnlock()

	state, exists := rm.connectionStates[connID]
	if !exists {
		return nil, &ReconnectError{
			Type:    "STATE_NOT_FOUND",
			Message: "连接状态不存在",
			ConnID:  connID,
		}
	}

	rm.updateStats("state_restored", connID)
	return state, nil
}

// SetReconnectStrategy 设置重连策略
func (rm *ReconnectManagerImpl) SetReconnectStrategy(strategy ReconnectStrategy) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.config.Strategy = strategy
}

// SetMaxReconnectAttempts 设置最大重连次数
func (rm *ReconnectManagerImpl) SetMaxReconnectAttempts(max int) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.config.MaxReconnectAttempts = max
}

// SetReconnectInterval 设置重连间隔
func (rm *ReconnectManagerImpl) SetReconnectInterval(interval time.Duration) {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	rm.config.ReconnectInterval = interval
}

// GetReconnectStats 获取重连统计
func (rm *ReconnectManagerImpl) GetReconnectStats() *ReconnectStats {
	rm.statsMu.RLock()
	defer rm.statsMu.RUnlock()

	// 返回副本
	stats := *rm.stats
	stats.ConnectionStats = make(map[string]int64)
	for k, v := range rm.stats.ConnectionStats {
		stats.ConnectionStats[k] = v
	}

	// 更新活跃重连数
	rm.connMu.RLock()
	activeReconnects := 0
	for _, conn := range rm.connections {
		if conn.IsReconnecting {
			activeReconnects++
		}
	}
	rm.connMu.RUnlock()
	stats.ActiveReconnects = activeReconnects

	return &stats
}

// GetConnectionReconnectStatus 获取连接重连状态
func (rm *ReconnectManagerImpl) GetConnectionReconnectStatus(connID string) *ConnectionReconnectStatus {
	rm.connMu.RLock()
	defer rm.connMu.RUnlock()

	conn, exists := rm.connections[connID]
	if !exists {
		return nil
	}

	// 返回副本
	status := *conn
	return &status
}

// GetAllReconnectStatus 获取所有重连状态
func (rm *ReconnectManagerImpl) GetAllReconnectStatus() map[string]*ConnectionReconnectStatus {
	rm.connMu.RLock()
	defer rm.connMu.RUnlock()

	status := make(map[string]*ConnectionReconnectStatus)
	for connID, conn := range rm.connections {
		status[connID] = &ConnectionReconnectStatus{
			ConnID:               conn.ConnID,
			IsConnected:          conn.IsConnected,
			IsReconnecting:       conn.IsReconnecting,
			ReconnectAttempts:    conn.ReconnectAttempts,
			LastReconnectTime:    conn.LastReconnectTime,
			NextReconnectTime:    conn.NextReconnectTime,
			ReconnectInterval:    conn.ReconnectInterval,
			TotalReconnects:      conn.TotalReconnects,
			SuccessfulReconnects: conn.SuccessfulReconnects,
			FailedReconnects:     conn.FailedReconnects,
			LastDisconnectTime:   conn.LastDisconnectTime,
			LastConnectTime:      conn.LastConnectTime,
			CreatedAt:            conn.CreatedAt,
		}
	}

	return status
}

// 辅助方法

func (rm *ReconnectManagerImpl) triggerReconnect(connID string) {
	rm.connMu.RLock()
	conn, exists := rm.connections[connID]
	if !exists {
		rm.connMu.RUnlock()
		return
	}
	rm.connMu.RUnlock()

	// 更新连接状态
	rm.connMu.Lock()
	conn.IsReconnecting = true
	conn.ReconnectAttempts++
	conn.LastReconnectTime = time.Now()
	rm.connMu.Unlock()

	// 发送到重连队列
	select {
	case rm.reconnectQueue <- connID:
		rm.updateStats("reconnect_triggered", connID)
	default:
		// 队列满，记录失败
		rm.updateStats("reconnect_queue_full", connID)
	}
}

func (rm *ReconnectManagerImpl) reconnectLoop() {
	for {
		select {
		case <-rm.ctx.Done():
			return
		case connID := <-rm.reconnectQueue:
			rm.processReconnect(connID)
		}
	}
}

func (rm *ReconnectManagerImpl) cleanupLoop() {
	ticker := time.NewTicker(rm.config.CleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-rm.ctx.Done():
			return
		case <-ticker.C:
			rm.cleanupInactiveConnections()
		}
	}
}

func (rm *ReconnectManagerImpl) processReconnect(connID string) {
	rm.connMu.RLock()
	conn, exists := rm.connections[connID]
	rm.connMu.RUnlock()

	if !exists {
		return
	}

	// 检查是否超过最大重连次数
	if conn.ReconnectAttempts >= rm.config.MaxReconnectAttempts {
		rm.handleMaxReconnectAttemptsReached(connID)
		return
	}

	// 计算重连间隔
	interval := rm.calculateReconnectInterval(conn)

	// 等待重连间隔
	time.Sleep(interval)

	// 执行重连
	rm.performReconnect(connID)
}

func (rm *ReconnectManagerImpl) calculateReconnectInterval(conn *ConnectionReconnectStatus) time.Duration {
	switch rm.config.Strategy {
	case ReconnectStrategyLinear:
		return rm.config.ReconnectInterval
	case ReconnectStrategyExponential:
		interval := time.Duration(float64(rm.config.ReconnectInterval) *
			pow(2.0, float64(conn.ReconnectAttempts-1)))
		if interval > rm.config.ReconnectInterval*10 {
			interval = rm.config.ReconnectInterval * 10
		}
		return interval
	case ReconnectStrategyFixed:
		return rm.config.ReconnectInterval
	default:
		return rm.config.ReconnectInterval
	}
}

func (rm *ReconnectManagerImpl) performReconnect(connID string) {
	rm.connMu.Lock()
	conn, exists := rm.connections[connID]
	if !exists {
		rm.connMu.Unlock()
		return
	}
	rm.connMu.Unlock()

	// 模拟重连过程
	success := rm.simulateReconnect(connID)

	rm.connMu.Lock()
	conn, exists = rm.connections[connID]
	if !exists {
		rm.connMu.Unlock()
		return
	}

	if success {
		conn.IsReconnecting = false
		conn.IsConnected = true
		conn.ReconnectAttempts = 0
		conn.SuccessfulReconnects++
		conn.LastConnectTime = time.Now()
		rm.connMu.Unlock()

		rm.updateStats("reconnect_successful", connID)

		// 恢复连接状态
		if rm.config.StateRecoveryEnabled {
			rm.restoreConnectionState(connID)
		}
	} else {
		conn.FailedReconnects++
		rm.connMu.Unlock()

		rm.updateStats("reconnect_failed", connID)

		// 如果还有重连次数，继续重连
		if conn.ReconnectAttempts < rm.config.MaxReconnectAttempts {
			rm.triggerReconnect(connID)
		}
	}
}

func (rm *ReconnectManagerImpl) simulateReconnect(connID string) bool {
	// 模拟重连过程，实际实现中这里会调用真实的连接逻辑
	// 这里简单模拟成功率
	return true // 简化测试，总是成功
}

func (rm *ReconnectManagerImpl) restoreConnectionState(connID string) {
	state, err := rm.RestoreConnectionState(connID)
	if err != nil {
		// 状态恢复失败，记录日志
		return
	}

	// 这里可以恢复订阅、配置等状态
	// 实际实现中会调用相应的恢复逻辑
	_ = state
}

func (rm *ReconnectManagerImpl) handleMaxReconnectAttemptsReached(connID string) {
	rm.connMu.Lock()
	conn, exists := rm.connections[connID]
	if !exists {
		rm.connMu.Unlock()
		return
	}

	conn.IsReconnecting = false
	rm.connMu.Unlock()

	rm.updateStats("max_reconnect_attempts_reached", connID)
}

func (rm *ReconnectManagerImpl) cleanupInactiveConnections() {
	rm.connMu.Lock()
	defer rm.connMu.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for connID, conn := range rm.connections {
		// 清理长时间不活跃且不在重连的连接
		if !conn.IsConnected && !conn.IsReconnecting &&
			now.Sub(conn.LastDisconnectTime) > rm.config.CleanupInterval {
			toRemove = append(toRemove, connID)
		}
	}

	for _, connID := range toRemove {
		delete(rm.connections, connID)
		rm.updateStats("connection_cleaned", connID)
	}
}

func (rm *ReconnectManagerImpl) updateStats(event, connID string) {
	rm.statsMu.Lock()
	defer rm.statsMu.Unlock()

	switch event {
	case "reconnect_triggered":
		rm.stats.TotalReconnects++
		rm.stats.LastReconnectTime = time.Now()
		rm.stats.ConnectionStats[connID]++
	case "reconnect_successful":
		rm.stats.SuccessfulReconnects++
	case "reconnect_failed":
		rm.stats.FailedReconnects++
	case "connection_added", "connection_removed", "connection_connected",
		"connection_disconnected", "reconnect_cancelled", "state_saved",
		"state_restored", "max_reconnect_attempts_reached", "connection_cleaned":
		// 这些事件不需要特殊处理
	}
}

// 工具函数
func pow(base, exp float64) float64 {
	result := 1.0
	for i := 0; i < int(exp); i++ {
		result *= base
	}
	return result
}

// ReconnectError 重连错误
type ReconnectError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	ConnID  string `json:"conn_id,omitempty"`
}

func (e *ReconnectError) Error() string {
	return e.Message
}
