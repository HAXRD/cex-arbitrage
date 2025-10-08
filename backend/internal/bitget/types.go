package bitget

import (
	"time"
)

// Symbol 交易对信息
type Symbol struct {
	Symbol           string `json:"symbol"`           // 交易对名称，如 BTCUSDT
	BaseCoin         string `json:"baseCoin"`         // 基础币种，如 BTC
	QuoteCoin        string `json:"quoteCoin"`        // 计价币种，如 USDT
	MinTradeNum      string `json:"minTradeNum"`      // 最小交易数量
	MaxTradeNum      string `json:"maxTradeNum"`      // 最大交易数量
	TakerFeeRate     string `json:"takerFeeRate"`     // Taker 手续费率
	MakerFeeRate     string `json:"makerFeeRate"`     // Maker 手续费率
	PriceScale       string `json:"priceScale"`       // 价格精度
	QuantityScale    string `json:"quantityScale"`    // 数量精度
	Status           string `json:"status"`           // 状态：online, offline
	SymbolType       string `json:"symbolType"`      // 交易对类型
	OffTime          string `json:"offTime"`          // 下线时间
	LimitOpenTime    string `json:"limitOpenTime"`    // 限价开仓时间
	DeliveryTime     string `json:"deliveryTime"`     // 交割时间
	LaunchTime       string `json:"launchTime"`      // 上线时间
	Volume24h        string `json:"volume24h"`        // 24小时交易量
	Amount24h        string `json:"amount24h"`        // 24小时交易额
	OpenPrice        string `json:"openPrice"`        // 开盘价
	High24h          string `json:"high24h"`          // 24小时最高价
	Low24h           string `json:"low24h"`          // 24小时最低价
	LastPrice        string `json:"lastPrice"`        // 最新价
	PriceChange      string `json:"priceChange"`      // 价格变化
	PriceChangeRate  string `json:"priceChangeRate"`  // 价格变化率
	IndexPrice       string `json:"indexPrice"`       // 指数价格
	MarkPrice        string `json:"markPrice"`        // 标记价格
	FundingRate      string `json:"fundingRate"`      // 资金费率
	NextFundingTime  string `json:"nextFundingTime"`  // 下次资金费率时间
	HoldingAmount    string `json:"holdingAmount"`    // 持仓量
}

// Kline K线数据
type Kline struct {
	Open      string `json:"open"`      // 开盘价
	High      string `json:"high"`       // 最高价
	Low       string `json:"low"`        // 最低价
	Close     string `json:"close"`      // 收盘价
	BaseVolume   string `json:"baseVolume"`   // 基础币种交易量
	QuoteVolume  string `json:"quoteVolume"`  // 计价币种交易量
	UsdtVolume   string `json:"usdtVolume"`   // USDT 交易量
	Ts        string `json:"ts"`        // 时间戳（毫秒）
}

// Ticker 行情数据
type Ticker struct {
	Symbol           string `json:"symbol"`           // 交易对
	LastPrice        string `json:"lastPr"`           // 最新价
	BidPrice         string `json:"bidPr"`           // 买一价
	AskPrice         string `json:"askPr"`           // 卖一价
	BidSize          string `json:"bidSz"`           // 买一量
	AskSize          string `json:"askSz"`           // 卖一量
	OpenPrice        string `json:"openUtc"`          // 开盘价
	High24h          string `json:"high24h"`         // 24小时最高价
	Low24h           string `json:"low24h"`          // 24小时最低价
	BaseVolume24h    string `json:"baseVolume"`      // 24小时基础币种交易量
	QuoteVolume24h   string `json:"quoteVolume"`     // 24小时计价币种交易量
	UsdtVolume24h    string `json:"usdtVolume"`      // 24小时USDT交易量
	Change24h        string `json:"changeUtc24h"`    // 24小时价格变化
	ChangeRate24h    string `json:"chgUTC"`          // 24小时价格变化率
	IndexPrice       string `json:"indexPrice"`      // 指数价格
	MarkPrice        string `json:"markPrice"`        // 标记价格
	FundingRate      string `json:"fundingRate"`      // 资金费率
	NextFundingTime  string `json:"nextFundingTime"`  // 下次资金费率时间
	HoldingAmount    string `json:"holdingAmount"`    // 持仓量
	Ts               string `json:"ts"`               // 时间戳（毫秒）
}

// ContractInfo 合约信息
type ContractInfo struct {
	Symbol           string `json:"symbol"`           // 交易对
	BaseCoin         string `json:"baseCoin"`         // 基础币种
	QuoteCoin        string `json:"quoteCoin"`        // 计价币种
	MinTradeNum      string `json:"minTradeNum"`      // 最小交易数量
	MaxTradeNum      string `json:"maxTradeNum"`      // 最大交易数量
	TakerFeeRate     string `json:"takerFeeRate"`     // Taker 手续费率
	MakerFeeRate     string `json:"makerFeeRate"`     // Maker 手续费率
	PriceScale       string `json:"priceScale"`       // 价格精度
	QuantityScale    string `json:"quantityScale"`    // 数量精度
	Status           string `json:"status"`           // 状态
	SymbolType       string `json:"symbolType"`      // 交易对类型
	OffTime          string `json:"offTime"`          // 下线时间
	LimitOpenTime    string `json:"limitOpenTime"`    // 限价开仓时间
	DeliveryTime     string `json:"deliveryTime"`     // 交割时间
	LaunchTime       string `json:"launchTime"`      // 上线时间
}

// KlineRequest K线请求参数
type KlineRequest struct {
	Symbol      string `json:"symbol"`      // 交易对，如 BTCUSDT
	Granularity string `json:"granularity"`  // K线周期
	StartTime   *int64 `json:"startTime,omitempty"`   // 开始时间（毫秒时间戳）
	EndTime     *int64 `json:"endTime,omitempty"`     // 结束时间（毫秒时间戳）
	Limit       int    `json:"limit,omitempty"`       // 返回条数，默认100，最大200
}

// WebSocketMessage WebSocket 消息
type WebSocketMessage struct {
	Event string      `json:"event"` // 事件类型：subscribe, unsubscribe, error
	Code  string      `json:"code"`  // 状态码
	Msg   string      `json:"msg"`   // 消息
	Data  interface{} `json:"data"`  // 数据
}

// WebSocketTickerData WebSocket Ticker 数据
type WebSocketTickerData struct {
	Action string `json:"action"` // 动作：snapshot, update
	Data   Ticker `json:"data"`   // Ticker 数据
}

// WebSocketSubscription WebSocket 订阅信息
type WebSocketSubscription struct {
	InstType string   `json:"instType"` // 实例类型：USDT-FUTURES
	Channel  string   `json:"channel"`  // 频道：ticker
	InstId   []string `json:"instId"`    // 交易对列表
}

// WebSocketRequest WebSocket 请求
type WebSocketRequest struct {
	Op   string                 `json:"op"`   // 操作：subscribe, unsubscribe
	Args []WebSocketSubscription `json:"args"` // 订阅参数
}

// BitgetConfig BitGet 配置
type BitgetConfig struct {
	RestBaseURL          string        `mapstructure:"rest_base_url"`
	RestBackupURL        string        `mapstructure:"rest_backup_url"`
	WebSocketURL         string        `mapstructure:"ws_url"`
	WebSocketBackupURL   string        `mapstructure:"ws_backup_url"`
	Timeout              time.Duration `mapstructure:"timeout"`
	RateLimit            int           `mapstructure:"rate_limit"`
	PingInterval         time.Duration `mapstructure:"ws_ping_interval"`
	PongTimeout          time.Duration `mapstructure:"ws_pong_timeout"`
	MaxReconnectAttempts int           `mapstructure:"max_reconnect_attempts"`
	ReconnectBaseDelay   time.Duration `mapstructure:"reconnect_base_delay"`
	ReconnectMaxDelay    time.Duration `mapstructure:"reconnect_max_delay"`
}

// 回调函数类型定义
type TickerCallback func(ticker Ticker)
type ErrorCallback func(err error)
type ConnectCallback func()
type DisconnectCallback func()
