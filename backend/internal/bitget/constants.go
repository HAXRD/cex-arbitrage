package bitget

// BitGet API 常量定义
const (
	// API 版本
	APIVersion = "v2"
	
	// REST API 基础 URL
	DefaultRestBaseURL = "https://api.bitget.com"
	BackupRestBaseURL  = "https://aws.bitget.com"
	
	// WebSocket URL
	DefaultWebSocketURL = "wss://ws.bitget.com/v2/ws/public"
	BackupWebSocketURL  = "wss://aws-ws.bitget.com/v2/ws/public"
	
	// 产品类型
	ProductTypeUSDTFutures = "USDT-FUTURES"
	
	// 实例类型
	InstTypeUSDTFutures = "USDT-FUTURES"
	
	// WebSocket 频道
	ChannelTicker = "ticker"
	
	// WebSocket 操作
	OpSubscribe   = "subscribe"
	OpUnsubscribe = "unsubscribe"
	
	// 默认配置
	DefaultTimeout         = 10 // 秒
	DefaultRateLimit       = 10 // 每秒请求数
	DefaultPingInterval    = 30 // 秒
	DefaultPongTimeout     = 60 // 秒
	DefaultMaxReconnects   = 10
	DefaultReconnectDelay = 1  // 秒
	DefaultMaxReconnectDelay = 60 // 秒
	
	// K线周期
	Granularity1m  = "1m"
	Granularity5m  = "5m"
	Granularity15m = "15m"
	Granularity30m = "30m"
	Granularity1H  = "1H"
	Granularity4H  = "4H"
	Granularity6H  = "6H"
	Granularity12H = "12H"
	Granularity1D  = "1D"
	Granularity1W  = "1W"
	
	// 支持的时间周期
	SupportedGranularities = "1m,5m,15m,30m,1H,4H,6H,12H,1D,1W"
	
	// 最大 K线数据条数
	MaxKlineLimit = 200
	
	// 默认 K线数据条数
	DefaultKlineLimit = 100
)
