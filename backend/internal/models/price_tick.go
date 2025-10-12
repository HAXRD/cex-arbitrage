package models

import (
	"time"
)

// PriceTick 实时价格数据模型
type PriceTick struct {
	Symbol            string    `gorm:"type:varchar(50);not null;index:idx_price_ticks_symbol_timestamp,priority:1" json:"symbol"`
	Timestamp         time.Time `gorm:"not null;index:idx_price_ticks_symbol_timestamp,priority:2,sort:desc;index:idx_price_ticks_timestamp,sort:desc" json:"timestamp"`
	LastPrice         float64   `gorm:"column:last_price;type:decimal(20,8);not null" json:"last_price"`
	AskPrice          *float64  `gorm:"column:ask_price;type:decimal(20,8)" json:"ask_price,omitempty"`
	BidPrice          *float64  `gorm:"column:bid_price;type:decimal(20,8)" json:"bid_price,omitempty"`
	BidSize           *float64  `gorm:"column:bid_size;type:decimal(20,8)" json:"bid_size,omitempty"`
	AskSize           *float64  `gorm:"column:ask_size;type:decimal(20,8)" json:"ask_size,omitempty"`
	High24h           *float64  `gorm:"column:high_24h;type:decimal(20,8)" json:"high_24h,omitempty"`
	Low24h            *float64  `gorm:"column:low_24h;type:decimal(20,8)" json:"low_24h,omitempty"`
	Change24h         *float64  `gorm:"column:change_24h;type:decimal(10,4)" json:"change_24h,omitempty"`
	BaseVolume        *float64  `gorm:"column:base_volume;type:decimal(30,8)" json:"base_volume,omitempty"`
	QuoteVolume       *float64  `gorm:"column:quote_volume;type:decimal(30,8)" json:"quote_volume,omitempty"`
	UsdtVolume        *float64  `gorm:"column:usdt_volume;type:decimal(30,8)" json:"usdt_volume,omitempty"`
	OpenUtc           *float64  `gorm:"column:open_utc;type:decimal(20,8)" json:"open_utc,omitempty"`
	ChangeUtc24h      *float64  `gorm:"column:change_utc_24h;type:decimal(10,4)" json:"change_utc_24h,omitempty"`
	IndexPrice        *float64  `gorm:"column:index_price;type:decimal(20,8)" json:"index_price,omitempty"`
	FundingRate       *float64  `gorm:"column:funding_rate;type:decimal(10,6)" json:"funding_rate,omitempty"`
	HoldingAmount     *float64  `gorm:"column:holding_amount;type:decimal(30,8)" json:"holding_amount,omitempty"`
	Open24h           *float64  `gorm:"column:open_24h;type:decimal(20,8)" json:"open_24h,omitempty"`
	MarkPrice         *float64  `gorm:"column:mark_price;type:decimal(20,8)" json:"mark_price,omitempty"`
	DeliveryStartTime *int64    `json:"delivery_start_time,omitempty"`
	DeliveryTime      *int64    `json:"delivery_time,omitempty"`
	DeliveryStatus    string    `gorm:"type:varchar(30)" json:"delivery_status,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
}

// TableName 指定表名
func (PriceTick) TableName() string {
	return "price_ticks"
}
