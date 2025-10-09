package models

import (
	"time"

	"github.com/lib/pq"
)

// Symbol 交易对信息模型
type Symbol struct {
	ID                  int64          `gorm:"primaryKey;autoIncrement" json:"id"`
	Symbol              string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"symbol"`
	BaseCoin            string         `gorm:"type:varchar(20);not null" json:"base_coin"`
	QuoteCoin           string         `gorm:"type:varchar(20);not null" json:"quote_coin"`
	BuyLimitPriceRatio  *float64       `gorm:"type:decimal(10,4)" json:"buy_limit_price_ratio,omitempty"`
	SellLimitPriceRatio *float64       `gorm:"type:decimal(10,4)" json:"sell_limit_price_ratio,omitempty"`
	FeeRateUpRatio      *float64       `gorm:"type:decimal(10,4)" json:"fee_rate_up_ratio,omitempty"`
	MakerFeeRate        *float64       `gorm:"type:decimal(10,6)" json:"maker_fee_rate,omitempty"`
	TakerFeeRate        *float64       `gorm:"type:decimal(10,6)" json:"taker_fee_rate,omitempty"`
	OpenCostUpRatio     *float64       `gorm:"type:decimal(10,4)" json:"open_cost_up_ratio,omitempty"`
	SupportMarginCoins  pq.StringArray `gorm:"type:text[]" json:"support_margin_coins,omitempty"`
	MinTradeNum         *float64       `gorm:"type:decimal(20,8)" json:"min_trade_num,omitempty"`
	PriceEndStep        *float64       `gorm:"type:decimal(20,8)" json:"price_end_step,omitempty"`
	VolumePlace         *int           `json:"volume_place,omitempty"`
	PricePlace          *int           `json:"price_place,omitempty"`
	SizeMultiplier      *float64       `gorm:"type:decimal(20,8)" json:"size_multiplier,omitempty"`
	SymbolType          string         `gorm:"type:varchar(20)" json:"symbol_type,omitempty"`
	MinTradeUSDT        *float64       `gorm:"type:decimal(20,2)" json:"min_trade_usdt,omitempty"`
	MaxSymbolOrderNum   *int           `json:"max_symbol_order_num,omitempty"`
	MaxProductOrderNum  *int           `json:"max_product_order_num,omitempty"`
	MaxPositionNum      *float64       `gorm:"type:decimal(20,8)" json:"max_position_num,omitempty"`
	SymbolStatus        string         `gorm:"type:varchar(20)" json:"symbol_status,omitempty"`
	OffTime             *int64         `json:"off_time,omitempty"`
	LimitOpenTime       *int64         `json:"limit_open_time,omitempty"`
	DeliveryTime        *int64         `json:"delivery_time,omitempty"`
	DeliveryStartTime   *int64         `json:"delivery_start_time,omitempty"`
	DeliveryPeriod      string         `gorm:"type:varchar(20)" json:"delivery_period,omitempty"`
	LaunchTime          *int64         `json:"launch_time,omitempty"`
	FundInterval        *int           `json:"fund_interval,omitempty"`
	MinLever            *float64       `gorm:"type:decimal(10,2)" json:"min_lever,omitempty"`
	MaxLever            *float64       `gorm:"type:decimal(10,2)" json:"max_lever,omitempty"`
	PosLimit            *float64       `gorm:"type:decimal(10,4)" json:"pos_limit,omitempty"`
	MaintainTime        *int64         `json:"maintain_time,omitempty"`
	MaxMarketOrderQty   *float64       `gorm:"type:decimal(20,8)" json:"max_market_order_qty,omitempty"`
	MaxOrderQty         *float64       `gorm:"type:decimal(20,8)" json:"max_order_qty,omitempty"`
	IsActive            bool           `gorm:"default:true;index:idx_symbols_is_active,where:is_active = true" json:"is_active"`
	CreatedAt           time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"created_at"`
	UpdatedAt           time.Time      `gorm:"type:timestamptz;default:CURRENT_TIMESTAMP" json:"updated_at"`
}

// TableName 指定表名
func (Symbol) TableName() string {
	return "symbols"
}
