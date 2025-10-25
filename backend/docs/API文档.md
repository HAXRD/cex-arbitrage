# CryptoSignal Hunter API 文档

## 概述

CryptoSignal Hunter API 是一个专为加密货币合约交易信号捕捉系统设计的RESTful API。该API提供实时价格数据、历史K线数据、交易对管理和监控配置等功能。

## 基础信息

- **基础URL**: `http://localhost:8080/api/v1`
- **协议**: HTTP/HTTPS
- **数据格式**: JSON
- **字符编码**: UTF-8

## 认证

当前版本暂不需要认证，所有API端点都是公开访问的。

## 通用响应格式

### 成功响应

```json
{
  "data": {},
  "message": "操作成功",
  "code": 200
}
```

### 分页响应

```json
{
  "data": [],
  "total": 100,
  "page": 1,
  "page_size": 20,
  "message": "查询成功",
  "code": 200
}
```

### 错误响应

```json
{
  "error": "错误信息",
  "code": 400,
  "message": "请求参数错误"
}
```

## API端点

### 1. 系统管理

#### 健康检查

**GET** `/health`

检查服务是否正常运行。

**响应示例:**
```json
{
  "status": "ok",
  "message": "CryptoSignal Hunter 服务运行正常"
}
```

### 2. 交易对管理

#### 获取交易对列表

**GET** `/api/v1/symbols`

获取所有可用的交易对列表，支持分页和筛选。

**查询参数:**
- `page` (integer, 可选): 页码，默认为1
- `page_size` (integer, 可选): 每页大小，默认为20
- `search` (string, 可选): 搜索关键词
- `base_asset` (string, 可选): 基础资产过滤
- `quote_asset` (string, 可选): 计价资产过滤
- `status` (string, 可选): 状态过滤

**响应示例:**
```json
{
  "data": [
    {
      "symbol": "BTCUSDT",
      "symbol_type": "SPOT",
      "status": "TRADING",
      "base_asset": "BTC",
      "quote_asset": "USDT",
      "is_active": true,
      "created_at": 1640995200,
      "updated_at": 1640995200
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

#### 获取交易对详情

**GET** `/api/v1/symbols/{symbol}`

获取指定交易对的详细信息。

**路径参数:**
- `symbol` (string, 必需): 交易对符号，如 "BTCUSDT"

**响应示例:**
```json
{
  "symbol": "BTCUSDT",
  "symbol_type": "SPOT",
  "status": "TRADING",
  "base_asset": "BTC",
  "quote_asset": "USDT",
  "is_active": true,
  "created_at": 1640995200,
  "updated_at": 1640995200
}
```

#### 搜索交易对

**GET** `/api/v1/symbols/search`

搜索交易对。

**查询参数:**
- `q` (string, 必需): 搜索关键词
- `page` (integer, 可选): 页码
- `page_size` (integer, 可选): 每页大小

### 3. 历史数据

#### 获取K线数据

**GET** `/api/v1/klines/{symbol}`

获取指定交易对的K线数据。

**路径参数:**
- `symbol` (string, 必需): 交易对符号

**查询参数:**
- `interval` (string, 必需): 时间间隔，支持: 1m, 5m, 15m, 30m, 1h, 4h, 1d, 1w, 1M
- `start_time` (string, 可选): 开始时间，支持Unix时间戳或ISO 8601格式
- `end_time` (string, 可选): 结束时间
- `page` (integer, 可选): 页码
- `page_size` (integer, 可选): 每页大小

**响应示例:**
```json
{
  "data": [
    {
      "symbol": "BTCUSDT",
      "interval": "1h",
      "open_time": 1640995200,
      "close_time": 1640998800,
      "open_price": 50000.0,
      "high_price": 51000.0,
      "low_price": 49000.0,
      "close_price": 50500.0,
      "volume": 100.0,
      "quote_volume": 5000000.0,
      "trade_count": 1000
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

#### 获取K线统计信息

**GET** `/api/v1/klines/{symbol}/statistics`

获取K线数据的统计信息。

**响应示例:**
```json
{
  "total_count": 100,
  "price_range": {
    "high": 51000.0,
    "low": 49000.0,
    "open": 50000.0,
    "close": 50500.0
  },
  "volume_stats": {
    "total_volume": 1000.0,
    "average_volume": 10.0
  }
}
```

#### 获取最新K线数据

**GET** `/api/v1/klines/{symbol}/latest`

获取最新的K线数据。

### 4. 实时价格

#### 获取实时价格

**GET** `/api/v1/prices/{symbol}`

获取指定交易对的实时价格。

**路径参数:**
- `symbol` (string, 必需): 交易对符号

**响应示例:**
```json
{
  "symbol": "BTCUSDT",
  "last_price": 50500.0,
  "bid_price": 50499.0,
  "ask_price": 50501.0,
  "base_volume": 100.0,
  "quote_volume": 5050000.0,
  "timestamp": 1640995200
}
```

#### 批量获取实时价格

**GET** `/api/v1/prices`

批量获取多个交易对的实时价格。

**查询参数:**
- `symbols` (string, 必需): 交易对符号列表，用逗号分隔

**响应示例:**
```json
{
  "data": [
    {
      "symbol": "BTCUSDT",
      "last_price": 50500.0,
      "bid_price": 50499.0,
      "ask_price": 50501.0,
      "base_volume": 100.0,
      "quote_volume": 5050000.0,
      "timestamp": 1640995200
    }
  ]
}
```

#### 获取价格历史数据

**GET** `/api/v1/prices/{symbol}/history`

获取价格历史数据。

**查询参数:**
- `start_time` (string, 可选): 开始时间
- `end_time` (string, 可选): 结束时间
- `page` (integer, 可选): 页码
- `page_size` (integer, 可选): 每页大小

#### 获取价格统计信息

**GET** `/api/v1/prices/{symbol}/statistics`

获取价格统计信息。

**响应示例:**
```json
{
  "current_price": 50500.0,
  "highest_price": 51000.0,
  "lowest_price": 49000.0,
  "average_price": 50000.0,
  "price_change": 500.0,
  "change_percent": 1.0,
  "volume": 100.0,
  "trade_count": 1
}
```

### 5. 配置管理

#### 获取监控配置列表

**GET** `/api/v1/monitoring-configs`

获取所有监控配置列表。

**查询参数:**
- `page` (integer, 可选): 页码
- `page_size` (integer, 可选): 每页大小
- `search` (string, 可选): 搜索关键词

**响应示例:**
```json
{
  "data": [
    {
      "id": 1,
      "name": "BTC监控配置",
      "symbol": "BTCUSDT",
      "interval": "1h",
      "parameters": {
        "threshold": 0.01
      },
      "is_default": true,
      "is_active": true,
      "created_at": 1640995200,
      "updated_at": 1640995200
    }
  ],
  "total": 1,
  "page": 1,
  "page_size": 20
}
```

#### 创建监控配置

**POST** `/api/v1/monitoring-configs`

创建新的监控配置。

**请求体:**
```json
{
  "name": "BTC监控配置",
  "symbol": "BTCUSDT",
  "interval": "1h",
  "parameters": {
    "threshold": 0.01
  }
}
```

**响应示例:**
```json
{
  "id": 1,
  "name": "BTC监控配置",
  "symbol": "BTCUSDT",
  "interval": "1h",
  "parameters": {
    "threshold": 0.01
  },
  "is_default": false,
  "is_active": true,
  "created_at": 1640995200,
  "updated_at": 1640995200
}
```

#### 获取监控配置详情

**GET** `/api/v1/monitoring-configs/{id}`

获取指定监控配置的详细信息。

**路径参数:**
- `id` (integer, 必需): 配置ID

#### 更新监控配置

**PUT** `/api/v1/monitoring-configs/{id}`

更新指定的监控配置。

**请求体:**
```json
{
  "name": "更新后的配置",
  "symbol": "BTCUSDT",
  "interval": "1h",
  "parameters": {
    "threshold": 0.02
  }
}
```

#### 删除监控配置

**DELETE** `/api/v1/monitoring-configs/{id}`

删除指定的监控配置。

**响应:** 204 No Content

#### 获取默认监控配置

**GET** `/api/v1/monitoring-configs/default`

获取默认的监控配置。

#### 设置默认监控配置

**POST** `/api/v1/monitoring-configs/{id}/set-default`

将指定配置设置为默认配置。

#### 搜索监控配置

**GET** `/api/v1/monitoring-configs/search`

搜索监控配置。

**查询参数:**
- `q` (string, 必需): 搜索关键词
- `page` (integer, 可选): 页码
- `page_size` (integer, 可选): 每页大小

#### 验证监控配置

**POST** `/api/v1/monitoring-configs/validate`

验证监控配置的有效性。

**请求体:**
```json
{
  "name": "验证配置",
  "symbol": "BTCUSDT",
  "interval": "1h",
  "parameters": {
    "threshold": 0.01
  }
}
```

**响应示例:**
```json
{
  "valid": true,
  "errors": []
}
```

## 错误代码

| 状态码 | 说明 |
|--------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 204 | 删除成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 500 | 服务器内部错误 |

## 限制和配额

- **请求频率**: 每分钟最多1000次请求
- **分页大小**: 最大100条记录
- **时间范围**: 历史数据查询最多支持30天
- **并发连接**: 最多100个并发连接

## 示例代码

### JavaScript (Fetch API)

```javascript
// 获取交易对列表
fetch('/api/v1/symbols')
  .then(response => response.json())
  .then(data => console.log(data));

// 获取实时价格
fetch('/api/v1/prices/BTCUSDT')
  .then(response => response.json())
  .then(data => console.log(data));

// 创建监控配置
fetch('/api/v1/monitoring-configs', {
  method: 'POST',
  headers: {
    'Content-Type': 'application/json',
  },
  body: JSON.stringify({
    name: 'BTC监控配置',
    symbol: 'BTCUSDT',
    interval: '1h',
    parameters: {
      threshold: 0.01
    }
  })
})
.then(response => response.json())
.then(data => console.log(data));
```

### Python (requests)

```python
import requests

# 获取交易对列表
response = requests.get('http://localhost:8080/api/v1/symbols')
data = response.json()
print(data)

# 获取实时价格
response = requests.get('http://localhost:8080/api/v1/prices/BTCUSDT')
data = response.json()
print(data)

# 创建监控配置
config_data = {
    'name': 'BTC监控配置',
    'symbol': 'BTCUSDT',
    'interval': '1h',
    'parameters': {
        'threshold': 0.01
    }
}
response = requests.post('http://localhost:8080/api/v1/monitoring-configs', json=config_data)
data = response.json()
print(data)
```

### cURL

```bash
# 获取交易对列表
curl -X GET "http://localhost:8080/api/v1/symbols"

# 获取实时价格
curl -X GET "http://localhost:8080/api/v1/prices/BTCUSDT"

# 创建监控配置
curl -X POST "http://localhost:8080/api/v1/monitoring-configs" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "BTC监控配置",
    "symbol": "BTCUSDT",
    "interval": "1h",
    "parameters": {
      "threshold": 0.01
    }
  }'
```

## 更新日志

### v1.0.0 (2024-01-01)
- 初始版本发布
- 支持交易对管理
- 支持历史K线数据查询
- 支持实时价格数据
- 支持监控配置管理
- 完整的Swagger文档

## 支持

如有问题或建议，请联系：
- 邮箱: support@swagger.io
- 文档: http://localhost:8080/swagger/index.html

