# API Specification

This is the API specification for the spec detailed in @.agent-os/specs/2025-10-07-bitget-api-integration/spec.md

## BitGet API 集成说明

本文档详细说明需要集成的 BitGet API 端点、参数和响应格式。

---

## REST API 端点

### 基础信息

- **Base URL**: `https://api.bitget.com`
- **请求方式**: GET
- **认证**: 无需认证（公开接口）
- **速率限制**: 
  - 公共接口：20次/秒（IP限制）

---

### 1. 获取合约列表

**端点**: `GET /api/v2/mix/market/contracts`

**用途**: 获取所有 USDT 永续合约的交易对列表

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| productType | string | 是 | 产品类型，固定值 `USDT-FUTURES` (USDT永续合约) |

**请求示例**:
```
GET https://api.bitget.com/api/v2/mix/market/contracts?productType=USDT-FUTURES
```

**响应示例**:
```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1678886400000,
  "data": [
    {
      "symbol": "BTCUSDT",
      "baseCoin": "BTC",
      "quoteCoin": "USDT",
      "buyLimitPriceRatio": "0.05",
      "sellLimitPriceRatio": "0.05",
      "feeRateUpRatio": "0.005",
      "makerFeeRate": "0.0002",
      "takerFeeRate": "0.0006",
      "openCostUpRatio": "0.01",
      "supportMarginCoins": ["USDT"],
      "minTradeNum": "0.001",
      "priceEndStep": "1",
      "volumePlace": "3",
      "pricePlace": "1",
      "sizeMultiplier": "0.001",
      "symbolType": "perpetual",
      "symbolStatus": "normal",
      "offTime": "-1",
      "limitOpenTime": "-1",
      "deliveryTime": "",
      "deliveryPeriod": ""
    }
  ]
}
```

**响应字段说明**:
- `symbol`: 交易对名称（如 `BTCUSDT`）
- `baseCoin`: 基础币种
- `quoteCoin`: 计价币种（USDT）
- `minTradeNum`: 最小交易数量
- `pricePlace`: 价格精度（小数位数）
- `volumePlace`: 数量精度（小数位数）
- `priceEndStep`: 价格最小变动单位
- `symbolStatus`: 状态（normal/offline）
- `symbolType`: 合约类型（perpetual=永续）

---

### 2. 获取K线数据

**端点**: `GET /api/v2/mix/market/candles`

**用途**: 获取指定交易对的历史K线数据

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| symbol | string | 是 | 交易对，如 `BTCUSDT` |
| productType | string | 是 | 产品类型：`USDT-FUTURES` |
| granularity | string | 是 | K线周期：1m, 5m, 15m, 30m, 1H, 4H, 6H, 12H, 1D, 1W |
| startTime | long | 否 | 开始时间（毫秒时间戳） |
| endTime | long | 否 | 结束时间（毫秒时间戳） |
| limit | string | 否 | 返回条数，默认100，最大1000 |

**请求示例**:
```
GET https://api.bitget.com/api/v2/mix/market/candles?symbol=BTCUSDT&productType=USDT-FUTURES&granularity=1m&limit=100
```

**响应示例**:
```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1678886400000,
  "data": [
    [
      "1678886400000",  // 时间戳
      "28000.5",        // 开盘价
      "28100.0",        // 最高价
      "27950.0",        // 最低价
      "28050.5",        // 收盘价
      "125.456",        // 基础币种成交量（BTC）
      "3521456.78",     // 计价币种成交量（USDT）
      "3521456.78"      // 计价币种成交额（USDT，与上一字段相同）
    ]
  ]
}
```

**响应数组字段顺序**:
1. 时间戳（毫秒）
2. 开盘价
3. 最高价
4. 最低价
5. 收盘价
6. 基础币种成交量（BTC 数量）
7. 计价币种成交量（USDT 数量）
8. 计价币种成交额（USDT 金额）

**注意**：
- 返回的数据按时间倒序排列（最新的在前）
- limit 最大值为 200（v2 API 限制）
- 时间范围不能超过指定的最大值（根据粒度不同而不同）

**K线周期说明**:
- `1m`: 1分钟
- `5m`: 5分钟
- `15m`: 15分钟
- `30m`: 30分钟
- `1H`: 1小时
- `4H`: 4小时
- `1D`: 1天

---

### 3. 获取单个合约行情

**端点**: `GET /api/v2/mix/market/ticker`

**用途**: 获取单个交易对的最新行情信息

**请求参数**:
| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| symbol | string | 是 | 交易对，如 `BTCUSDT` |
| productType | string | 是 | 产品类型：`USDT-FUTURES` |

**请求示例**:
```
GET https://api.bitget.com/api/v2/mix/market/ticker?symbol=BTCUSDT&productType=USDT-FUTURES
```

**响应示例**:
```json
{
  "code": "00000",
  "msg": "success",
  "requestTime": 1678886400000,
  "data": {
    "symbol": "BTCUSDT",
    "lastPr": "28050.5",
    "askPr": "28051.0",
    "bidPr": "28050.0",
    "high24h": "28500.0",
    "low24h": "27500.0",
    "priceChangePercent": "0.0125",
    "baseVolume": "15234.567",
    "quoteVolume": "428765432.10",
    "usdtVolume": "428765432.10",
    "ts": "1678886400000",
    "openUtc": "27800.0",
    "changeUtc24h": "0.009",
    "indexPrice": "28050.0",
    "fundingRate": "0.0001",
    "holdingAmount": "12345.678"
  }
}
```

**响应字段说明**:
- `symbol`: 交易对
- `lastPr`: 最新成交价
- `askPr`: 卖一价
- `bidPr`: 买一价
- `high24h`: 24小时最高价
- `low24h`: 24小时最低价
- `priceChangePercent`: 24小时涨跌幅（小数形式）
- `baseVolume`: 24小时基础币种成交量
- `quoteVolume`: 24小时计价币种成交量（USDT）
- `usdtVolume`: 24小时 USDT 成交量
- `ts`: 时间戳（毫秒）
- `indexPrice`: 指数价格
- `fundingRate`: 资金费率
- `holdingAmount`: 持仓量

---

## WebSocket API

### 基础信息

- **WebSocket URL**: `wss://ws.bitget.com/v2/ws/public`
- **协议**: WebSocket
- **认证**: 无需认证（公开频道）
- **连接限制**: 
  - 单个连接最多订阅 100 个频道
  - 每个 IP 最多建立 20 个连接
  - 消息推送频率：100ms/次

---

### 1. 建立连接

**连接 URL**: `wss://ws.bitget.com/v2/ws/public`

**连接成功后无响应消息**，客户端可直接发送订阅请求。

---

### 2. 订阅 Ticker 频道

**订阅消息格式**:
```json
{
  "op": "subscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "ticker",
      "instId": "BTCUSDT"
    },
    {
      "instType": "USDT-FUTURES",
      "channel": "ticker",
      "instId": "ETHUSDT"
    }
  ]
}
```

**参数说明**:
- `op`: 操作类型，固定值 `subscribe`
- `args`: 订阅参数数组
  - `instType`: 产品类型，固定值 `USDT-FUTURES`（USDT 永续合约）
  - `channel`: 频道类型，固定值 `ticker`（行情数据）
  - `instId`: 交易对名称，如 `BTCUSDT`

**订阅成功响应**:
```json
{
  "event": "subscribe",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "ticker",
    "instId": "BTCUSDT"
  },
  "code": "0",
  "msg": "Success"
}
```

**Ticker 数据推送**:
```json
{
  "action": "snapshot",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "ticker",
    "instId": "BTCUSDT"
  },
  "data": [
    {
      "instId": "BTCUSDT",
      "lastPr": "28050.5",
      "open24h": "27800.0",
      "high24h": "28500.0",
      "low24h": "27500.0",
      "bidPr": "28050.0",
      "askPr": "28051.0",
      "baseVolume": "15234.567",
      "quoteVolume": "428765432.10",
      "ts": "1678886400123",
      "openUtc": "27800.0",
      "chgUTC": "0.009",
      "change24h": "0.0125",
      "bidSz": "1.234",
      "askSz": "0.987",
      "indexPrice": "28050.0",
      "markPrice": "28050.5",
      "fundingRate": "0.0001",
      "nextFundingTime": "1678892400000",
      "holdingAmount": "12345.678"
    }
  ],
  "ts": 1678886400123
}
```

**Ticker 字段说明**:
- `instId`: 交易对
- `lastPr`: 最新成交价
- `open24h`: 24h 开盘价
- `high24h`: 24h 最高价
- `low24h`: 24h 最低价
- `bidPr`: 买一价
- `askPr`: 卖一价
- `baseVolume`: 24h 基础币种成交量
- `quoteVolume`: 24h 计价币种成交量（USDT）
- `change24h`: 24h 涨跌幅（小数）
- `openUtc`: UTC 0点开盘价
- `chgUTC`: UTC 0点至今涨跌幅
- `indexPrice`: 指数价格
- `markPrice`: 标记价格
- `fundingRate`: 当前资金费率
- `nextFundingTime`: 下次资金费时间
- `holdingAmount`: 持仓量
- `ts`: 时间戳（毫秒）

---

### 3. 取消订阅

**取消订阅消息格式**:
```json
{
  "op": "unsubscribe",
  "args": [
    {
      "instType": "USDT-FUTURES",
      "channel": "ticker",
      "instId": "BTCUSDT"
    }
  ]
}
```

**取消订阅成功响应**:
```json
{
  "event": "unsubscribe",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "ticker",
    "instId": "BTCUSDT"
  },
  "code": "0",
  "msg": "Success"
}
```

---

### 4. 心跳机制

**客户端发送 Ping**:
```
ping
```

**服务器响应 Pong**:
```
pong
```

**心跳要求**:
- 客户端每 30 秒发送一次 `ping`（字符串，非 JSON）
- 服务器会立即响应 `pong`（字符串，非 JSON）
- 如果 30 秒内未发送心跳，服务器会主动断开连接
- 建议客户端设置 30 秒定时器发送心跳

---

### 5. 错误处理

**订阅失败响应**:
```json
{
  "event": "error",
  "code": "30001",
  "msg": "Invalid channel",
  "arg": {
    "instType": "USDT-FUTURES",
    "channel": "ticker",
    "instId": "INVALID"
  }
}
```

**常见错误码**:
- `30001`: 参数错误
- `30002`: 频道不存在
- `30003`: 交易对不存在
- `30004`: 订阅数量超限（>100）
- `30005`: 参数类型错误
- `30006`: 连接数超限

---

## 错误码说明

### REST API 错误码

| 错误码 | 说明 | 处理方式 |
|--------|------|---------|
| 00000 | 成功 | - |
| 40001 | 参数错误 | 检查请求参数 |
| 40002 | 交易对不存在 | 检查交易对名称 |
| 40003 | 请求过于频繁 | 降低请求频率 |
| 50001 | 服务器内部错误 | 重试请求 |

### WebSocket 错误码

| 错误码 | 说明 | 处理方式 |
|--------|------|---------|
| 30001 | 无效的频道 | 检查订阅参数 |
| 30002 | 无效的交易对 | 检查交易对名称 |
| 30003 | 订阅数量超限 | 减少订阅数量 |
| 30004 | 频率限制 | 降低订阅频率 |

---

## 集成注意事项

1. **API 版本**:
   - 使用 v2 版本 API（更稳定）
   - REST API: `/api/v2/mix/market/*`
   - WebSocket: `wss://ws.bitget.com/v2/ws/public`

2. **交易对名称格式**:
   - v2 API 统一使用 `BTCUSDT`（无后缀）
   - REST API 和 WebSocket 格式一致

3. **产品类型标识**:
   - REST API: `productType=USDT-FUTURES`
   - WebSocket: `instType=USDT-FUTURES`
   - 注意大小写敏感

4. **数字类型**:
   - BitGet API 返回的数字均为字符串类型
   - 客户端需要将字符串转换为 float64 进行计算
   - 避免使用 JSON 数字类型解析（可能丢失精度）

5. **时间戳**:
   - 所有时间戳均为毫秒级 Unix 时间戳
   - Go 中需要除以 1000 转换为秒级 time.Time

6. **速率限制**:
   - REST API: 公共接口 20次/秒（IP限制）
   - 建议客户端限流 10次/秒，留有余量
   - WebSocket: 单个连接最多 100 个订阅

7. **WebSocket 心跳**:
   - 使用文本格式 `ping`/`pong`（不是 JSON）
   - 必须每 30 秒内发送一次，否则连接会被断开
   - 使用独立 goroutine 处理心跳

8. **重连策略**:
   - 使用指数退避算法（1s → 2s → 4s → ... → 60s）
   - 最大延迟不超过 60 秒
   - 重连成功后需要重新订阅所有频道
   - 记录每次重连尝试的日志

9. **数据精度**:
   - 价格精度：根据 `pricePlace` 字段（小数位数）
   - 数量精度：根据 `volumePlace` 字段（小数位数）
   - 建议使用 `shopspring/decimal` 库处理高精度计算

10. **域名备用**:
    - 主域名：api.bitget.com / ws.bitget.com
    - 备用域名：aws.bitget.com / aws-ws.bitget.com
    - 遇到网络问题时可切换备用域名

11. **K线数据限制**:
    - v2 API 单次最多返回 200 条（不是 1000）
    - 数据按时间倒序排列（最新的在前）
    - 需要多次请求获取更多历史数据

12. **错误处理**:
    - 所有错误码均为字符串类型（如 `"30001"`）
    - 需要解析 code 和 msg 字段判断错误类型
    - 网络错误需要重试，业务错误返回给调用方

