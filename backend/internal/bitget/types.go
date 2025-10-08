package bitget

import (
	"time"
)

// Symbol 交易对信息
type Symbol struct {
	Symbol              string   `json:"symbol"`              // 交易对名称，如 BTCUSDT
	BaseCoin            string   `json:"baseCoin"`            // 基础币种，如 ETHUSDT 中，特指ETH
	QuoteCoin           string   `json:"quoteCoin"`           // 计价币种，如 ETHUSDT 中，特指USDT
	BuyLimitPriceRatio  string   `json:"buyLimitPriceRatio"`  // 买限价比例
	SellLimitPriceRatio string   `json:"sellLimitPriceRatio"` // 卖限价比例
	FeeRateUpRatio      string   `json:"feeRateUpRatio"`      // 手续费上浮比例
	MakerFeeRate        string   `json:"makerFeeRate"`        // Maker 手续费率
	TakerFeeRate        string   `json:"takerFeeRate"`        // Taker 手续费率
	OpenCostUpRatio     string   `json:"openCostUpRatio"`     // 开仓成本上浮比例
	SupportMarginCoins  []string `json:"supportMarginCoins"`  // 支持保证金币种
	MinTradeNum         string   `json:"minTradeNum"`         // 最小开单数量(基础币)
	PriceEndStep        string   `json:"priceEndStep"`        // 价格步长
	VolumePlace         string   `json:"volumePlace"`         // 数量精度
	PricePlace          string   `json:"pricePlace"`          // 价格精度
	SizeMultiplier      string   `json:"sizeMultiplier"`      // 数量乘数 下单数量要大于 minTradeNum 并且满足 sizeMulti 的倍数
	SymbolType          string   `json:"symbolType"`          // 合约类型，perpetual永续，delivery交割
	MinTradeUSDT        string   `json:"minTradeUSDT"`        // 最小交易数量（USDT）
	MaxSymbolOrderNum   string   `json:"maxSymbolOrderNum"`   // 最大持有订单数（symbol维度）
	MaxProductOrderNum  string   `json:"maxProductOrderNum"`  // 最大持有订单数（产品类型维度）
	MaxPositionNum      string   `json:"maxPositionNum"`      // 最大持仓数量
	SymbolStatus        string   `json:"symbolStatus"`        // 交易对状态，listed 上架，normal 正常/开盘，maintain 禁止交易(禁止开平仓)，limit_open 限制下单(可平仓)，restrictedAPI API限制下单，off 下架
	OffTime             string   `json:"offTime"`             // 下线时间, '-1'表示正常
	LimitOpenTime       string   `json:"limitOpenTime"`       // 可开仓时间, '-1' 表示正常; 其它值表示symbol正在/计划维护，指定时间后禁止交易
	DeliveryTime        string   `json:"deliveryTime"`        // 交割时间
	DeliveryStartTime   string   `json:"deliveryStartTime"`   // 交割开始时间
	DeliveryPeriod      string   `json:"deliveryPeriod"`      // 交割周期, this_quarter当季, next_quarter次季
	LaunchTime          string   `json:"launchTime"`          // 上线时间
	FundInterval        string   `json:"fundInterval"`        // 资金费率间隔
	MinLever            string   `json:"minLever"`            // 最小杠杆
	MaxLever            string   `json:"maxLever"`            // 最大杠杆
	PosLimit            string   `json:"posLimit"`            // 持仓限制
	MaintainTime        string   `json:"maintainTime"`        // 维护时间
	MaxMarketOrderQty   string   `json:"maxMarketOrderQty"`   // 单笔市价单最大下单数量
	MaxOrderQty         string   `json:"maxOrderQty"`         // 单笔限价单最大下单数量
}

// Kline K线数据
type Kline struct {
	Ts          string `json:"ts"`          // 时间戳（毫秒）
	Open        string `json:"open"`        // 开盘价
	High        string `json:"high"`        // 最高价
	Low         string `json:"low"`         // 最低价
	Close       string `json:"close"`       // 收盘价
	BaseVolume  string `json:"baseVolume"`  // 交易币成交量
	QuoteVolume string `json:"quoteVolume"` // 计价币成交量
}

// KlineArray K线数组格式（原始API响应格式）
type KlineArray []string

// ParseKlineArray 解析K线数组数据
func ParseKlineArray(data []string) Kline {
	if len(data) < 7 {
		return Kline{}
	}

	return Kline{
		Ts:          data[0], // 时间戳
		Open:        data[1], // 开盘价
		High:        data[2], // 最高价
		Low:         data[3], // 最低价
		Close:       data[4], // 收盘价
		BaseVolume:  data[5], // 交易币成交量
		QuoteVolume: data[6], // 计价币成交量
	}
}

// Ticker 行情数据
type Ticker struct {
	Symbol            string `json:"symbol"`            // 币对名称
	LastPr            string `json:"lastPr"`            // 最新成交价
	AskPr             string `json:"askPr"`             // 卖一价
	BidPr             string `json:"bidPr"`             // 买一价
	BidSz             string `json:"bidSz"`             // 买一量
	AskSz             string `json:"askSz"`             // 卖一量
	High24h           string `json:"high24h"`           // 24小时最高价
	Low24h            string `json:"low24h"`            // 24小时最低价
	Ts                string `json:"ts"`                // 当前数据时间戳 Unix时间戳的毫秒数格式，如 1597026383085
	Change24h         string `json:"change24h"`         // 24小时价格涨跌幅
	BaseVolume        string `json:"baseVolume"`        // 交易币交易量
	QuoteVolume       string `json:"quoteVolume"`       // 计价币交易量
	UsdtVolume        string `json:"usdtVolume"`        // USDT交易量
	OpenUtc           string `json:"openUtc"`           // 开盘价(UTC+0时区)
	ChangeUtc24h      string `json:"changeUtc24h"`      // 24小时价格涨跌幅(UTC+0时区)
	IndexPrice        string `json:"indexPrice"`        // 指数价格
	FundingRate       string `json:"fundingRate"`       // 资金费率
	HoldingAmount     string `json:"holdingAmount"`     // 当前持仓, 单位是交易币(base coin)数量
	Open24h           string `json:"open24h"`           // 开盘价 24小时，开盘时间为24小时相对比，即：现在为2号19点，那么开盘时间对应为1号19点。
	DeliveryStartTime string `json:"deliveryStartTime"` // 交割开始时间
	DeliveryTime      string `json:"deliveryTime"`      // 交割时间
	DeliveryStatus    string `json:"deliveryStatus"`    // 交割状态，delivery_config_period: 新上币对配置中，delivery_normal: 交易中，delivery_before: 交割前10分钟，禁止开仓，delivery_period: 交割中，禁止开平仓、撤单
	MarkPrice         string `json:"markPrice"`         // 标记价格
}

// KlineRequest K线请求参数
type KlineRequest struct {
	Symbol      string `json:"symbol"`              // 交易对，如 BTCUSDT
	Granularity string `json:"granularity"`         // K线周期
	StartTime   *int64 `json:"startTime,omitempty"` // 开始时间（毫秒时间戳）
	EndTime     *int64 `json:"endTime,omitempty"`   // 结束时间（毫秒时间戳）
	Limit       int    `json:"limit,omitempty"`     // 返回条数，默认100，最大200
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
	Action string   `json:"action"` // 动作：snapshot, update
	Data   []Ticker `json:"data"`   // Ticker 数据数组
}

// WebSocketSubscription WebSocket 订阅信息
type WebSocketSubscription struct {
	InstType string `json:"instType"` // 实例类型：USDT-FUTURES
	Channel  string `json:"channel"`  // 频道：ticker
	InstId   string `json:"instId"`   // 交易对（单个）
}

// WebSocketRequest WebSocket 请求
type WebSocketRequest struct {
	Op   string                  `json:"op"`   // 操作：subscribe, unsubscribe
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
