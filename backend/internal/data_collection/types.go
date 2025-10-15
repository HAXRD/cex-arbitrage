package data_collection

import (
	"context"
	"time"
)

// DataCollectionService 数据采集服务接口
type DataCollectionService interface {
	// 服务生命周期管理
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 状态和健康检查
	GetStatus() *ServiceStatus
	HealthCheck() *HealthStatus
	GetConfig() *ServiceConfig

	// 配置管理
	UpdateConfig(config *ServiceConfig) error

	// 监控和指标
	GetMetrics() *ServiceMetrics
}

// ServiceConfig 服务配置
type ServiceConfig struct {
	// 基础配置
	MaxConnections      int           `json:"max_connections" yaml:"max_connections"`
	ReconnectInterval   time.Duration `json:"reconnect_interval" yaml:"reconnect_interval"`
	HealthCheckInterval time.Duration `json:"health_check_interval" yaml:"health_check_interval"`

	// 数据采集配置
	Symbols            []string      `json:"symbols" yaml:"symbols"`
	CollectionInterval time.Duration `json:"collection_interval" yaml:"collection_interval"`
	BatchSize          int           `json:"batch_size" yaml:"batch_size"`

	// 重连配置
	MaxRetries   int           `json:"max_retries" yaml:"max_retries"`
	RetryBackoff time.Duration `json:"retry_backoff" yaml:"retry_backoff"`

	// 性能配置
	WorkerPoolSize    int `json:"worker_pool_size" yaml:"worker_pool_size"`
	ChannelBufferSize int `json:"channel_buffer_size" yaml:"channel_buffer_size"`

	// 外部依赖配置
	WebSocketURL   string          `json:"websocket_url" yaml:"websocket_url"`
	RedisConfig    *RedisConfig    `json:"redis_config" yaml:"redis_config"`
	DatabaseConfig *DatabaseConfig `json:"database_config" yaml:"database_config"`
}

// RedisConfig Redis配置
type RedisConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	Password string `json:"password" yaml:"password"`
	DB       int    `json:"db" yaml:"db"`
	PoolSize int    `json:"pool_size" yaml:"pool_size"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host" yaml:"host"`
	Port     int    `json:"port" yaml:"port"`
	User     string `json:"user" yaml:"user"`
	Password string `json:"password" yaml:"password"`
	DBName   string `json:"dbname" yaml:"dbname"`
	SSLMode  string `json:"sslmode" yaml:"sslmode"`
}

// ServiceStatus 服务状态
type ServiceStatus struct {
	State             string        `json:"state"`
	Uptime            time.Duration `json:"uptime"`
	ActiveConnections int           `json:"active_connections"`
	TotalCollections  int64         `json:"total_collections"`
	ErrorCount        int64         `json:"error_count"`
	LastUpdated       time.Time     `json:"last_updated"`
	StartTime         time.Time     `json:"start_time"`
	LastError         string        `json:"last_error,omitempty"`
}

// HealthStatus 健康检查状态
// 注意：HealthStatus和HealthCheck类型已在monitoring_types.go中定义

// ServiceMetrics 服务指标
type ServiceMetrics struct {
	// 基础指标
	Uptime                time.Duration `json:"uptime"`
	TotalCollections      int64         `json:"total_collections"`
	SuccessfulCollections int64         `json:"successful_collections"`
	FailedCollections     int64         `json:"failed_collections"`
	ErrorRate             float64       `json:"error_rate"`

	// 性能指标
	AverageLatency      time.Duration `json:"average_latency"`
	MaxLatency          time.Duration `json:"max_latency"`
	ThroughputPerSecond float64       `json:"throughput_per_second"`

	// 连接指标
	ActiveConnections int   `json:"active_connections"`
	TotalConnections  int64 `json:"total_connections"`
	ReconnectCount    int64 `json:"reconnect_count"`

	// 资源使用
	MemoryUsageMB   float64 `json:"memory_usage_mb"`
	CPUUsagePercent float64 `json:"cpu_usage_percent"`

	// 时间戳
	LastUpdated time.Time `json:"last_updated"`
}

// DataCollectionEvent 数据采集事件
type DataCollectionEvent struct {
	Type      string      `json:"type"`
	Symbol    string      `json:"symbol"`
	Data      interface{} `json:"data"`
	Timestamp time.Time   `json:"timestamp"`
	Error     error       `json:"error,omitempty"`
}

// PriceData 价格数据
type PriceData struct {
	Symbol    string        `json:"symbol"`
	Price     float64       `json:"price"`
	BidPrice  float64       `json:"bid_price"`
	AskPrice  float64       `json:"ask_price"`
	Volume    float64       `json:"volume"`
	Timestamp time.Time     `json:"timestamp"`
	Source    string        `json:"source"`
	Latency   time.Duration `json:"latency"`
}

// PriceChangeRate 价格变化率
type PriceChangeRate struct {
	Symbol      string    `json:"symbol"`
	WindowSize  string    `json:"window_size"`
	ChangeRate  float64   `json:"change_rate"`
	PriceBefore float64   `json:"price_before"`
	PriceAfter  float64   `json:"price_after"`
	Timestamp   time.Time `json:"timestamp"`
}

// CollectionError 采集错误
type CollectionError struct {
	Symbol     string    `json:"symbol"`
	Error      string    `json:"error"`
	Timestamp  time.Time `json:"timestamp"`
	RetryCount int       `json:"retry_count"`
}

// 服务状态常量
const (
	StateStopped  = "stopped"
	StateStarting = "starting"
	StateRunning  = "running"
	StateStopping = "stopping"
	StateError    = "error"
)

// 注意：健康状态常量已在monitoring_types.go中定义

// 事件类型常量
const (
	EventTypeStart       = "start"
	EventTypeStop        = "stop"
	EventTypeData        = "data"
	EventTypeError       = "error"
	EventTypeReconnect   = "reconnect"
	EventTypeHealthCheck = "health_check"
)

// 默认配置
func DefaultServiceConfig() *ServiceConfig {
	return &ServiceConfig{
		MaxConnections:      100,
		ReconnectInterval:   5 * time.Second,
		HealthCheckInterval: 30 * time.Second,
		CollectionInterval:  1 * time.Second,
		BatchSize:           100,
		MaxRetries:          3,
		RetryBackoff:        1 * time.Second,
		WorkerPoolSize:      10,
		ChannelBufferSize:   1000,
		WebSocketURL:        "wss://ws.bitget.com/spot/v1/stream",
		Symbols:             []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
		RedisConfig: &RedisConfig{
			Host:     "localhost",
			Port:     6379,
			Password: "",
			DB:       0,
			PoolSize: 10,
		},
		DatabaseConfig: &DatabaseConfig{
			Host:     "localhost",
			Port:     5432,
			User:     "postgres",
			Password: "password",
			DBName:   "cryptosignal",
			SSLMode:  "disable",
		},
	}
}
