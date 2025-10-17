package websocket

import (
	"context"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocketServer WebSocket服务器接口
type WebSocketServer interface {
	// 服务器生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 连接管理
	GetConnectionCount() int
	GetConnections() []*Connection

	// 消息广播
	BroadcastToSymbol(symbol string, message interface{}) error
	BroadcastToAll(message interface{}) error
	SendToConnection(connID string, message interface{}) error

	// 订阅管理
	Subscribe(connID string, symbols []string) error
	Unsubscribe(connID string, symbols []string) error
	GetSubscriptions(connID string) []string
	GetSubscribers(symbol string) []string
}

// Connection WebSocket连接
type Connection struct {
	ID            string          `json:"id"`
	Conn          *websocket.Conn `json:"-"`
	Subscriptions []string        `json:"subscriptions"`
	LastPing      time.Time       `json:"last_ping"`
	CreatedAt     time.Time       `json:"created_at"`
	IsActive      bool            `json:"is_active"`
}

// Message WebSocket消息
type Message struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol,omitempty"`
	Symbols   []string    `json:"symbols,omitempty"`
	Data      interface{} `json:"data,omitempty"`
	Timestamp int64       `json:"timestamp"`
}

// PriceUpdateMessage 价格更新消息
type PriceUpdateMessage struct {
	Symbol     string  `json:"symbol"`
	Price      float64 `json:"price"`
	ChangeRate float64 `json:"change_rate"`
	Volume     float64 `json:"volume"`
	Timestamp  int64   `json:"timestamp"`
}

// SubscribeMessage 订阅消息
type SubscribeMessage struct {
	Symbols []string `json:"symbols"`
}

// UnsubscribeMessage 取消订阅消息
type UnsubscribeMessage struct {
	Symbols []string `json:"symbols"`
}

// PingMessage 心跳消息
type PingMessage struct{}

// PongMessage 心跳响应消息
type PongMessage struct{}

// ErrorMessage 错误消息
type ErrorMessage struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	// 基础配置
	Host            string `json:"host" yaml:"host"`
	Port            int    `json:"port" yaml:"port"`
	ReadBufferSize  int    `json:"read_buffer_size" yaml:"read_buffer_size"`
	WriteBufferSize int    `json:"write_buffer_size" yaml:"write_buffer_size"`

	// 心跳配置
	PingInterval time.Duration `json:"ping_interval" yaml:"ping_interval"`
	PongWait     time.Duration `json:"pong_wait" yaml:"pong_wait"`
	WriteWait    time.Duration `json:"write_wait" yaml:"write_wait"`

	// 连接配置
	MaxConnections   int           `json:"max_connections" yaml:"max_connections"`
	HandshakeTimeout time.Duration `json:"handshake_timeout" yaml:"handshake_timeout"`

	// 消息配置
	MessageQueueSize int `json:"message_queue_size" yaml:"message_queue_size"`
	MaxMessageSize   int `json:"max_message_size" yaml:"max_message_size"`
}

// ServerStatus 服务器状态
type ServerStatus struct {
	IsRunning       bool      `json:"is_running"`
	ConnectionCount int       `json:"connection_count"`
	StartTime       time.Time `json:"start_time"`
	Uptime          string    `json:"uptime"`
	LastActivity    time.Time `json:"last_activity"`
}

// DefaultServerConfig 默认服务器配置
func DefaultServerConfig() *ServerConfig {
	return &ServerConfig{
		Host:             "localhost",
		Port:             8080,
		ReadBufferSize:   1024,
		WriteBufferSize:  1024,
		PingInterval:     30 * time.Second,
		PongWait:         5 * time.Second,
		WriteWait:        10 * time.Second,
		MaxConnections:   1000,
		HandshakeTimeout: 10 * time.Second,
		MessageQueueSize: 256,
		MaxMessageSize:   512,
	}
}
