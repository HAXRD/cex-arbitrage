package data_collection

import (
	"context"
	"time"
)

// WebSocketClient WebSocket客户端接口
type WebSocketClient interface {
	// 连接管理
	Connect(ctx context.Context) error
	Disconnect(ctx context.Context) error
	IsConnected() bool
	Reconnect(ctx context.Context) error

	// 订阅管理
	Subscribe(symbol string) error
	Unsubscribe(symbol string) error
	BatchSubscribe(symbols []string) error
	GetSubscriptions() []string

	// 消息处理
	SetMessageHandler(handler MessageHandler)
	SetErrorHandler(handler ErrorHandler)
	SendMessage(data []byte) error

	// 心跳管理
	StartHeartbeat(interval time.Duration)
	StopHeartbeat()

	// 重连管理
	SetAutoReconnect(enabled bool)
	GetReconnectCount() int64

	// 状态查询
	GetConnectionInfo() *ConnectionInfo
	GetConfig() *WebSocketConfig
	SetConfig(config *WebSocketConfig) error
}

// WebSocketConfig WebSocket配置
type WebSocketConfig struct {
	URL                  string        `json:"url" yaml:"url"`
	ReconnectInterval    time.Duration `json:"reconnect_interval" yaml:"reconnect_interval"`
	MaxReconnectAttempts int           `json:"max_reconnect_attempts" yaml:"max_reconnect_attempts"`
	HeartbeatInterval    time.Duration `json:"heartbeat_interval" yaml:"heartbeat_interval"`
	ConnectionTimeout    time.Duration `json:"connection_timeout" yaml:"connection_timeout"`
	ReadBufferSize       int           `json:"read_buffer_size" yaml:"read_buffer_size"`
	WriteBufferSize      int           `json:"write_buffer_size" yaml:"write_buffer_size"`
	EnableCompression    bool          `json:"enable_compression" yaml:"enable_compression"`
	EnablePingPong       bool          `json:"enable_ping_pong" yaml:"enable_ping_pong"`
	PingTimeout          time.Duration `json:"ping_timeout" yaml:"ping_timeout"`
	PongTimeout          time.Duration `json:"pong_timeout" yaml:"pong_timeout"`
}

// ConnectionInfo 连接信息
type ConnectionInfo struct {
	Connected      bool          `json:"connected"`
	URL            string        `json:"url"`
	ConnectedAt    time.Time     `json:"connected_at"`
	LastHeartbeat  time.Time     `json:"last_heartbeat"`
	ReconnectCount int64         `json:"reconnect_count"`
	LastError      string        `json:"last_error,omitempty"`
	Latency        time.Duration `json:"latency"`
	MessageCount   int64         `json:"message_count"`
	ErrorCount     int64         `json:"error_count"`
}

// MessageHandler 消息处理器
type MessageHandler func(data []byte) error

// ErrorHandler 错误处理器
type ErrorHandler func(err error)

// WebSocketMessage WebSocket消息
type WebSocketMessage struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol,omitempty"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Error     error       `json:"error,omitempty"`
}

// SubscriptionRequest 订阅请求
type SubscriptionRequest struct {
	Symbol    string `json:"symbol"`
	Channel   string `json:"channel"`
	Action    string `json:"action"` // subscribe, unsubscribe
	Timestamp int64  `json:"timestamp"`
}

// HeartbeatMessage 心跳消息
type HeartbeatMessage struct {
	Type      string    `json:"type"`
	Timestamp time.Time `json:"timestamp"`
	Ping      bool      `json:"ping,omitempty"`
	Pong      bool      `json:"pong,omitempty"`
}

// 消息类型常量
const (
	MessageTypeData        = "data"
	MessageTypeHeartbeat   = "heartbeat"
	MessageTypeError       = "error"
	MessageTypeInfo        = "info"
	MessageTypeSubscribe   = "subscribe"
	MessageTypeUnsubscribe = "unsubscribe"
)

// 订阅动作常量
const (
	ActionSubscribe   = "subscribe"
	ActionUnsubscribe = "unsubscribe"
)

// 默认WebSocket配置
func DefaultWebSocketConfig() *WebSocketConfig {
	return &WebSocketConfig{
		URL:                  "wss://ws.bitget.com/spot/v1/stream",
		ReconnectInterval:    5 * time.Second,
		MaxReconnectAttempts: 3,
		HeartbeatInterval:    30 * time.Second,
		ConnectionTimeout:    10 * time.Second,
		ReadBufferSize:       4096,
		WriteBufferSize:      4096,
		EnableCompression:    true,
		EnablePingPong:       true,
		PingTimeout:          5 * time.Second,
		PongTimeout:          5 * time.Second,
	}
}
