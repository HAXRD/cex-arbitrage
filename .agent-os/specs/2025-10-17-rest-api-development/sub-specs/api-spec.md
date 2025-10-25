# API规范

这是针对 @.agent-os/specs/2025-10-17-rest-api-development/spec.md 中详细规范的API规范文档

## 端点

### GET /api/v1/symbols

**用途：** 获取所有可用的交易对列表
**参数：**
- `page` (query, optional): 页码，默认1
- `limit` (query, optional): 每页数量，默认50，最大100
- `search` (query, optional): 搜索关键词（符号、基础币、报价币）
- `status` (query, optional): 状态筛选（normal, suspended, delisted）
- `symbol_type` (query, optional): 类型筛选（perpetual, spot）

**响应格式：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "symbols": [
      {
        "id": 1,
        "symbol": "BTCUSDT",
        "base_coin": "BTC",
        "quote_coin": "USDT",
        "symbol_type": "perpetual",
        "symbol_status": "normal",
        "min_trade_num": "0.01",
        "price_place": 1,
        "volume_place": 2,
        "is_active": true,
        "created_at": "2025-10-17T10:00:00Z",
        "updated_at": "2025-10-17T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 50,
      "total": 100,
      "pages": 2
    }
  }
}
```

**错误：**
- 400: 参数错误
- 500: 服务器内部错误

### GET /api/v1/symbols/{symbol}

**用途：** 获取特定交易对的详细信息
**参数：**
- `symbol` (path, required): 交易对符号（如BTCUSDT）

**响应格式：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "id": 1,
    "symbol": "BTCUSDT",
    "base_coin": "BTC",
    "quote_coin": "USDT",
    "symbol_type": "perpetual",
    "symbol_status": "normal",
    "min_trade_num": "0.01",
    "price_place": 1,
    "volume_place": 2,
    "is_active": true,
    "created_at": "2025-10-17T10:00:00Z",
    "updated_at": "2025-10-17T10:00:00Z"
  }
}
```

**错误：**
- 404: 交易对不存在
- 500: 服务器内部错误

### GET /api/v1/klines

**用途：** 获取历史K线数据
**参数：**
- `symbol` (query, required): 交易对符号
- `interval` (query, required): 时间间隔（1m, 5m, 15m, 1h, 4h, 1d）
- `start_time` (query, optional): 开始时间（Unix时间戳）
- `end_time` (query, optional): 结束时间（Unix时间戳）
- `limit` (query, optional): 数据条数，默认100，最大1000

**响应格式：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "symbol": "BTCUSDT",
    "interval": "1m",
    "klines": [
      {
        "timestamp": 1695793701000,
        "open": 45000.0,
        "high": 45100.0,
        "low": 44900.0,
        "close": 45050.0,
        "volume": 100.5,
        "amount": 4522500.0
      }
    ],
    "count": 100
  }
}
```

**错误：**
- 400: 参数错误（无效的时间间隔、时间范围等）
- 404: 交易对不存在
- 500: 服务器内部错误

### GET /api/v1/prices

**用途：** 获取实时价格数据
**参数：**
- `symbols` (query, optional): 交易对列表，逗号分隔（如BTCUSDT,ETHUSDT），不传则返回所有
- `fields` (query, optional): 返回字段，逗号分隔（price,change,volume等）

**响应格式：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "prices": [
      {
        "symbol": "BTCUSDT",
        "price": 45000.0,
        "change": 500.0,
        "change_percent": 1.12,
        "volume": 1000.5,
        "high_24h": 46000.0,
        "low_24h": 44000.0,
        "timestamp": 1695793701000
      }
    ],
    "count": 1,
    "updated_at": "2025-10-17T10:00:00Z"
  }
}
```

**错误：**
- 400: 参数错误
- 500: 服务器内部错误

### GET /api/v1/configs

**用途：** 获取用户配置列表
**参数：**
- `page` (query, optional): 页码，默认1
- `limit` (query, optional): 每页数量，默认20

**响应格式：**
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "configs": [
      {
        "id": 1,
        "name": "默认监控配置",
        "description": "基础价格监控配置",
        "filters": {
          "time_windows": ["1m", "5m", "15m"],
          "change_threshold": 5.0,
          "volume_threshold": 1000.0
        },
        "is_default": true,
        "created_at": "2025-10-17T10:00:00Z",
        "updated_at": "2025-10-17T10:00:00Z"
      }
    ],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 5,
      "pages": 1
    }
  }
}
```

**错误：**
- 500: 服务器内部错误

### POST /api/v1/configs

**用途：** 创建新的监控配置
**请求体：**
```json
{
  "name": "自定义配置",
  "description": "用户自定义的监控配置",
  "filters": {
    "time_windows": ["1m", "5m", "15m"],
    "change_threshold": 3.0,
    "volume_threshold": 500.0,
    "symbols": ["BTCUSDT", "ETHUSDT"]
  }
}
```

**响应格式：**
```json
{
  "code": 201,
  "message": "配置创建成功",
  "data": {
    "id": 2,
    "name": "自定义配置",
    "description": "用户自定义的监控配置",
    "filters": {
      "time_windows": ["1m", "5m", "15m"],
      "change_threshold": 3.0,
      "volume_threshold": 500.0,
      "symbols": ["BTCUSDT", "ETHUSDT"]
    },
    "is_default": false,
    "created_at": "2025-10-17T10:00:00Z",
    "updated_at": "2025-10-17T10:00:00Z"
  }
}
```

**错误：**
- 400: 请求参数错误
- 500: 服务器内部错误

### PUT /api/v1/configs/{id}

**用途：** 更新指定配置
**参数：**
- `id` (path, required): 配置ID

**请求体：** 同POST /api/v1/configs

**响应格式：** 同POST /api/v1/configs

**错误：**
- 400: 请求参数错误
- 404: 配置不存在
- 500: 服务器内部错误

### DELETE /api/v1/configs/{id}

**用途：** 删除指定配置
**参数：**
- `id` (path, required): 配置ID

**响应格式：**
```json
{
  "code": 200,
  "message": "配置删除成功",
  "data": null
}
```

**错误：**
- 404: 配置不存在
- 500: 服务器内部错误

### GET /api/v1/health

**用途：** 健康检查接口
**响应格式：**
```json
{
  "code": 200,
  "message": "服务正常",
  "data": {
    "status": "healthy",
    "timestamp": "2025-10-17T10:00:00Z",
    "version": "1.0.0",
    "services": {
      "database": "healthy",
      "redis": "healthy",
      "websocket": "healthy"
    }
  }
}
```

## 统一响应格式

### 成功响应
```json
{
  "code": 200,
  "message": "success",
  "data": { ... }
}
```

### 错误响应
```json
{
  "code": 400,
  "message": "参数错误",
  "data": null,
  "error": {
    "type": "validation_error",
    "details": "具体错误信息"
  }
}
```

## 错误码规范

- **200**: 成功
- **201**: 创建成功
- **400**: 请求参数错误
- **404**: 资源不存在
- **500**: 服务器内部错误
