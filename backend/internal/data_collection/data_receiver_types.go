package data_collection

import (
	"context"
	"time"
)

// DataReceiver 数据接收器接口
type DataReceiver interface {
	// 启动和停止
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	IsRunning() bool

	// 数据接收
	ReceiveData() <-chan *PriceData
	GetReceivedCount() int64
	GetErrorCount() int64

	// 状态监控
	GetStatus() *ReceiverStatus
	HealthCheck() bool
}

// ReceiverConfig 接收器配置
type ReceiverConfig struct {
	// 基础配置
	BufferSize        int           `json:"buffer_size" yaml:"buffer_size"`               // 数据缓冲区大小
	MaxRetries        int           `json:"max_retries" yaml:"max_retries"`               // 最大重试次数
	RetryInterval     time.Duration `json:"retry_interval" yaml:"retry_interval"`         // 重试间隔
	ConnectionTimeout time.Duration `json:"connection_timeout" yaml:"connection_timeout"` // 连接超时

	// 数据源配置
	DataSources []DataSourceConfig `json:"data_sources" yaml:"data_sources"` // 数据源配置列表
	Symbols     []string           `json:"symbols" yaml:"symbols"`           // 要接收的交易对列表

	// 性能配置
	WorkerCount   int           `json:"worker_count" yaml:"worker_count"`     // 工作协程数
	BatchSize     int           `json:"batch_size" yaml:"batch_size"`         // 批处理大小
	FlushInterval time.Duration `json:"flush_interval" yaml:"flush_interval"` // 刷新间隔

	// 监控配置
	MetricsEnabled bool          `json:"metrics_enabled" yaml:"metrics_enabled"` // 是否启用指标收集
	LogLevel       string        `json:"log_level" yaml:"log_level"`             // 日志级别
	LogInterval    time.Duration `json:"log_interval" yaml:"log_interval"`       // 日志输出间隔
}

// DataSourceConfig 数据源配置
type DataSourceConfig struct {
	Name       string            `json:"name" yaml:"name"`             // 数据源名称
	Type       string            `json:"type" yaml:"type"`             // 数据源类型 (websocket, rest, file)
	URL        string            `json:"url" yaml:"url"`               // 数据源URL
	Auth       map[string]string `json:"auth" yaml:"auth"`             // 认证信息
	Headers    map[string]string `json:"headers" yaml:"headers"`       // 请求头
	Parameters map[string]string `json:"parameters" yaml:"parameters"` // 请求参数
	Enabled    bool              `json:"enabled" yaml:"enabled"`       // 是否启用
	Priority   int               `json:"priority" yaml:"priority"`     // 优先级
	Weight     float64           `json:"weight" yaml:"weight"`         // 权重
}

// ReceiverStatus 接收器状态
type ReceiverStatus struct {
	Running       bool          `json:"running"`        // 是否运行中
	StartTime     time.Time     `json:"start_time"`     // 启动时间
	Uptime        time.Duration `json:"uptime"`         // 运行时间
	ReceivedCount int64         `json:"received_count"` // 接收数据总数
	ErrorCount    int64         `json:"error_count"`    // 错误总数
	LastReceived  time.Time     `json:"last_received"`  // 最后接收时间
	ActiveSources int           `json:"active_sources"` // 活跃数据源数量
	BufferUsage   float64       `json:"buffer_usage"`   // 缓冲区使用率
	Throughput    float64       `json:"throughput"`     // 吞吐量 (数据/秒)
	LastError     string        `json:"last_error"`     // 最后错误信息
}

// MessageParser 消息解析器接口
type MessageParser interface {
	// 解析消息
	ParseMessage(data []byte) (*PriceData, error)
	ParseBatch(data []byte) ([]*PriceData, error)

	// 验证消息格式
	ValidateMessage(data []byte) bool
	GetMessageType(data []byte) string

	// 配置管理
	SetConfig(config *ParserConfig) error
	GetConfig() *ParserConfig
}

// ParserConfig 解析器配置
type ParserConfig struct {
	// 消息格式配置
	MessageFormat string            `json:"message_format" yaml:"message_format"` // 消息格式 (json, xml, csv)
	FieldMapping  map[string]string `json:"field_mapping" yaml:"field_mapping"`   // 字段映射
	TimeFormat    string            `json:"time_format" yaml:"time_format"`       // 时间格式
	TimeZone      string            `json:"time_zone" yaml:"time_zone"`           // 时区

	// 数据验证配置
	RequiredFields []string `json:"required_fields" yaml:"required_fields"` // 必需字段
	PriceFields    []string `json:"price_fields" yaml:"price_fields"`       // 价格字段
	SymbolFields   []string `json:"symbol_fields" yaml:"symbol_fields"`     // 交易对字段
	TimeFields     []string `json:"time_fields" yaml:"time_fields"`         // 时间字段

	// 数据清洗配置
	MinPrice       float64 `json:"min_price" yaml:"min_price"`             // 最小价格
	MaxPrice       float64 `json:"max_price" yaml:"max_price"`             // 最大价格
	PricePrecision int     `json:"price_precision" yaml:"price_precision"` // 价格精度

	// 错误处理配置
	SkipInvalidData bool    `json:"skip_invalid_data" yaml:"skip_invalid_data"` // 跳过无效数据
	LogErrors       bool    `json:"log_errors" yaml:"log_errors"`               // 记录错误
	ErrorThreshold  float64 `json:"error_threshold" yaml:"error_threshold"`     // 错误阈值
}

// DefaultReceiverConfig 返回默认的接收器配置
func DefaultReceiverConfig() *ReceiverConfig {
	return &ReceiverConfig{
		BufferSize:        1000,
		MaxRetries:        3,
		RetryInterval:     5 * time.Second,
		ConnectionTimeout: 30 * time.Second,
		DataSources: []DataSourceConfig{
			{
				Name:     "bitget-websocket",
				Type:     "websocket",
				URL:      "wss://ws.bitget.com/spot/v1/stream",
				Enabled:  true,
				Priority: 1,
				Weight:   1.0,
			},
		},
		Symbols:        []string{"BTCUSDT", "ETHUSDT", "ADAUSDT"},
		WorkerCount:    5,
		BatchSize:      100,
		FlushInterval:  1 * time.Second,
		MetricsEnabled: true,
		LogLevel:       "info",
		LogInterval:    30 * time.Second,
	}
}

// DefaultParserConfig 返回默认的解析器配置
func DefaultParserConfig() *ParserConfig {
	return &ParserConfig{
		MessageFormat: "json",
		FieldMapping: map[string]string{
			"symbol":    "symbol",
			"price":     "price",
			"timestamp": "timestamp",
			"source":    "source",
		},
		TimeFormat:      "2006-01-02T15:04:05.000Z",
		TimeZone:        "UTC",
		RequiredFields:  []string{"symbol", "price", "timestamp"},
		PriceFields:     []string{"price", "last_price", "close_price"},
		SymbolFields:    []string{"symbol", "pair", "instrument"},
		TimeFields:      []string{"timestamp", "time", "created_at"},
		MinPrice:        0.000001,
		MaxPrice:        1000000.0,
		PricePrecision:  8,
		SkipInvalidData: true,
		LogErrors:       true,
		ErrorThreshold:  0.1, // 10% 错误率阈值
	}
}
