package data_collection

import (
	"time"
)

// DataValidator 数据验证器接口
type DataValidator interface {
	// 验证价格数据
	ValidatePriceData(data *PriceData) *ValidationResult

	// 验证批量数据
	ValidateBatchData(data []*PriceData) []*ValidationResult

	// 设置验证规则
	SetValidationRules(rules *ValidationRules) error

	// 获取验证规则
	GetValidationRules() *ValidationRules

	// 获取验证统计
	GetValidationStats() *ValidationStats
}

// DataCleaner 数据清洗器接口
type DataCleaner interface {
	// 清洗价格数据
	CleanPriceData(data *PriceData) *CleanedPriceData

	// 清洗批量数据
	CleanBatchData(data []*PriceData) []*CleanedPriceData

	// 设置清洗规则
	SetCleaningRules(rules *CleaningRules) error

	// 获取清洗规则
	GetCleaningRules() *CleaningRules

	// 获取清洗统计
	GetCleaningStats() *CleaningStats
}

// ValidationResult 验证结果
type ValidationResult struct {
	IsValid   bool                   `json:"is_valid"`
	Errors    []ValidationError      `json:"errors"`
	Warnings  []ValidationWarning    `json:"warnings"`
	Score     float64                `json:"score"` // 数据质量评分 (0-100)
	Timestamp time.Time              `json:"timestamp"`
	Metadata  map[string]interface{} `json:"metadata"`
}

// ValidationError 验证错误
type ValidationError struct {
	Field     string `json:"field"`
	Code      string `json:"code"`
	Message   string `json:"message"`
	Severity  string `json:"severity"`  // error, warning, info
	Suggested string `json:"suggested"` // 建议的修复方法
}

// ValidationWarning 验证警告
type ValidationWarning struct {
	Field      string  `json:"field"`
	Code       string  `json:"code"`
	Message    string  `json:"message"`
	Confidence float64 `json:"confidence"` // 警告置信度 (0-1)
}

// CleanedPriceData 清洗后的价格数据
type CleanedPriceData struct {
	Original   *PriceData             `json:"original"`
	Cleaned    *PriceData             `json:"cleaned"`
	Changes    []DataChange           `json:"changes"`
	Quality    float64                `json:"quality"`    // 数据质量评分
	Confidence float64                `json:"confidence"` // 清洗置信度
	Timestamp  time.Time              `json:"timestamp"`
	Metadata   map[string]interface{} `json:"metadata"`
}

// DataChange 数据变更记录
type DataChange struct {
	Field      string      `json:"field"`
	Original   interface{} `json:"original"`
	Cleaned    interface{} `json:"cleaned"`
	Reason     string      `json:"reason"`
	Confidence float64     `json:"confidence"`
}

// ValidationRules 验证规则
type ValidationRules struct {
	// 基础验证规则
	RequiredFields []string    `json:"required_fields" yaml:"required_fields"`
	PriceRange     *PriceRange `json:"price_range" yaml:"price_range"`
	TimeRange      *TimeRange  `json:"time_range" yaml:"time_range"`

	// 数据质量规则
	MaxPriceChange float64       `json:"max_price_change" yaml:"max_price_change"` // 最大价格变化率
	MinVolume      float64       `json:"min_volume" yaml:"min_volume"`             // 最小交易量
	MaxLatency     time.Duration `json:"max_latency" yaml:"max_latency"`           // 最大延迟

	// 异常检测规则
	AnomalyThreshold float64 `json:"anomaly_threshold" yaml:"anomaly_threshold"` // 异常阈值
	OutlierThreshold float64 `json:"outlier_threshold" yaml:"outlier_threshold"` // 异常值阈值

	// 数据一致性规则
	ConsistencyCheck bool `json:"consistency_check" yaml:"consistency_check"` // 是否检查数据一致性
	CrossValidation  bool `json:"cross_validation" yaml:"cross_validation"`   // 是否进行交叉验证

	// 评分规则
	QualityWeights map[string]float64 `json:"quality_weights" yaml:"quality_weights"` // 质量评分权重
}

// CleaningRules 清洗规则
type CleaningRules struct {
	// 基础清洗规则
	RemoveOutliers  bool `json:"remove_outliers" yaml:"remove_outliers"`     // 是否移除异常值
	FillMissingData bool `json:"fill_missing_data" yaml:"fill_missing_data"` // 是否填充缺失数据
	NormalizePrices bool `json:"normalize_prices" yaml:"normalize_prices"`   // 是否标准化价格

	// 价格清洗规则
	PricePrecision int     `json:"price_precision" yaml:"price_precision"`   // 价格精度
	PriceRounding  string  `json:"price_rounding" yaml:"price_rounding"`     // 价格舍入方式
	MinPriceChange float64 `json:"min_price_change" yaml:"min_price_change"` // 最小价格变化

	// 时间清洗规则
	TimeAlignment bool          `json:"time_alignment" yaml:"time_alignment"` // 是否对齐时间戳
	TimePrecision time.Duration `json:"time_precision" yaml:"time_precision"` // 时间精度

	// 数据修复规则
	AutoFixErrors       bool   `json:"auto_fix_errors" yaml:"auto_fix_errors"`           // 是否自动修复错误
	InterpolationMethod string `json:"interpolation_method" yaml:"interpolation_method"` // 插值方法
}

// PriceRange 价格范围
type PriceRange struct {
	Min     float64                `json:"min" yaml:"min"`
	Max     float64                `json:"max" yaml:"max"`
	Symbols map[string]*PriceRange `json:"symbols" yaml:"symbols"` // 按交易对的价格范围
}

// TimeRange 时间范围
type TimeRange struct {
	Min           time.Time     `json:"min" yaml:"min"`
	Max           time.Time     `json:"max" yaml:"max"`
	AllowedFuture time.Duration `json:"allowed_future" yaml:"allowed_future"` // 允许的未来时间
}

// ValidationStats 验证统计
type ValidationStats struct {
	TotalValidated    int64            `json:"total_validated"`
	ValidCount        int64            `json:"valid_count"`
	InvalidCount      int64            `json:"invalid_count"`
	WarningCount      int64            `json:"warning_count"`
	AverageScore      float64          `json:"average_score"`
	ErrorDistribution map[string]int64 `json:"error_distribution"`
	LastUpdated       time.Time        `json:"last_updated"`
}

// CleaningStats 清洗统计
type CleaningStats struct {
	TotalCleaned       int64            `json:"total_cleaned"`
	ChangesCount       int64            `json:"changes_count"`
	AverageQuality     float64          `json:"average_quality"`
	AverageConfidence  float64          `json:"average_confidence"`
	ChangeDistribution map[string]int64 `json:"change_distribution"`
	LastUpdated        time.Time        `json:"last_updated"`
}

// DefaultValidationRules 返回默认的验证规则
func DefaultValidationRules() *ValidationRules {
	return &ValidationRules{
		RequiredFields: []string{"symbol", "price", "timestamp", "source"},
		PriceRange: &PriceRange{
			Min: 0.000001,
			Max: 1000000.0,
		},
		TimeRange: &TimeRange{
			Min:           time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
			Max:           time.Now().Add(1 * time.Hour),
			AllowedFuture: 5 * time.Minute,
		},
		MaxPriceChange:   50.0, // 50% 最大价格变化
		MinVolume:        0.0,  // 最小交易量
		MaxLatency:       10 * time.Second,
		AnomalyThreshold: 10.0, // 10% 异常阈值
		OutlierThreshold: 3.0,  // 3倍标准差异常值阈值
		ConsistencyCheck: true,
		CrossValidation:  false,
		QualityWeights: map[string]float64{
			"price_accuracy": 0.3,
			"time_accuracy":  0.2,
			"completeness":   0.2,
			"consistency":    0.3,
		},
	}
}

// DefaultCleaningRules 返回默认的清洗规则
func DefaultCleaningRules() *CleaningRules {
	return &CleaningRules{
		RemoveOutliers:      true,
		FillMissingData:     true,
		NormalizePrices:     false,
		PricePrecision:      8,
		PriceRounding:       "round",
		MinPriceChange:      0.000001,
		TimeAlignment:       false,
		TimePrecision:       time.Millisecond,
		AutoFixErrors:       true,
		InterpolationMethod: "linear",
	}
}
