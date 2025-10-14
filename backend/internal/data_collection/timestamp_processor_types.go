package data_collection

import (
	"time"
)

// TimestampProcessor 时间戳处理器接口
type TimestampProcessor interface {
	// 解析时间戳
	ParseTimestamp(input interface{}) (time.Time, error)

	// 格式化时间戳
	FormatTimestamp(t time.Time, format string) (string, error)

	// 时区转换
	ConvertTimezone(t time.Time, fromTZ, toTZ string) (time.Time, error)

	// 时间戳对齐
	AlignTimestamp(t time.Time, interval time.Duration) time.Time

	// 时间戳验证
	ValidateTimestamp(t time.Time, rules *TimestampRules) error

	// 获取时区信息
	GetTimezoneInfo(tz string) (*TimezoneInfo, error)

	// 批量处理时间戳
	ProcessBatchTimestamps(inputs []interface{}) ([]time.Time, []error)
}

// TimestampRules 时间戳处理规则
type TimestampRules struct {
	// 基础规则
	AllowedFormats  []string      `json:"allowed_formats" yaml:"allowed_formats"`     // 允许的时间格式
	DefaultTimezone string        `json:"default_timezone" yaml:"default_timezone"`   // 默认时区
	MaxFutureOffset time.Duration `json:"max_future_offset" yaml:"max_future_offset"` // 最大未来偏移
	MaxPastOffset   time.Duration `json:"max_past_offset" yaml:"max_past_offset"`     // 最大过去偏移

	// 精度规则
	Precision         time.Duration `json:"precision" yaml:"precision"`                   // 时间精度
	AlignmentInterval time.Duration `json:"alignment_interval" yaml:"alignment_interval"` // 对齐间隔

	// 验证规则
	StrictValidation   bool `json:"strict_validation" yaml:"strict_validation"`     // 严格验证
	AllowLeapSeconds   bool `json:"allow_leap_seconds" yaml:"allow_leap_seconds"`   // 允许闰秒
	TimezoneValidation bool `json:"timezone_validation" yaml:"timezone_validation"` // 时区验证
}

// TimezoneInfo 时区信息
type TimezoneInfo struct {
	Name         string         `json:"name"`         // 时区名称 (e.g., "UTC", "Asia/Shanghai")
	Offset       int            `json:"offset"`       // UTC偏移秒数
	OffsetHours  float64        `json:"offset_hours"` // UTC偏移小时数
	IsDST        bool           `json:"is_dst"`       // 是否夏令时
	Abbreviation string         `json:"abbreviation"` // 时区缩写 (e.g., "CST", "UTC")
	Location     *time.Location `json:"-"`            // Go time.Location对象
}

// TimestampFormat 时间戳格式
type TimestampFormat struct {
	Format    string `json:"format"`     // 格式字符串
	Example   string `json:"example"`    // 示例
	IsDefault bool   `json:"is_default"` // 是否默认格式
	Precision string `json:"precision"`  // 精度描述
}

// TimestampResult 时间戳处理结果
type TimestampResult struct {
	Original  interface{}            `json:"original"`  // 原始输入
	Parsed    time.Time              `json:"parsed"`    // 解析后的时间
	Formatted string                 `json:"formatted"` // 格式化后的字符串
	Timezone  string                 `json:"timezone"`  // 时区
	IsValid   bool                   `json:"is_valid"`  // 是否有效
	Errors    []string               `json:"errors"`    // 错误信息
	Warnings  []string               `json:"warnings"`  // 警告信息
	Metadata  map[string]interface{} `json:"metadata"`  // 元数据
}

// TimestampStats 时间戳处理统计
type TimestampStats struct {
	TotalProcessed       int64            `json:"total_processed"`
	SuccessCount         int64            `json:"success_count"`
	ErrorCount           int64            `json:"error_count"`
	WarningCount         int64            `json:"warning_count"`
	AverageLatency       time.Duration    `json:"average_latency"`
	FormatDistribution   map[string]int64 `json:"format_distribution"`
	TimezoneDistribution map[string]int64 `json:"timezone_distribution"`
	LastUpdated          time.Time        `json:"last_updated"`
}

// DefaultTimestampRules 返回默认的时间戳处理规则
func DefaultTimestampRules() *TimestampRules {
	return &TimestampRules{
		AllowedFormats: []string{
			time.RFC3339,
			time.RFC3339Nano,
			"2006-01-02T15:04:05.000Z",
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05Z",
			"2006-01-02T15:04:05.000000Z",
		},
		DefaultTimezone:    "UTC",
		MaxFutureOffset:    1 * time.Hour,
		MaxPastOffset:      365 * 24 * time.Hour, // 1年
		Precision:          time.Millisecond,
		AlignmentInterval:  time.Second,
		StrictValidation:   false,
		AllowLeapSeconds:   true,
		TimezoneValidation: true,
	}
}

// DefaultTimestampFormats 返回默认的时间戳格式
func DefaultTimestampFormats() []TimestampFormat {
	return []TimestampFormat{
		{
			Format:    time.RFC3339,
			Example:   "2006-01-02T15:04:05Z07:00",
			IsDefault: true,
			Precision: "second",
		},
		{
			Format:    time.RFC3339Nano,
			Example:   "2006-01-02T15:04:05.999999999Z07:00",
			IsDefault: false,
			Precision: "nanosecond",
		},
		{
			Format:    "2006-01-02T15:04:05.000Z",
			Example:   "2006-01-02T15:04:05.000Z",
			IsDefault: false,
			Precision: "millisecond",
		},
		{
			Format:    "2006-01-02 15:04:05",
			Example:   "2006-01-02 15:04:05",
			IsDefault: false,
			Precision: "second",
		},
	}
}

// CommonTimezones 常用时区列表
var CommonTimezones = map[string]string{
	"UTC":  "UTC",
	"GMT":  "GMT",
	"EST":  "America/New_York",
	"PST":  "America/Los_Angeles",
	"CST":  "Asia/Shanghai",
	"JST":  "Asia/Tokyo",
	"KST":  "Asia/Seoul",
	"IST":  "Asia/Kolkata",
	"MSK":  "Europe/Moscow",
	"BST":  "Europe/London",
	"CET":  "Europe/Paris",
	"AEST": "Australia/Sydney",
}

// UnixTimestampFormats Unix时间戳格式
var UnixTimestampFormats = map[string]string{
	"seconds":      "秒级时间戳",
	"milliseconds": "毫秒级时间戳",
	"microseconds": "微秒级时间戳",
	"nanoseconds":  "纳秒级时间戳",
}
