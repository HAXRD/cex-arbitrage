package data_collection

import (
	"time"
)

// ProcessedPriceData 处理后的价格数据结构
type ProcessedPriceData struct {
	Symbol    string    `json:"symbol"`    // 交易对符号
	Price     float64   `json:"price"`     // 价格
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Source    string    `json:"source"`    // 数据源
}

// ProcessedPriceChangeRate 处理后的价格变化率
type ProcessedPriceChangeRate struct {
	Symbol     string    `json:"symbol"`      // 交易对符号
	TimeWindow string    `json:"time_window"` // 时间窗口 (1m, 5m, 15m)
	ChangeRate float64   `json:"change_rate"` // 变化率 (百分比)
	StartPrice float64   `json:"start_price"` // 起始价格
	EndPrice   float64   `json:"end_price"`   // 结束价格
	Timestamp  time.Time `json:"timestamp"`   // 计算时间戳
	IsValid    bool      `json:"is_valid"`    // 数据是否有效
	IsAnomaly  bool      `json:"is_anomaly"`  // 是否为异常数据
}

// TimeWindow 时间窗口类型
type TimeWindow string

const (
	TimeWindow1m  TimeWindow = "1m"
	TimeWindow5m  TimeWindow = "5m"
	TimeWindow15m TimeWindow = "15m"
)

// PriceProcessor 价格处理器接口
type PriceProcessor interface {
	// ProcessPrice 处理单个价格数据
	ProcessPrice(price *PriceData) error

	// ProcessBatch 批量处理价格数据
	ProcessBatch(prices []*PriceData) error

	// GetChangeRate 获取指定时间窗口的变化率
	GetChangeRate(symbol string, window TimeWindow) (*ProcessedPriceChangeRate, error)

	// GetChangeRates 获取所有时间窗口的变化率
	GetChangeRates(symbol string) (map[TimeWindow]*ProcessedPriceChangeRate, error)

	// ValidateData 验证数据有效性
	ValidateData(price *PriceData) bool

	// DetectAnomaly 检测异常数据
	DetectAnomaly(price *PriceData) bool

	// CleanData 清洗数据
	CleanData(price *PriceData) *PriceData
}

// ProcessorConfig 处理器配置
type ProcessorConfig struct {
	TimeWindows      []TimeWindow  `json:"time_windows" yaml:"time_windows"`           // 支持的时间窗口
	MaxPriceChange   float64       `json:"max_price_change" yaml:"max_price_change"`   // 最大价格变化率阈值
	AnomalyThreshold float64       `json:"anomaly_threshold" yaml:"anomaly_threshold"` // 异常检测阈值
	DataRetention    time.Duration `json:"data_retention" yaml:"data_retention"`       // 数据保留时间
	CleanupInterval  time.Duration `json:"cleanup_interval" yaml:"cleanup_interval"`   // 清理间隔
}

// DefaultProcessorConfig 返回默认的处理器配置
func DefaultProcessorConfig() *ProcessorConfig {
	return &ProcessorConfig{
		TimeWindows:      []TimeWindow{TimeWindow1m, TimeWindow5m, TimeWindow15m},
		MaxPriceChange:   50.0,           // 50% 最大变化率
		AnomalyThreshold: 10.0,           // 10% 异常检测阈值
		DataRetention:    24 * time.Hour, // 24小时数据保留
		CleanupInterval:  time.Hour,      // 1小时清理间隔
	}
}

// ProcessorStatus 处理器状态
type ProcessorStatus struct {
	Running           bool          `json:"running"`
	ProcessedCount    int64         `json:"processed_count"`
	ErrorCount        int64         `json:"error_count"`
	AnomalyCount      int64         `json:"anomaly_count"`
	LastProcessedTime time.Time     `json:"last_processed_time"`
	Uptime            time.Duration `json:"uptime"`
	MemoryUsage       int64         `json:"memory_usage"`
}
