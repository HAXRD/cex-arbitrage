package models

import (
	"time"
)

// Kline K线数据模型
type Kline struct {
	Symbol      string    `gorm:"type:varchar(50);not null;uniqueIndex:idx_klines_unique,priority:1;index:idx_klines_symbol_granularity_timestamp,priority:1" json:"symbol"`
	Timestamp   time.Time `gorm:"type:timestamptz;not null;uniqueIndex:idx_klines_unique,priority:2;index:idx_klines_symbol_granularity_timestamp,priority:3,sort:desc;index:idx_klines_timestamp,sort:desc" json:"timestamp"`
	Granularity string    `gorm:"type:varchar(10);not null;uniqueIndex:idx_klines_unique,priority:3;index:idx_klines_symbol_granularity_timestamp,priority:2;index:idx_klines_granularity" json:"granularity"`
	Open        float64   `gorm:"type:decimal(20,8);not null" json:"open"`
	High        float64   `gorm:"type:decimal(20,8);not null" json:"high"`
	Low         float64   `gorm:"type:decimal(20,8);not null" json:"low"`
	Close       float64   `gorm:"type:decimal(20,8);not null" json:"close"`
	BaseVolume  float64   `gorm:"type:decimal(30,8);not null" json:"base_volume"`
	QuoteVolume float64   `gorm:"type:decimal(30,8);not null" json:"quote_volume"`
	CreatedAt   time.Time `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"created_at"`
}

// TableName 指定表名
func (Kline) TableName() string {
	return "klines"
}
