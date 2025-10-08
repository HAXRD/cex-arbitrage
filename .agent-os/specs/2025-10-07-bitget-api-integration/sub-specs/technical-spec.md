# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-07-bitget-api-integration/spec.md

## Technical Requirements

### 后端架构设计

**目录结构：**
```
backend/
├── internal/
│   ├── bitget/              # BitGet API 客户端包
│   │   ├── client.go        # REST API 客户端
│   │   ├── websocket.go     # WebSocket 客户端
│   │   ├── types.go         # 数据类型定义
│   │   ├── errors.go        # 错误定义
│   │   └── constants.go     # 常量定义
│   └── ...
```

---

### REST API 客户端实现

**核心接口：**
```go
type BitgetClient interface {
    // 获取所有 USDT 永续合约列表
    GetContractSymbols(ctx context.Context) ([]Symbol, error)
    
    // 获取 K线数据
    GetKlines(ctx context.Context, req *KlineRequest) ([]Kline, error)
    
    // 获取单个合约详情
    GetContractInfo(ctx context.Context, symbol string) (*ContractInfo, error)
}
```

**数据结构：**
```go
// Symbol 交易对信息
type Symbol struct {
    Symbol          string  `json:"symbol"`           // 交易对名称，如 "BTCUSDT"
    BaseCoin        string  `json:"baseCoin"`         // 基础币种，如 "BTC"
    QuoteCoin       string  `json:"quoteCoin"`        // 计价币种，如 "USDT"
    MinTradeAmount  string  `json:"minTradeNum"`      // 最小交易数量
    PriceEndStep    string  `json:"priceEndStep"`     // 价格最小变动单位
    PriceScale      int     `json:"pricePlace"`       // 价格精度（小数位数）
    VolumeScale     int     `json:"volumePlace"`      // 数量精度（小数位数）
    SymbolType      string  `json:"symbolType"`       // 合约类型：perpetual
    Status          string  `json:"symbolStatus"`     // 状态：normal/offline
}

// Kline K线数据
type Kline struct {
    Timestamp     int64   `json:"ts"`               // 时间戳（毫秒）
    Open          string  `json:"open"`             // 开盘价
    High          string  `json:"high"`             // 最高价
    Low           string  `json:"low"`              // 最低价
    Close         string  `json:"close"`            // 收盘价
    BaseVolume    string  `json:"baseVolume"`       // 基础币种成交量
    QuoteVolume   string  `json:"quoteVolume"`      // 计价币种成交量（USDT）
}

// KlineRequest K线请求参数
type KlineRequest struct {
    Symbol      string      // 交易对，如 "BTCUSDT"
    ProductType string      // 产品类型，固定值 "USDT-FUTURES"
    Granularity string      // 周期：1m, 5m, 15m, 30m, 1H, 4H, 6H, 12H, 1D, 1W
    StartTime   int64       // 开始时间（毫秒，可选）
    EndTime     int64       // 结束时间（毫秒，可选）
    Limit       int         // 返回条数，默认100，最大200
}
```

**API 端点：**
- **合约列表**：`GET /api/v2/mix/market/contracts?productType=USDT-FUTURES`
  - productType: `USDT-FUTURES` (USDT 永续合约)
- **K线数据**：`GET /api/v2/mix/market/candles`
  - 参数：symbol, granularity, startTime, endTime, limit

**技术要点：**
1. **HTTP 客户端**：使用 `net/http` 标准库，设置合理的超时时间（10秒）
2. **速率限制**：使用 `golang.org/x/time/rate` 实现客户端限流（10次/秒，实际限制20次/秒）
3. **域名备用**：支持主域名和备用域名自动切换（api.bitget.com / aws.bitget.com）
4. **错误处理**：
   - 网络错误：自动重试 3 次，指数退避（1s, 2s, 4s）
   - API 错误：解析错误码，返回自定义错误类型
   - 超时错误：记录日志，返回 context.DeadlineExceeded
5. **JSON 解析**：使用 `encoding/json`，处理数字字符串（BitGet 返回数字为字符串）
6. **productType 参数**：所有请求添加 `productType=USDT-FUTURES` 参数
7. **日志记录**：使用 Zap 记录所有 API 请求和响应（生产环境隐藏敏感信息）

---

### WebSocket 客户端实现

**核心接口：**
```go
type WebSocketClient interface {
    // 连接 WebSocket
    Connect(ctx context.Context) error
    
    // 订阅 Ticker 数据
    SubscribeTicker(symbols []string, callback TickerCallback) error
    
    // 取消订阅
    Unsubscribe(symbols []string) error
    
    // 关闭连接
    Close() error
}

type TickerCallback func(ticker *Ticker)
```

**Ticker 数据结构：**
```go
type Ticker struct {
    Symbol          string  `json:"instId"`          // 交易对
    LastPrice       string  `json:"lastPr"`          // 最新价
    BidPrice        string  `json:"bidPr"`           // 买一价
    AskPrice        string  `json:"askPr"`           // 卖一价
    Open24h         string  `json:"open24h"`         // 24h开盘价
    High24h         string  `json:"high24h"`         // 24h最高价
    Low24h          string  `json:"low24h"`          // 24h最低价
    BaseVolume24h   string  `json:"baseVolume"`      // 24h基础币种成交量
    QuoteVolume24h  string  `json:"quoteVolume"`     // 24h计价币种成交量（USDT）
    Change24h       string  `json:"change24h"`       // 24h涨跌幅（小数）
    OpenUtc         string  `json:"openUtc"`         // UTC 0点开盘价
    ChangeUtc       string  `json:"chgUTC"`          // UTC至今涨跌幅
    IndexPrice      string  `json:"indexPrice"`      // 指数价格
    MarkPrice       string  `json:"markPrice"`       // 标记价格
    FundingRate     string  `json:"fundingRate"`     // 资金费率
    NextFundingTime string  `json:"nextFundingTime"` // 下次资金费时间
    HoldingAmount   string  `json:"holdingAmount"`   // 持仓量
    Timestamp       int64   `json:"ts"`              // 时间戳（毫秒）
}
```

**WebSocket 端点：**
- **URL**：`wss://ws.bitget.com/v2/ws/public`
- **备用 URL**：`wss://aws-ws.bitget.com/v2/ws/public`
- **订阅消息**：
  ```json
  {
    "op": "subscribe",
    "args": [
      {
        "instType": "USDT-FUTURES",
        "channel": "ticker",
        "instId": "BTCUSDT"
      }
    ]
  }
  ```

**技术要点：**

1. **连接管理**：
   - 使用 `gorilla/websocket` 库
   - 单例模式，全局维护一个 WebSocket 连接
   - 支持并发订阅（使用 sync.Map 存储订阅信息）

2. **心跳机制**：
   - 每 30 秒发送文本 `ping`（字符串，非 JSON）
   - 服务器立即响应文本 `pong`（字符串，非 JSON）
   - 30 秒内未发送心跳，服务器会主动断开连接
   - 使用独立 goroutine 发送心跳

3. **自动重连策略**：
   ```go
   // 指数退避重连
   baseDelay := 1 * time.Second
   maxDelay := 60 * time.Second
   maxRetries := 10
   
   for attempt := 0; attempt < maxRetries; attempt++ {
       delay := min(baseDelay * (1 << attempt), maxDelay)
       time.Sleep(delay)
       
       if err := reconnect(); err == nil {
           restoreSubscriptions() // 重新订阅之前的交易对
           break
       }
   }
   ```

4. **消息处理**：
   - 启动独立 goroutine 读取消息
   - 使用 channel 分发消息到各个回调函数
   - 异步处理，避免阻塞读取

5. **错误处理**：
   - 连接错误：记录日志，触发重连
   - 数据解析错误：记录日志，跳过该消息
   - 订阅失败：返回错误，不影响其他订阅

6. **并发安全**：
   - 使用 `sync.RWMutex` 保护订阅列表
   - 使用 `sync.Once` 确保连接只初始化一次
   - 回调函数在独立 goroutine 中执行

---

### 配置管理

**扩展 config.yaml：**
```yaml
bitget:
  rest_base_url: "https://api.bitget.com"
  rest_backup_url: "https://aws.bitget.com"
  ws_url: "wss://ws.bitget.com/v2/ws/public"
  ws_backup_url: "wss://aws-ws.bitget.com/v2/ws/public"
  timeout: 10s                    # HTTP 请求超时
  rate_limit: 10                  # 每秒请求限制（建议值，实际限制20）
  ws_ping_interval: 30s           # WebSocket 心跳间隔
  ws_pong_timeout: 60s            # 心跳超时时间
  max_reconnect_attempts: 10      # 最大重连次数
  reconnect_base_delay: 1s        # 重连基础延迟
  reconnect_max_delay: 60s        # 重连最大延迟
```

**配置结构体：**
```go
type BitgetConfig struct {
    RestBaseURL          string        `mapstructure:"rest_base_url"`
    RestBackupURL        string        `mapstructure:"rest_backup_url"`
    WebSocketURL         string        `mapstructure:"ws_url"`
    WebSocketBackupURL   string        `mapstructure:"ws_backup_url"`
    Timeout              time.Duration `mapstructure:"timeout"`
    RateLimit            int           `mapstructure:"rate_limit"`
    WSPingInterval       time.Duration `mapstructure:"ws_ping_interval"`
    WSPongTimeout        time.Duration `mapstructure:"ws_pong_timeout"`
    MaxReconnectAttempts int           `mapstructure:"max_reconnect_attempts"`
    ReconnectBaseDelay   time.Duration `mapstructure:"reconnect_base_delay"`
    ReconnectMaxDelay    time.Duration `mapstructure:"reconnect_max_delay"`
}
```

---

### 错误处理和日志

**自定义错误类型：**
```go
var (
    ErrRateLimitExceeded = errors.New("rate limit exceeded")
    ErrInvalidSymbol     = errors.New("invalid symbol")
    ErrConnectionClosed  = errors.New("websocket connection closed")
    ErrSubscribeFailed   = errors.New("subscribe failed")
)

type BitgetAPIError struct {
    Code    string
    Message string
}

func (e *BitgetAPIError) Error() string {
    return fmt.Sprintf("bitget api error: code=%s, msg=%s", e.Code, e.Message)
}
```

**日志记录规范：**
```go
// REST API 请求日志
logger.Info("bitget rest api request",
    zap.String("method", "GET"),
    zap.String("url", url),
    zap.Any("params", params),
)

// WebSocket 连接日志
logger.Info("bitget websocket connected",
    zap.String("url", wsURL),
)

// 重连日志
logger.Warn("bitget websocket reconnecting",
    zap.Int("attempt", attempt),
    zap.Duration("delay", delay),
)

// 错误日志
logger.Error("bitget api error",
    zap.Error(err),
    zap.String("symbol", symbol),
)
```

---

### 测试要求

**单元测试：**
- REST 客户端：模拟 HTTP 响应，测试数据解析
- WebSocket 客户端：模拟 WebSocket 服务器，测试订阅和重连
- 错误处理：测试各种错误场景

**集成测试：**
- 真实 API 调用测试（使用测试环境或限流）
- WebSocket 连接测试（真实连接，订阅少量交易对）

**性能要求：**
- REST API 响应时间：< 2秒
- WebSocket 消息延迟：< 500ms
- 支持同时订阅：100+ 交易对
- 内存占用：< 50MB

---

## External Dependencies

所有依赖均为 Go 生态标准库或成熟第三方库：

- **github.com/gorilla/websocket** `v1.5.0` - WebSocket 客户端
  - **用途**：实现 WebSocket 连接和消息处理
  - **理由**：Go 社区最流行的 WebSocket 库，稳定可靠

- **golang.org/x/time/rate** - 速率限制
  - **用途**：实现客户端请求限流
  - **理由**：Go 官方扩展库，性能优异

无需其他外部依赖，所有功能使用 Go 标准库和已有依赖即可实现。

