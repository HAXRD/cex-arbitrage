package websocket

import (
	"context"
	"sync"
	"time"
)

// HeartbeatManager 心跳管理器接口
type HeartbeatManager interface {
	// 心跳管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 连接管理
	AddConnection(connID string) error
	RemoveConnection(connID string) error
	UpdateLastPing(connID string) error

	// 心跳控制
	SendHeartbeat(connID string) error
	ProcessPong(connID string, pongData string) error

	// 配置管理
	SetHeartbeatInterval(interval time.Duration)
	SetPongTimeout(timeout time.Duration)
	SetMaxMissedHeartbeats(max int)

	// 统计和监控
	GetHeartbeatStats() *HeartbeatStats
	GetConnectionStatus(connID string) *ConnectionHeartbeatStatus
	GetAllConnectionStatus() map[string]*ConnectionHeartbeatStatus
}

// HeartbeatStats 心跳统计信息
type HeartbeatStats struct {
	TotalHeartbeatsSent int64            `json:"total_heartbeats_sent"`
	TotalPongsReceived  int64            `json:"total_pongs_received"`
	TotalTimeouts       int64            `json:"total_timeouts"`
	ActiveConnections   int              `json:"active_connections"`
	FailedHeartbeats    int64            `json:"failed_heartbeats"`
	AverageResponseTime time.Duration    `json:"average_response_time"`
	LastHeartbeatSent   time.Time        `json:"last_heartbeat_sent"`
	LastPongReceived    time.Time        `json:"last_pong_received"`
	StartTime           time.Time        `json:"start_time"`
	ConnectionStats     map[string]int64 `json:"connection_stats"`
}

// ConnectionHeartbeatStatus 连接心跳状态
type ConnectionHeartbeatStatus struct {
	ConnID            string        `json:"conn_id"`
	IsActive          bool          `json:"is_active"`
	LastHeartbeatSent time.Time     `json:"last_heartbeat_sent"`
	LastPongReceived  time.Time     `json:"last_pong_received"`
	MissedHeartbeats  int           `json:"missed_heartbeats"`
	ResponseTime      time.Duration `json:"response_time"`
	TotalHeartbeats   int64         `json:"total_heartbeats"`
	TotalPongs        int64         `json:"total_pongs"`
	CreatedAt         time.Time     `json:"created_at"`
}

// HeartbeatConfig 心跳配置
type HeartbeatConfig struct {
	HeartbeatInterval    time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`
	PongTimeout          time.Duration `json:"pong_timeout" yaml:"pong_timeout"`
	MaxMissedHeartbeats  int           `json:"max_missed_heartbeats" yaml:"max_missed_heartbeats"`
	HeartbeatMessage     string        `json:"heartbeat_message" yaml:"heartbeat_message"`
	PongMessage          string        `json:"pong_message" yaml:"pong_message"`
	EnableHeartbeatStats bool          `json:"enable_heartbeat_stats" yaml:"enable_heartbeat_stats"`
	CleanupInterval      time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`
}

// DefaultHeartbeatConfig 默认心跳配置
func DefaultHeartbeatConfig() *HeartbeatConfig {
	return &HeartbeatConfig{
		HeartbeatInterval:    30 * time.Second,
		PongTimeout:          10 * time.Second,
		MaxMissedHeartbeats:  3,
		HeartbeatMessage:     "ping",
		PongMessage:          "pong",
		EnableHeartbeatStats: true,
		CleanupInterval:      60 * time.Second,
	}
}

// HeartbeatManagerImpl 心跳管理器实现
type HeartbeatManagerImpl struct {
	// 配置
	config *HeartbeatConfig

	// 连接管理器
	connManager ConnectionManager

	// 连接状态
	connections map[string]*ConnectionHeartbeatStatus
	connMu      sync.RWMutex

	// 控制
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.RWMutex

	// 统计
	stats   *HeartbeatStats
	statsMu sync.RWMutex

	// 定时器
	heartbeatTicker *time.Ticker
	cleanupTicker   *time.Ticker

	// 日志记录器
	logger interface{}
}

// NewHeartbeatManager 创建心跳管理器
func NewHeartbeatManager(config *HeartbeatConfig, connManager ConnectionManager, logger interface{}) HeartbeatManager {
	if config == nil {
		config = DefaultHeartbeatConfig()
	}

	return &HeartbeatManagerImpl{
		config:      config,
		connManager: connManager,
		connections: make(map[string]*ConnectionHeartbeatStatus),
		stats: &HeartbeatStats{
			ConnectionStats: make(map[string]int64),
			StartTime:       time.Now(),
		},
		logger: logger,
	}
}

// Start 启动心跳管理器
func (hm *HeartbeatManagerImpl) Start(ctx context.Context) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if hm.running {
		return &HeartbeatError{
			Type:    "ALREADY_RUNNING",
			Message: "心跳管理器已在运行",
		}
	}

	hm.ctx, hm.cancel = context.WithCancel(ctx)
	hm.running = true

	// 启动心跳定时器
	hm.heartbeatTicker = time.NewTicker(hm.config.HeartbeatInterval)
	go hm.heartbeatLoop()

	// 启动清理定时器
	hm.cleanupTicker = time.NewTicker(hm.config.CleanupInterval)
	go hm.cleanupLoop()

	return nil
}

// Stop 停止心跳管理器
func (hm *HeartbeatManagerImpl) Stop(ctx context.Context) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	if !hm.running {
		return &HeartbeatError{
			Type:    "NOT_RUNNING",
			Message: "心跳管理器未运行",
		}
	}

	hm.running = false
	hm.cancel()

	// 停止定时器
	if hm.heartbeatTicker != nil {
		hm.heartbeatTicker.Stop()
	}
	if hm.cleanupTicker != nil {
		hm.cleanupTicker.Stop()
	}

	return nil
}

// IsRunning 检查是否运行
func (hm *HeartbeatManagerImpl) IsRunning() bool {
	hm.mu.RLock()
	defer hm.mu.RUnlock()
	return hm.running
}

// AddConnection 添加连接
func (hm *HeartbeatManagerImpl) AddConnection(connID string) error {
	hm.connMu.Lock()
	defer hm.connMu.Unlock()

	hm.connections[connID] = &ConnectionHeartbeatStatus{
		ConnID:           connID,
		IsActive:         true,
		CreatedAt:        time.Now(),
		MissedHeartbeats: 0,
	}

	hm.updateStats("connection_added", connID)
	return nil
}

// RemoveConnection 移除连接
func (hm *HeartbeatManagerImpl) RemoveConnection(connID string) error {
	hm.connMu.Lock()
	defer hm.connMu.Unlock()

	delete(hm.connections, connID)
	hm.updateStats("connection_removed", connID)
	return nil
}

// UpdateLastPing 更新最后ping时间
func (hm *HeartbeatManagerImpl) UpdateLastPing(connID string) error {
	hm.connMu.Lock()
	defer hm.connMu.Unlock()

	conn, exists := hm.connections[connID]
	if !exists {
		return &HeartbeatError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	conn.LastHeartbeatSent = time.Now()
	hm.updateStats("heartbeat_sent", connID)
	return nil
}

// SendHeartbeat 发送心跳
func (hm *HeartbeatManagerImpl) SendHeartbeat(connID string) error {
	hm.connMu.RLock()
	conn, exists := hm.connections[connID]
	hm.connMu.RUnlock()

	if !exists {
		return &HeartbeatError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	// 检查连接是否活跃
	if !hm.connManager.IsConnectionActive(connID) {
		hm.markConnectionInactive(connID)
		return &HeartbeatError{
			Type:    "CONNECTION_INACTIVE",
			Message: "连接不活跃",
			ConnID:  connID,
		}
	}

	// 发送心跳消息
	heartbeatMsg := map[string]interface{}{
		"type":      "heartbeat",
		"message":   hm.config.HeartbeatMessage,
		"timestamp": time.Now().Unix(),
	}

	err := hm.connManager.SendToConnection(connID, heartbeatMsg)
	if err != nil {
		hm.updateFailureStats(connID)
		return err
	}

	// 更新状态
	hm.connMu.Lock()
	conn.LastHeartbeatSent = time.Now()
	conn.TotalHeartbeats++
	hm.connMu.Unlock()

	hm.updateStats("heartbeat_sent", connID)
	return nil
}

// ProcessPong 处理pong响应
func (hm *HeartbeatManagerImpl) ProcessPong(connID string, pongData string) error {
	hm.connMu.Lock()
	defer hm.connMu.Unlock()

	conn, exists := hm.connections[connID]
	if !exists {
		return &HeartbeatError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	// 更新pong时间
	conn.LastPongReceived = time.Now()
	conn.TotalPongs++
	conn.MissedHeartbeats = 0 // 重置未收到的心跳计数

	// 计算响应时间
	if !conn.LastHeartbeatSent.IsZero() {
		conn.ResponseTime = conn.LastPongReceived.Sub(conn.LastHeartbeatSent)
		hm.updateResponseTimeStats(conn.ResponseTime)
	}

	hm.updateStats("pong_received", connID)
	return nil
}

// SetHeartbeatInterval 设置心跳间隔
func (hm *HeartbeatManagerImpl) SetHeartbeatInterval(interval time.Duration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.config.HeartbeatInterval = interval
}

// SetPongTimeout 设置pong超时时间
func (hm *HeartbeatManagerImpl) SetPongTimeout(timeout time.Duration) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.config.PongTimeout = timeout
}

// SetMaxMissedHeartbeats 设置最大未收到心跳次数
func (hm *HeartbeatManagerImpl) SetMaxMissedHeartbeats(max int) {
	hm.mu.Lock()
	defer hm.mu.Unlock()
	hm.config.MaxMissedHeartbeats = max
}

// GetHeartbeatStats 获取心跳统计
func (hm *HeartbeatManagerImpl) GetHeartbeatStats() *HeartbeatStats {
	hm.statsMu.RLock()
	defer hm.statsMu.RUnlock()

	// 返回副本
	stats := *hm.stats
	stats.ConnectionStats = make(map[string]int64)
	for k, v := range hm.stats.ConnectionStats {
		stats.ConnectionStats[k] = v
	}

	// 更新活跃连接数
	hm.connMu.RLock()
	stats.ActiveConnections = len(hm.connections)
	hm.connMu.RUnlock()

	return &stats
}

// GetConnectionStatus 获取连接状态
func (hm *HeartbeatManagerImpl) GetConnectionStatus(connID string) *ConnectionHeartbeatStatus {
	hm.connMu.RLock()
	defer hm.connMu.RUnlock()

	conn, exists := hm.connections[connID]
	if !exists {
		return nil
	}

	// 返回副本
	status := *conn
	return &status
}

// GetAllConnectionStatus 获取所有连接状态
func (hm *HeartbeatManagerImpl) GetAllConnectionStatus() map[string]*ConnectionHeartbeatStatus {
	hm.connMu.RLock()
	defer hm.connMu.RUnlock()

	status := make(map[string]*ConnectionHeartbeatStatus)
	for connID, conn := range hm.connections {
		status[connID] = &ConnectionHeartbeatStatus{
			ConnID:            conn.ConnID,
			IsActive:          conn.IsActive,
			LastHeartbeatSent: conn.LastHeartbeatSent,
			LastPongReceived:  conn.LastPongReceived,
			MissedHeartbeats:  conn.MissedHeartbeats,
			ResponseTime:      conn.ResponseTime,
			TotalHeartbeats:   conn.TotalHeartbeats,
			TotalPongs:        conn.TotalPongs,
			CreatedAt:         conn.CreatedAt,
		}
	}

	return status
}

// 辅助方法

func (hm *HeartbeatManagerImpl) heartbeatLoop() {
	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-hm.heartbeatTicker.C:
			hm.processHeartbeats()
		}
	}
}

func (hm *HeartbeatManagerImpl) cleanupLoop() {
	for {
		select {
		case <-hm.ctx.Done():
			return
		case <-hm.cleanupTicker.C:
			hm.cleanupInactiveConnections()
		}
	}
}

func (hm *HeartbeatManagerImpl) processHeartbeats() {
	hm.connMu.RLock()
	connections := make([]string, 0, len(hm.connections))
	for connID := range hm.connections {
		connections = append(connections, connID)
	}
	hm.connMu.RUnlock()

	for _, connID := range connections {
		hm.processConnectionHeartbeat(connID)
	}
}

func (hm *HeartbeatManagerImpl) processConnectionHeartbeat(connID string) {
	hm.connMu.Lock()
	conn, exists := hm.connections[connID]
	if !exists {
		hm.connMu.Unlock()
		return
	}
	hm.connMu.Unlock()

	// 检查是否超时
	if hm.isConnectionTimeout(conn) {
		hm.handleConnectionTimeout(connID)
		return
	}

	// 发送心跳
	err := hm.SendHeartbeat(connID)
	if err != nil {
		hm.handleHeartbeatError(connID, err)
	}
}

func (hm *HeartbeatManagerImpl) isConnectionTimeout(conn *ConnectionHeartbeatStatus) bool {
	// 检查是否有未响应的心跳
	if !conn.LastHeartbeatSent.IsZero() && !conn.LastPongReceived.IsZero() {
		timeSinceLastHeartbeat := time.Since(conn.LastHeartbeatSent)

		// 如果最后的心跳时间比最后的pong时间晚，说明有心跳未响应
		if conn.LastHeartbeatSent.After(conn.LastPongReceived) {
			return timeSinceLastHeartbeat > hm.config.PongTimeout
		}
	}

	// 检查未收到心跳次数
	return conn.MissedHeartbeats >= hm.config.MaxMissedHeartbeats
}

func (hm *HeartbeatManagerImpl) handleConnectionTimeout(connID string) {
	hm.connMu.Lock()
	conn, exists := hm.connections[connID]
	if !exists {
		hm.connMu.Unlock()
		return
	}

	conn.MissedHeartbeats++
	hm.connMu.Unlock()

	hm.updateStats("timeout", connID)

	// 如果超过最大未收到心跳次数，标记为不活跃
	if conn.MissedHeartbeats >= hm.config.MaxMissedHeartbeats {
		hm.markConnectionInactive(connID)
	}
}

func (hm *HeartbeatManagerImpl) handleHeartbeatError(connID string, err error) {
	hm.connMu.Lock()
	conn, exists := hm.connections[connID]
	if !exists {
		hm.connMu.Unlock()
		return
	}

	conn.MissedHeartbeats++
	hm.connMu.Unlock()

	hm.updateFailureStats(connID)
}

func (hm *HeartbeatManagerImpl) markConnectionInactive(connID string) {
	hm.connMu.Lock()
	defer hm.connMu.Unlock()

	conn, exists := hm.connections[connID]
	if !exists {
		return
	}

	conn.IsActive = false
	hm.updateStats("connection_inactive", connID)
}

func (hm *HeartbeatManagerImpl) cleanupInactiveConnections() {
	hm.connMu.Lock()
	defer hm.connMu.Unlock()

	now := time.Now()
	toRemove := make([]string, 0)

	for connID, conn := range hm.connections {
		// 清理不活跃且超过清理间隔的连接
		if !conn.IsActive && now.Sub(conn.LastPongReceived) > hm.config.CleanupInterval {
			toRemove = append(toRemove, connID)
		}
	}

	for _, connID := range toRemove {
		delete(hm.connections, connID)
		hm.updateStats("connection_cleaned", connID)
	}
}

func (hm *HeartbeatManagerImpl) updateStats(event, connID string) {
	hm.statsMu.Lock()
	defer hm.statsMu.Unlock()

	switch event {
	case "heartbeat_sent":
		hm.stats.TotalHeartbeatsSent++
		hm.stats.LastHeartbeatSent = time.Now()
		hm.stats.ConnectionStats[connID]++
	case "pong_received":
		hm.stats.TotalPongsReceived++
		hm.stats.LastPongReceived = time.Now()
	case "timeout":
		hm.stats.TotalTimeouts++
	case "connection_added", "connection_removed", "connection_inactive", "connection_cleaned":
		// 这些事件不需要特殊处理
	}
}

func (hm *HeartbeatManagerImpl) updateFailureStats(connID string) {
	hm.statsMu.Lock()
	defer hm.statsMu.Unlock()
	hm.stats.FailedHeartbeats++
}

func (hm *HeartbeatManagerImpl) updateResponseTimeStats(responseTime time.Duration) {
	hm.statsMu.Lock()
	defer hm.statsMu.Unlock()

	if hm.stats.AverageResponseTime == 0 {
		hm.stats.AverageResponseTime = responseTime
	} else {
		hm.stats.AverageResponseTime = (hm.stats.AverageResponseTime + responseTime) / 2
	}
}

// HeartbeatError 心跳错误
type HeartbeatError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	ConnID  string `json:"conn_id,omitempty"`
}

func (e *HeartbeatError) Error() string {
	return e.Message
}
