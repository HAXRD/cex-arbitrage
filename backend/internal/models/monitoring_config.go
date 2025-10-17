package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// MonitoringConfigFilters 监控配置过滤器
type MonitoringConfigFilters struct {
	TimeWindows     []string `json:"time_windows"`         // 时间窗口 ["1m", "5m", "15m"]
	ChangeThreshold float64  `json:"change_threshold"`     // 变化阈值
	VolumeThreshold float64  `json:"volume_threshold"`     // 成交量阈值
	Symbols         []string `json:"symbols"`              // 指定交易对
	MinPrice        *float64 `json:"min_price,omitempty"`  // 最小价格
	MaxPrice        *float64 `json:"max_price,omitempty"`  // 最大价格
	MinVolume       *float64 `json:"min_volume,omitempty"` // 最小成交量
	MaxVolume       *float64 `json:"max_volume,omitempty"` // 最大成交量
}

// Value 实现 driver.Valuer 接口
func (f MonitoringConfigFilters) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// Scan 实现 sql.Scanner 接口
func (f *MonitoringConfigFilters) Scan(value interface{}) error {
	if value == nil {
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}

	return json.Unmarshal(bytes, f)
}

// MonitoringConfig 监控配置模型
type MonitoringConfig struct {
	ID          int64                   `gorm:"primaryKey;autoIncrement" json:"id"`
	Name        string                  `gorm:"type:varchar(255);uniqueIndex;not null" json:"name"`
	Description *string                 `gorm:"type:text" json:"description,omitempty"`
	Filters     MonitoringConfigFilters `gorm:"type:jsonb;not null" json:"filters"`
	IsDefault   bool                    `gorm:"index:idx_monitoring_configs_is_default" json:"is_default"`
	CreatedAt   time.Time               `json:"created_at"`
	UpdatedAt   time.Time               `json:"updated_at"`
}

// TableName 指定表名
func (MonitoringConfig) TableName() string {
	return "monitoring_configs"
}

// BeforeCreate GORM钩子：创建前检查默认配置唯一性
func (mc *MonitoringConfig) BeforeCreate(tx *gorm.DB) error {
	if mc.IsDefault {
		// 这里需要在DAO层处理唯一性检查
		// 因为GORM钩子中无法直接查询数据库
	}
	return nil
}

// BeforeUpdate GORM钩子：更新前检查默认配置唯一性
func (mc *MonitoringConfig) BeforeUpdate(tx *gorm.DB) error {
	if mc.IsDefault {
		// 这里需要在DAO层处理唯一性检查
	}
	return nil
}

// IsValid 验证配置是否有效
func (mc *MonitoringConfig) IsValid() error {
	if mc.Name == "" {
		return &ValidationError{Field: "name", Message: "配置名称不能为空"}
	}

	if len(mc.Filters.TimeWindows) == 0 {
		return &ValidationError{Field: "filters.time_windows", Message: "时间窗口不能为空"}
	}

	if mc.Filters.ChangeThreshold <= 0 {
		return &ValidationError{Field: "filters.change_threshold", Message: "变化阈值必须大于0"}
	}

	if mc.Filters.VolumeThreshold < 0 {
		return &ValidationError{Field: "filters.volume_threshold", Message: "成交量阈值不能为负数"}
	}

	return nil
}

// ValidationError 验证错误
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func (e *ValidationError) Error() string {
	return e.Message
}
