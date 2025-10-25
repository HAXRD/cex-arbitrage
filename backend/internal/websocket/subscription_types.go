package websocket

import (
	"sync"
	"time"
)

// SubscriptionManager 订阅管理器接口
type SubscriptionManager interface {
	// 订阅管理
	Subscribe(connID string, symbols []string) error
	Unsubscribe(connID string, symbols []string) error
	GetSubscriptions(connID string) []string
	GetSubscribers(symbol string) []string

	// 连接管理
	AddConnection(connID string) error
	RemoveConnection(connID string) error
	IsConnectionActive(connID string) bool

	// 统计和查询
	GetConnectionCount() int
	GetSymbolCount() int
	GetSubscriptionStats() *SubscriptionStats
	GetAllSubscriptions() map[string][]string

	// 清理和维护
	CleanupInactiveConnections() int
	GetInactiveConnections() []string
}

// SubscriptionStats 订阅统计信息
type SubscriptionStats struct {
	TotalConnections        int            `json:"total_connections"`
	ActiveConnections       int            `json:"active_connections"`
	TotalSubscriptions      int            `json:"total_subscriptions"`
	SymbolSubscriptions     map[string]int `json:"symbol_subscriptions"`
	ConnectionSubscriptions map[string]int `json:"connection_subscriptions"`
	LastUpdated             time.Time      `json:"last_updated"`
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	ID            string    `json:"id"`
	IsActive      bool      `json:"is_active"`
	Subscriptions []string  `json:"subscriptions"`
	CreatedAt     time.Time `json:"created_at"`
	LastActivity  time.Time `json:"last_activity"`
}

// SymbolInfo 交易对信息
type SymbolInfo struct {
	Symbol          string    `json:"symbol"`
	SubscriberCount int       `json:"subscriber_count"`
	Subscribers     []string  `json:"subscribers"`
	LastUpdated     time.Time `json:"last_updated"`
}

// SubscriptionManagerImpl 订阅管理器实现
type SubscriptionManagerImpl struct {
	// 连接 -> 订阅的交易对列表
	connections map[string]*ConnectionInfo
	// 交易对 -> 订阅者连接ID列表
	symbols map[string][]string
	// 读写锁
	mu sync.RWMutex
	// 日志记录器
	logger interface{}
}

// NewSubscriptionManager 创建订阅管理器
func NewSubscriptionManager(logger interface{}) SubscriptionManager {
	return &SubscriptionManagerImpl{
		connections: make(map[string]*ConnectionInfo),
		symbols:     make(map[string][]string),
		logger:      logger,
	}
}

// Subscribe 订阅交易对
func (sm *SubscriptionManagerImpl) Subscribe(connID string, symbols []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查连接是否存在
	conn, exists := sm.connections[connID]
	if !exists {
		return &SubscriptionError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	// 添加订阅
	for _, symbol := range symbols {
		// 添加到连接的订阅列表
		if !contains(conn.Subscriptions, symbol) {
			conn.Subscriptions = append(conn.Subscriptions, symbol)
		}

		// 添加到交易对的订阅者列表
		if !contains(sm.symbols[symbol], connID) {
			sm.symbols[symbol] = append(sm.symbols[symbol], connID)
		}
	}

	conn.LastActivity = time.Now()
	return nil
}

// Unsubscribe 取消订阅交易对
func (sm *SubscriptionManagerImpl) Unsubscribe(connID string, symbols []string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查连接是否存在
	conn, exists := sm.connections[connID]
	if !exists {
		return &SubscriptionError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	// 移除订阅
	for _, symbol := range symbols {
		// 从连接的订阅列表移除
		conn.Subscriptions = removeFromSlice(conn.Subscriptions, symbol)

		// 从交易对的订阅者列表移除
		sm.symbols[symbol] = removeFromSlice(sm.symbols[symbol], connID)

		// 如果交易对没有订阅者，删除该交易对
		if len(sm.symbols[symbol]) == 0 {
			delete(sm.symbols, symbol)
		}
	}

	conn.LastActivity = time.Now()
	return nil
}

// GetSubscriptions 获取连接的订阅列表
func (sm *SubscriptionManagerImpl) GetSubscriptions(connID string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	conn, exists := sm.connections[connID]
	if !exists {
		return nil
	}

	// 返回副本，避免外部修改
	result := make([]string, len(conn.Subscriptions))
	copy(result, conn.Subscriptions)
	return result
}

// GetSubscribers 获取交易对的订阅者列表
func (sm *SubscriptionManagerImpl) GetSubscribers(symbol string) []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	subscribers, exists := sm.symbols[symbol]
	if !exists {
		return nil
	}

	// 返回副本，避免外部修改
	result := make([]string, len(subscribers))
	copy(result, subscribers)
	return result
}

// AddConnection 添加连接
func (sm *SubscriptionManagerImpl) AddConnection(connID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if _, exists := sm.connections[connID]; exists {
		return &SubscriptionError{
			Type:    "CONNECTION_EXISTS",
			Message: "连接已存在",
			ConnID:  connID,
		}
	}

	sm.connections[connID] = &ConnectionInfo{
		ID:            connID,
		IsActive:      true,
		Subscriptions: make([]string, 0),
		CreatedAt:     time.Now(),
		LastActivity:  time.Now(),
	}

	return nil
}

// RemoveConnection 移除连接
func (sm *SubscriptionManagerImpl) RemoveConnection(connID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	conn, exists := sm.connections[connID]
	if !exists {
		return &SubscriptionError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	// 清理所有订阅
	for _, symbol := range conn.Subscriptions {
		sm.symbols[symbol] = removeFromSlice(sm.symbols[symbol], connID)
		if len(sm.symbols[symbol]) == 0 {
			delete(sm.symbols, symbol)
		}
	}

	delete(sm.connections, connID)
	return nil
}

// IsConnectionActive 检查连接是否活跃
func (sm *SubscriptionManagerImpl) IsConnectionActive(connID string) bool {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	conn, exists := sm.connections[connID]
	return exists && conn.IsActive
}

// GetConnectionCount 获取连接数量
func (sm *SubscriptionManagerImpl) GetConnectionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.connections)
}

// GetSymbolCount 获取交易对数量
func (sm *SubscriptionManagerImpl) GetSymbolCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.symbols)
}

// GetSubscriptionStats 获取订阅统计信息
func (sm *SubscriptionManagerImpl) GetSubscriptionStats() *SubscriptionStats {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	stats := &SubscriptionStats{
		TotalConnections:        len(sm.connections),
		ActiveConnections:       0,
		TotalSubscriptions:      0,
		SymbolSubscriptions:     make(map[string]int),
		ConnectionSubscriptions: make(map[string]int),
		LastUpdated:             time.Now(),
	}

	// 统计活跃连接
	for _, conn := range sm.connections {
		if conn.IsActive {
			stats.ActiveConnections++
		}
		stats.ConnectionSubscriptions[conn.ID] = len(conn.Subscriptions)
		stats.TotalSubscriptions += len(conn.Subscriptions)
	}

	// 统计交易对订阅
	for symbol, subscribers := range sm.symbols {
		stats.SymbolSubscriptions[symbol] = len(subscribers)
	}

	return stats
}

// GetAllSubscriptions 获取所有订阅关系
func (sm *SubscriptionManagerImpl) GetAllSubscriptions() map[string][]string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	result := make(map[string][]string)
	for symbol, subscribers := range sm.symbols {
		result[symbol] = make([]string, len(subscribers))
		copy(result[symbol], subscribers)
	}
	return result
}

// CleanupInactiveConnections 清理非活跃连接
func (sm *SubscriptionManagerImpl) CleanupInactiveConnections() int {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	cleanedCount := 0
	cutoffTime := time.Now().Add(-5 * time.Minute) // 5分钟未活跃

	for connID, conn := range sm.connections {
		if !conn.IsActive || conn.LastActivity.Before(cutoffTime) {
			// 清理订阅
			for _, symbol := range conn.Subscriptions {
				sm.symbols[symbol] = removeFromSlice(sm.symbols[symbol], connID)
				if len(sm.symbols[symbol]) == 0 {
					delete(sm.symbols, symbol)
				}
			}
			delete(sm.connections, connID)
			cleanedCount++
		}
	}

	return cleanedCount
}

// GetInactiveConnections 获取非活跃连接列表
func (sm *SubscriptionManagerImpl) GetInactiveConnections() []string {
	sm.mu.RLock()
	defer sm.mu.RUnlock()

	var inactive []string
	cutoffTime := time.Now().Add(-5 * time.Minute) // 5分钟未活跃

	for connID, conn := range sm.connections {
		if !conn.IsActive || conn.LastActivity.Before(cutoffTime) {
			inactive = append(inactive, connID)
		}
	}

	return inactive
}

// SubscriptionError 订阅错误
type SubscriptionError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	ConnID  string `json:"conn_id,omitempty"`
	Symbol  string `json:"symbol,omitempty"`
}

func (e *SubscriptionError) Error() string {
	return e.Message
}
