package websocket

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// BroadcastManager 消息广播管理器接口
type BroadcastManager interface {
	// 广播方法
	BroadcastToSymbol(symbol string, message interface{}) error
	BroadcastToAll(message interface{}) error
	SendToConnection(connID string, message interface{}) error

	// 批量广播
	BatchBroadcastToSymbol(symbol string, messages []interface{}) error
	BatchBroadcastToAll(messages []interface{}) error

	// 广播管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 统计和监控
	GetBroadcastStats() *BroadcastStats
	GetQueueSize() int
	GetFailedCount() int64

	// 配置管理
	SetMaxQueueSize(size int)
	SetRetryAttempts(attempts int)
	SetRetryDelay(delay time.Duration)
}

// BroadcastMessage 广播消息
type BroadcastMessage struct {
	ID         string      `json:"id"`
	Type       string      `json:"type"`
	Symbol     string      `json:"symbol,omitempty"`
	ConnID     string      `json:"conn_id,omitempty"`
	Data       interface{} `json:"data"`
	Timestamp  time.Time   `json:"timestamp"`
	RetryCount int         `json:"retry_count"`
	Priority   int         `json:"priority"` // 优先级，数字越小优先级越高
}

// BroadcastStats 广播统计信息
type BroadcastStats struct {
	TotalBroadcasts      int64            `json:"total_broadcasts"`
	SuccessfulBroadcasts int64            `json:"successful_broadcasts"`
	FailedBroadcasts     int64            `json:"failed_broadcasts"`
	QueueSize            int              `json:"queue_size"`
	AverageLatency       time.Duration    `json:"average_latency"`
	SymbolStats          map[string]int64 `json:"symbol_stats"`
	LastBroadcast        time.Time        `json:"last_broadcast"`
	StartTime            time.Time        `json:"start_time"`
}

// BroadcastConfig 广播配置
type BroadcastConfig struct {
	MaxQueueSize    int           `json:"max_queue_size" yaml:"max_queue_size"`
	RetryAttempts   int           `json:"retry_attempts" yaml:"retry_attempts"`
	RetryDelay      time.Duration `json:"retry_delay" yaml:"retry_delay"`
	BatchSize       int           `json:"batch_size" yaml:"batch_size"`
	FlushInterval   time.Duration `json:"flush_interval" yaml:"flush_interval"`
	WorkerCount     int           `json:"worker_count" yaml:"worker_count"`
	PriorityEnabled bool          `json:"priority_enabled" yaml:"priority_enabled"`
}

// DefaultBroadcastConfig 默认广播配置
func DefaultBroadcastConfig() *BroadcastConfig {
	return &BroadcastConfig{
		MaxQueueSize:    1000,
		RetryAttempts:   3,
		RetryDelay:      100 * time.Millisecond,
		BatchSize:       10,
		FlushInterval:   50 * time.Millisecond,
		WorkerCount:     5,
		PriorityEnabled: true,
	}
}

// BroadcastManagerImpl 消息广播管理器实现
type BroadcastManagerImpl struct {
	// 配置
	config *BroadcastConfig

	// 连接管理器
	connManager ConnectionManager

	// 消息队列
	messageQueue  chan *BroadcastMessage
	priorityQueue chan *BroadcastMessage

	// 控制
	ctx     context.Context
	cancel  context.CancelFunc
	running bool
	mu      sync.RWMutex

	// 统计
	stats   *BroadcastStats
	statsMu sync.RWMutex

	// 日志记录器
	logger interface{}
}

// ConnectionManager 连接管理器接口
type ConnectionManager interface {
	GetConnections() []*Connection
	GetSubscribers(symbol string) []string
	IsConnectionActive(connID string) bool
	SendToConnection(connID string, message interface{}) error
}

// NewBroadcastManager 创建广播管理器
func NewBroadcastManager(config *BroadcastConfig, connManager ConnectionManager, logger interface{}) BroadcastManager {
	if config == nil {
		config = DefaultBroadcastConfig()
	}

	return &BroadcastManagerImpl{
		config:        config,
		connManager:   connManager,
		messageQueue:  make(chan *BroadcastMessage, config.MaxQueueSize),
		priorityQueue: make(chan *BroadcastMessage, config.MaxQueueSize/2),
		stats: &BroadcastStats{
			SymbolStats: make(map[string]int64),
			StartTime:   time.Now(),
		},
		logger: logger,
	}
}

// BroadcastToSymbol 向特定交易对的所有订阅者广播消息
func (bm *BroadcastManagerImpl) BroadcastToSymbol(symbol string, message interface{}) error {
	bm.mu.RLock()
	if !bm.running {
		bm.mu.RUnlock()
		return &BroadcastError{
			Type:    "BROADCASTER_NOT_RUNNING",
			Message: "广播管理器未运行",
		}
	}
	bm.mu.RUnlock()

	// 获取订阅者
	subscribers := bm.connManager.GetSubscribers(symbol)
	if len(subscribers) == 0 {
		return &BroadcastError{
			Type:    "NO_SUBSCRIBERS",
			Message: "交易对没有订阅者",
			Symbol:  symbol,
		}
	}

	// 创建广播消息
	broadcastMsg := &BroadcastMessage{
		ID:         generateMessageID(),
		Type:       "symbol_broadcast",
		Symbol:     symbol,
		Data:       message,
		Timestamp:  time.Now(),
		RetryCount: 0,
		Priority:   1,
	}

	// 发送到队列
	select {
	case bm.messageQueue <- broadcastMsg:
		bm.updateStats("symbol_broadcast", symbol)
		return nil
	default:
		return &BroadcastError{
			Type:    "QUEUE_FULL",
			Message: "消息队列已满",
		}
	}
}

// BroadcastToAll 向所有连接广播消息
func (bm *BroadcastManagerImpl) BroadcastToAll(message interface{}) error {
	bm.mu.RLock()
	if !bm.running {
		bm.mu.RUnlock()
		return &BroadcastError{
			Type:    "BROADCASTER_NOT_RUNNING",
			Message: "广播管理器未运行",
		}
	}
	bm.mu.RUnlock()

	// 创建广播消息
	broadcastMsg := &BroadcastMessage{
		ID:         generateMessageID(),
		Type:       "broadcast_all",
		Data:       message,
		Timestamp:  time.Now(),
		RetryCount: 0,
		Priority:   2,
	}

	// 发送到队列
	select {
	case bm.messageQueue <- broadcastMsg:
		bm.updateStats("broadcast_all", "")
		return nil
	default:
		return &BroadcastError{
			Type:    "QUEUE_FULL",
			Message: "消息队列已满",
		}
	}
}

// SendToConnection 向特定连接发送消息
func (bm *BroadcastManagerImpl) SendToConnection(connID string, message interface{}) error {
	bm.mu.RLock()
	if !bm.running {
		bm.mu.RUnlock()
		return &BroadcastError{
			Type:    "BROADCASTER_NOT_RUNNING",
			Message: "广播管理器未运行",
		}
	}
	bm.mu.RUnlock()

	// 检查连接是否存在
	if !bm.connManager.IsConnectionActive(connID) {
		return &BroadcastError{
			Type:    "CONNECTION_NOT_FOUND",
			Message: "连接不存在",
			ConnID:  connID,
		}
	}

	// 创建单播消息
	unicastMsg := &BroadcastMessage{
		ID:         generateMessageID(),
		Type:       "unicast",
		ConnID:     connID,
		Data:       message,
		Timestamp:  time.Now(),
		RetryCount: 0,
		Priority:   0, // 单播消息优先级最高
	}

	// 发送到优先级队列
	select {
	case bm.priorityQueue <- unicastMsg:
		bm.updateStats("unicast", "")
		return nil
	default:
		return &BroadcastError{
			Type:    "QUEUE_FULL",
			Message: "消息队列已满",
		}
	}
}

// BatchBroadcastToSymbol 批量向特定交易对广播消息
func (bm *BroadcastManagerImpl) BatchBroadcastToSymbol(symbol string, messages []interface{}) error {
	bm.mu.RLock()
	if !bm.running {
		bm.mu.RUnlock()
		return &BroadcastError{
			Type:    "BROADCASTER_NOT_RUNNING",
			Message: "广播管理器未运行",
		}
	}
	bm.mu.RUnlock()

	// 获取订阅者
	subscribers := bm.connManager.GetSubscribers(symbol)
	if len(subscribers) == 0 {
		return &BroadcastError{
			Type:    "NO_SUBSCRIBERS",
			Message: "交易对没有订阅者",
			Symbol:  symbol,
		}
	}

	// 批量发送消息
	for _, message := range messages {
		broadcastMsg := &BroadcastMessage{
			ID:         generateMessageID(),
			Type:       "batch_symbol_broadcast",
			Symbol:     symbol,
			Data:       message,
			Timestamp:  time.Now(),
			RetryCount: 0,
			Priority:   1,
		}

		select {
		case bm.messageQueue <- broadcastMsg:
			bm.updateStats("batch_symbol_broadcast", symbol)
		default:
			return &BroadcastError{
				Type:    "QUEUE_FULL",
				Message: "消息队列已满",
			}
		}
	}

	return nil
}

// BatchBroadcastToAll 批量向所有连接广播消息
func (bm *BroadcastManagerImpl) BatchBroadcastToAll(messages []interface{}) error {
	bm.mu.RLock()
	if !bm.running {
		bm.mu.RUnlock()
		return &BroadcastError{
			Type:    "BROADCASTER_NOT_RUNNING",
			Message: "广播管理器未运行",
		}
	}
	bm.mu.RUnlock()

	// 批量发送消息
	for _, message := range messages {
		broadcastMsg := &BroadcastMessage{
			ID:         generateMessageID(),
			Type:       "batch_broadcast_all",
			Data:       message,
			Timestamp:  time.Now(),
			RetryCount: 0,
			Priority:   2,
		}

		select {
		case bm.messageQueue <- broadcastMsg:
			bm.updateStats("batch_broadcast_all", "")
		default:
			return &BroadcastError{
				Type:    "QUEUE_FULL",
				Message: "消息队列已满",
			}
		}
	}

	return nil
}

// Start 启动广播管理器
func (bm *BroadcastManagerImpl) Start(ctx context.Context) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if bm.running {
		return &BroadcastError{
			Type:    "ALREADY_RUNNING",
			Message: "广播管理器已在运行",
		}
	}

	bm.ctx, bm.cancel = context.WithCancel(ctx)
	bm.running = true

	// 启动工作协程
	for i := 0; i < bm.config.WorkerCount; i++ {
		go bm.worker(i)
	}

	// 启动批量处理协程
	go bm.batchProcessor()

	return nil
}

// Stop 停止广播管理器
func (bm *BroadcastManagerImpl) Stop(ctx context.Context) error {
	bm.mu.Lock()
	defer bm.mu.Unlock()

	if !bm.running {
		return &BroadcastError{
			Type:    "NOT_RUNNING",
			Message: "广播管理器未运行",
		}
	}

	bm.running = false
	bm.cancel()

	// 等待所有消息处理完成
	time.Sleep(100 * time.Millisecond)

	return nil
}

// IsRunning 检查是否运行
func (bm *BroadcastManagerImpl) IsRunning() bool {
	bm.mu.RLock()
	defer bm.mu.RUnlock()
	return bm.running
}

// GetBroadcastStats 获取广播统计
func (bm *BroadcastManagerImpl) GetBroadcastStats() *BroadcastStats {
	bm.statsMu.RLock()
	defer bm.statsMu.RUnlock()

	// 返回副本
	stats := *bm.stats
	stats.SymbolStats = make(map[string]int64)
	for k, v := range bm.stats.SymbolStats {
		stats.SymbolStats[k] = v
	}

	return &stats
}

// GetQueueSize 获取队列大小
func (bm *BroadcastManagerImpl) GetQueueSize() int {
	return len(bm.messageQueue) + len(bm.priorityQueue)
}

// GetFailedCount 获取失败次数
func (bm *BroadcastManagerImpl) GetFailedCount() int64 {
	bm.statsMu.RLock()
	defer bm.statsMu.RUnlock()
	return bm.stats.FailedBroadcasts
}

// SetMaxQueueSize 设置最大队列大小
func (bm *BroadcastManagerImpl) SetMaxQueueSize(size int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.config.MaxQueueSize = size
}

// SetRetryAttempts 设置重试次数
func (bm *BroadcastManagerImpl) SetRetryAttempts(attempts int) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.config.RetryAttempts = attempts
}

// SetRetryDelay 设置重试延迟
func (bm *BroadcastManagerImpl) SetRetryDelay(delay time.Duration) {
	bm.mu.Lock()
	defer bm.mu.Unlock()
	bm.config.RetryDelay = delay
}

// 辅助方法

func (bm *BroadcastManagerImpl) worker(id int) {
	for {
		select {
		case <-bm.ctx.Done():
			return
		case msg := <-bm.priorityQueue:
			if msg != nil {
				bm.processMessage(msg)
			}
		case msg := <-bm.messageQueue:
			if msg != nil {
				bm.processMessage(msg)
			}
		}
	}
}

func (bm *BroadcastManagerImpl) batchProcessor() {
	ticker := time.NewTicker(bm.config.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-bm.ctx.Done():
			return
		case <-ticker.C:
			// 批量处理逻辑可以在这里实现
		}
	}
}

func (bm *BroadcastManagerImpl) processMessage(msg *BroadcastMessage) {
	start := time.Now()

	switch msg.Type {
	case "symbol_broadcast", "batch_symbol_broadcast":
		bm.processSymbolBroadcast(msg)
	case "broadcast_all", "batch_broadcast_all":
		bm.processBroadcastAll(msg)
	case "unicast":
		bm.processUnicast(msg)
	}

	// 更新延迟统计
	latency := time.Since(start)
	bm.updateLatencyStats(latency)
}

func (bm *BroadcastManagerImpl) processSymbolBroadcast(msg *BroadcastMessage) {
	subscribers := bm.connManager.GetSubscribers(msg.Symbol)

	for _, connID := range subscribers {
		if bm.connManager.IsConnectionActive(connID) {
			err := bm.connManager.SendToConnection(connID, msg.Data)
			if err != nil {
				bm.handleSendError(msg, connID, err)
			} else {
				bm.updateSuccessStats()
			}
		}
	}
}

func (bm *BroadcastManagerImpl) processBroadcastAll(msg *BroadcastMessage) {
	connections := bm.connManager.GetConnections()

	for _, conn := range connections {
		if conn.IsActive {
			err := bm.connManager.SendToConnection(conn.ID, msg.Data)
			if err != nil {
				bm.handleSendError(msg, conn.ID, err)
			} else {
				bm.updateSuccessStats()
			}
		}
	}
}

func (bm *BroadcastManagerImpl) processUnicast(msg *BroadcastMessage) {
	err := bm.connManager.SendToConnection(msg.ConnID, msg.Data)
	if err != nil {
		bm.handleSendError(msg, msg.ConnID, err)
	} else {
		bm.updateSuccessStats()
	}
}

func (bm *BroadcastManagerImpl) handleSendError(msg *BroadcastMessage, connID string, err error) {
	// 重试逻辑
	if msg.RetryCount < bm.config.RetryAttempts {
		msg.RetryCount++
		time.Sleep(bm.config.RetryDelay)

		// 重新发送
		select {
		case bm.priorityQueue <- msg:
			return
		default:
			// 队列满，记录失败
		}
	}

	// 重试次数用完，记录失败
	bm.updateFailureStats()
}

func (bm *BroadcastManagerImpl) updateStats(msgType, symbol string) {
	bm.statsMu.Lock()
	defer bm.statsMu.Unlock()

	bm.stats.TotalBroadcasts++
	bm.stats.LastBroadcast = time.Now()

	if symbol != "" {
		bm.stats.SymbolStats[symbol]++
	}
}

func (bm *BroadcastManagerImpl) updateSuccessStats() {
	bm.statsMu.Lock()
	defer bm.statsMu.Unlock()
	bm.stats.SuccessfulBroadcasts++
}

func (bm *BroadcastManagerImpl) updateFailureStats() {
	bm.statsMu.Lock()
	defer bm.statsMu.Unlock()
	bm.stats.FailedBroadcasts++
}

func (bm *BroadcastManagerImpl) updateLatencyStats(latency time.Duration) {
	bm.statsMu.Lock()
	defer bm.statsMu.Unlock()

	// 简单的移动平均
	if bm.stats.AverageLatency == 0 {
		bm.stats.AverageLatency = latency
	} else {
		bm.stats.AverageLatency = (bm.stats.AverageLatency + latency) / 2
	}
}

// BroadcastError 广播错误
type BroadcastError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
	Symbol  string `json:"symbol,omitempty"`
	ConnID  string `json:"conn_id,omitempty"`
}

func (e *BroadcastError) Error() string {
	return e.Message
}

// 工具函数
func generateMessageID() string {
	return fmt.Sprintf("msg_%d", time.Now().UnixNano())
}
