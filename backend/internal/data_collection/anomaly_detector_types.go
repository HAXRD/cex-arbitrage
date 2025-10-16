package data_collection

import (
	"time"
)

// AnomalyDetector 异常检测器接口
type AnomalyDetector interface {
	// 检测单个数据点的异常
	DetectAnomaly(data *PriceData) (*AnomalyResult, error)

	// 批量检测异常
	DetectBatchAnomalies(data []*PriceData) ([]*AnomalyResult, error)

	// 更新历史数据（用于统计学习）
	UpdateHistory(data *PriceData) error

	// 获取异常统计
	GetAnomalyStats() *AnomalyStats

	// 重置检测器状态
	Reset() error

	// 设置检测规则
	SetRules(rules *AnomalyRules) error
}

// AnomalyResult 异常检测结果
type AnomalyResult struct {
	Data        *PriceData             `json:"data"`         // 原始数据
	IsAnomaly   bool                   `json:"is_anomaly"`   // 是否异常
	AnomalyType string                 `json:"anomaly_type"` // 异常类型
	Confidence  float64                `json:"confidence"`   // 置信度 (0-1)
	Severity    string                 `json:"severity"`     // 严重程度 (low, medium, high, critical)
	Score       float64                `json:"score"`        // 异常评分 (0-100)
	Reasons     []string               `json:"reasons"`      // 异常原因
	Suggestions []string               `json:"suggestions"`  // 处理建议
	Metadata    map[string]interface{} `json:"metadata"`     // 元数据
	DetectedAt  time.Time              `json:"detected_at"`  // 检测时间
}

// AnomalyRules 异常检测规则
type AnomalyRules struct {
	// 价格异常检测
	PriceAnomaly PriceAnomalyRules `json:"price_anomaly" yaml:"price_anomaly"`

	// 时间异常检测
	TimeAnomaly TimeAnomalyRules `json:"time_anomaly" yaml:"time_anomaly"`

	// 交易量异常检测
	VolumeAnomaly VolumeAnomalyRules `json:"volume_anomaly" yaml:"volume_anomaly"`

	// 统计异常检测
	StatisticalAnomaly StatisticalAnomalyRules `json:"statistical_anomaly" yaml:"statistical_anomaly"`

	// 模式异常检测
	PatternAnomaly PatternAnomalyRules `json:"pattern_anomaly" yaml:"pattern_anomaly"`

	// 全局设置
	GlobalSettings GlobalAnomalySettings `json:"global_settings" yaml:"global_settings"`
}

// PriceAnomalyRules 价格异常检测规则
type PriceAnomalyRules struct {
	Enabled             bool    `json:"enabled" yaml:"enabled"`
	MaxPriceChange      float64 `json:"max_price_change" yaml:"max_price_change"`           // 最大价格变化率 (0.5 = 50%)
	MinPriceChange      float64 `json:"min_price_change" yaml:"min_price_change"`           // 最小价格变化率
	PriceSpikeThreshold float64 `json:"price_spike_threshold" yaml:"price_spike_threshold"` // 价格尖峰阈值
	PriceDropThreshold  float64 `json:"price_drop_threshold" yaml:"price_drop_threshold"`   // 价格下跌阈值
	OutlierThreshold    float64 `json:"outlier_threshold" yaml:"outlier_threshold"`         // 异常值阈值 (标准差倍数)
}

// TimeAnomalyRules 时间异常检测规则
type TimeAnomalyRules struct {
	Enabled                bool          `json:"enabled" yaml:"enabled"`
	MaxTimeGap             time.Duration `json:"max_time_gap" yaml:"max_time_gap"`                         // 最大时间间隔
	MinTimeGap             time.Duration `json:"min_time_gap" yaml:"min_time_gap"`                         // 最小时间间隔
	FutureTimeAllowed      time.Duration `json:"future_time_allowed" yaml:"future_time_allowed"`           // 允许的未来时间
	DuplicateTimeThreshold time.Duration `json:"duplicate_time_threshold" yaml:"duplicate_time_threshold"` // 重复时间阈值
}

// VolumeAnomalyRules 交易量异常检测规则
type VolumeAnomalyRules struct {
	Enabled              bool    `json:"enabled" yaml:"enabled"`
	MaxVolumeChange      float64 `json:"max_volume_change" yaml:"max_volume_change"`           // 最大交易量变化率
	MinVolumeChange      float64 `json:"min_volume_change" yaml:"min_volume_change"`           // 最小交易量变化率
	VolumeSpikeThreshold float64 `json:"volume_spike_threshold" yaml:"volume_spike_threshold"` // 交易量尖峰阈值
	ZeroVolumeAllowed    bool    `json:"zero_volume_allowed" yaml:"zero_volume_allowed"`       // 是否允许零交易量
}

// StatisticalAnomalyRules 统计异常检测规则
type StatisticalAnomalyRules struct {
	Enabled                    bool    `json:"enabled" yaml:"enabled"`
	ZScoreThreshold            float64 `json:"z_score_threshold" yaml:"z_score_threshold"`                       // Z分数阈值
	IQRMultiplier              float64 `json:"iqr_multiplier" yaml:"iqr_multiplier"`                             // IQR倍数
	MovingAverageWindow        int     `json:"moving_average_window" yaml:"moving_average_window"`               // 移动平均窗口
	StandardDeviationThreshold float64 `json:"standard_deviation_threshold" yaml:"standard_deviation_threshold"` // 标准差阈值
}

// PatternAnomalyRules 模式异常检测规则
type PatternAnomalyRules struct {
	Enabled                  bool    `json:"enabled" yaml:"enabled"`
	SequenceLength           int     `json:"sequence_length" yaml:"sequence_length"`                       // 序列长度
	PatternThreshold         float64 `json:"pattern_threshold" yaml:"pattern_threshold"`                   // 模式阈值
	TrendChangeThreshold     float64 `json:"trend_change_threshold" yaml:"trend_change_threshold"`         // 趋势变化阈值
	CyclicalPatternThreshold float64 `json:"cyclical_pattern_threshold" yaml:"cyclical_pattern_threshold"` // 周期性模式阈值
}

// GlobalAnomalySettings 全局异常检测设置
type GlobalAnomalySettings struct {
	Enabled           bool    `json:"enabled" yaml:"enabled"`
	MinConfidence     float64 `json:"min_confidence" yaml:"min_confidence"`         // 最小置信度
	MaxAnomalyRate    float64 `json:"max_anomaly_rate" yaml:"max_anomaly_rate"`     // 最大异常率
	HistorySize       int     `json:"history_size" yaml:"history_size"`             // 历史数据大小
	LearningRate      float64 `json:"learning_rate" yaml:"learning_rate"`           // 学习率
	AdaptiveThreshold bool    `json:"adaptive_threshold" yaml:"adaptive_threshold"` // 自适应阈值
}

// AnomalyStats 异常检测统计
type AnomalyStats struct {
	TotalProcessed       int64            `json:"total_processed"`
	AnomalyCount         int64            `json:"anomaly_count"`
	NormalCount          int64            `json:"normal_count"`
	AnomalyRate          float64          `json:"anomaly_rate"`
	TypeDistribution     map[string]int64 `json:"type_distribution"`
	SeverityDistribution map[string]int64 `json:"severity_distribution"`
	AverageConfidence    float64          `json:"average_confidence"`
	AverageScore         float64          `json:"average_score"`
	LastUpdated          time.Time        `json:"last_updated"`
}

// AnomalyType 异常类型常量
const (
	AnomalyTypePriceSpike    = "price_spike"    // 价格尖峰
	AnomalyTypePriceDrop     = "price_drop"     // 价格下跌
	AnomalyTypePriceOutlier  = "price_outlier"  // 价格异常值
	AnomalyTypeTimeGap       = "time_gap"       // 时间间隔异常
	AnomalyTypeFutureTime    = "future_time"    // 未来时间
	AnomalyTypeDuplicateTime = "duplicate_time" // 重复时间
	AnomalyTypeVolumeSpike   = "volume_spike"   // 交易量尖峰
	AnomalyTypeVolumeDrop    = "volume_drop"    // 交易量下跌
	AnomalyTypeZeroVolume    = "zero_volume"    // 零交易量
	AnomalyTypeStatistical   = "statistical"    // 统计异常
	AnomalyTypePattern       = "pattern"        // 模式异常
	AnomalyTypeTrend         = "trend"          // 趋势异常
	AnomalyTypeCyclical      = "cyclical"       // 周期性异常
)

// SeverityLevel 严重程度常量
const (
	SeverityLow      = "low"      // 低
	SeverityMedium   = "medium"   // 中
	SeverityHigh     = "high"     // 高
	SeverityCritical = "critical" // 严重
)

// DefaultAnomalyRules 返回默认的异常检测规则
func DefaultAnomalyRules() *AnomalyRules {
	return &AnomalyRules{
		PriceAnomaly: PriceAnomalyRules{
			Enabled:             true,
			MaxPriceChange:      0.5,  // 50%
			MinPriceChange:      -0.5, // -50%
			PriceSpikeThreshold: 0.2,  // 20%
			PriceDropThreshold:  -0.2, // -20%
			OutlierThreshold:    3.0,  // 3个标准差
		},
		TimeAnomaly: TimeAnomalyRules{
			Enabled:                true,
			MaxTimeGap:             5 * time.Minute,
			MinTimeGap:             1 * time.Second,
			FutureTimeAllowed:      1 * time.Minute,
			DuplicateTimeThreshold: 1 * time.Second,
		},
		VolumeAnomaly: VolumeAnomalyRules{
			Enabled:              true,
			MaxVolumeChange:      2.0,  // 200%
			MinVolumeChange:      -0.8, // -80%
			VolumeSpikeThreshold: 1.0,  // 100%
			ZeroVolumeAllowed:    false,
		},
		StatisticalAnomaly: StatisticalAnomalyRules{
			Enabled:                    true,
			ZScoreThreshold:            2.5,
			IQRMultiplier:              1.5,
			MovingAverageWindow:        20,
			StandardDeviationThreshold: 2.0,
		},
		PatternAnomaly: PatternAnomalyRules{
			Enabled:                  true,
			SequenceLength:           10,
			PatternThreshold:         0.8,
			TrendChangeThreshold:     0.5, // 提高阈值，减少误报
			CyclicalPatternThreshold: 0.7,
		},
		GlobalSettings: GlobalAnomalySettings{
			Enabled:           true,
			MinConfidence:     0.6,
			MaxAnomalyRate:    0.1, // 10%
			HistorySize:       1000,
			LearningRate:      0.1,
			AdaptiveThreshold: true,
		},
	}
}
