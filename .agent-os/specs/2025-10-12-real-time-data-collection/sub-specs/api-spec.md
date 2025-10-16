# API Specification

This is the API specification for the spec detailed in @.agent-os/specs/2025-10-12-real-time-data-collection/spec.md

## Endpoints

### GET /api/v1/data-collection/status

**Purpose:** 获取数据采集服务状态概览
**Parameters:** 
- `symbol` (optional): 指定交易对，不传则返回所有交易对状态
**Response:** 
```json
{
  "status": "running",
  "total_symbols": 100,
  "active_connections": 95,
  "error_rate": 0.02,
  "last_updated": "2024-01-01T12:00:00Z",
  "symbols": [
    {
      "symbol": "BTCUSDT",
      "status": "connected",
      "last_collected_at": "2024-01-01T12:00:00Z",
      "collection_count": 1000,
      "error_count": 5,
      "latency_ms": 50
    }
  ]
}
```
**Errors:** 
- 500: 服务内部错误
- 503: 服务不可用

### GET /api/v1/data-collection/config

**Purpose:** 获取数据采集配置列表
**Parameters:** 
- `active_only` (optional): 只返回活跃配置，默认false
- `symbol` (optional): 指定交易对配置
**Response:** 
```json
{
  "configs": [
    {
      "id": 1,
      "symbol": "BTCUSDT",
      "is_active": true,
      "collection_interval": 1000,
      "price_change_threshold": 0.01,
      "created_at": "2024-01-01T00:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ]
}
```
**Errors:** 
- 400: 参数错误
- 500: 服务内部错误

### PUT /api/v1/data-collection/config/{symbol}

**Purpose:** 更新指定交易对的数据采集配置
**Parameters:** 
- `symbol`: 交易对符号（路径参数）
**Request Body:** 
```json
{
  "is_active": true,
  "collection_interval": 500,
  "price_change_threshold": 0.005
}
```
**Response:** 
```json
{
  "success": true,
  "message": "配置更新成功",
  "config": {
    "id": 1,
    "symbol": "BTCUSDT",
    "is_active": true,
    "collection_interval": 500,
    "price_change_threshold": 0.005,
    "updated_at": "2024-01-01T12:00:00Z"
  }
}
```
**Errors:** 
- 400: 请求参数错误
- 404: 交易对不存在
- 500: 服务内部错误

### GET /api/v1/data-collection/price-changes

**Purpose:** 获取价格变化率数据
**Parameters:** 
- `symbol`: 交易对符号（必需）
- `window_size`: 时间窗口（1m, 5m, 15m）
- `start_time`: 开始时间（ISO 8601格式）
- `end_time`: 结束时间（ISO 8601格式）
- `limit`: 返回条数限制，默认100
- `offset`: 偏移量，默认0
**Response:** 
```json
{
  "data": [
    {
      "id": 1,
      "symbol": "BTCUSDT",
      "timestamp": "2024-01-01T12:00:00Z",
      "window_size": "1m",
      "change_rate": 0.025,
      "price_before": 50000.0,
      "price_after": 50125.0,
      "volume_24h": 1000000.0
    }
  ],
  "total": 1000,
  "limit": 100,
  "offset": 0
}
```
**Errors:** 
- 400: 参数错误
- 404: 数据不存在
- 500: 服务内部错误

### POST /api/v1/data-collection/start

**Purpose:** 启动数据采集服务
**Parameters:** 无
**Request Body:** 
```json
{
  "symbols": ["BTCUSDT", "ETHUSDT"],
  "force_restart": false
}
```
**Response:** 
```json
{
  "success": true,
  "message": "数据采集服务启动成功",
  "started_symbols": ["BTCUSDT", "ETHUSDT"],
  "started_at": "2024-01-01T12:00:00Z"
}
```
**Errors:** 
- 400: 请求参数错误
- 409: 服务已在运行
- 500: 服务启动失败

### POST /api/v1/data-collection/stop

**Purpose:** 停止数据采集服务
**Parameters:** 无
**Request Body:** 
```json
{
  "symbols": ["BTCUSDT", "ETHUSDT"],
  "graceful_shutdown": true
}
```
**Response:** 
```json
{
  "success": true,
  "message": "数据采集服务停止成功",
  "stopped_symbols": ["BTCUSDT", "ETHUSDT"],
  "stopped_at": "2024-01-01T12:00:00Z"
}
```
**Errors:** 
- 400: 请求参数错误
- 404: 服务未运行
- 500: 服务停止失败

### GET /api/v1/data-collection/health

**Purpose:** 数据采集服务健康检查
**Parameters:** 无
**Response:** 
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "uptime_seconds": 3600,
  "memory_usage_mb": 128,
  "cpu_usage_percent": 15.5,
  "active_connections": 95,
  "error_rate": 0.02
}
```
**Errors:** 
- 503: 服务不健康

### GET /api/v1/data-collection/metrics

**Purpose:** 获取数据采集性能指标
**Parameters:** 
- `time_range`: 时间范围（1h, 6h, 24h, 7d）
- `symbol` (optional): 指定交易对
**Response:** 
```json
{
  "time_range": "1h",
  "metrics": {
    "total_collections": 10000,
    "successful_collections": 9950,
    "failed_collections": 50,
    "average_latency_ms": 45,
    "max_latency_ms": 200,
    "error_rate": 0.005,
    "throughput_per_second": 2.78
  },
  "symbol_metrics": [
    {
      "symbol": "BTCUSDT",
      "collections": 1000,
      "success_rate": 0.995,
      "avg_latency_ms": 40
    }
  ]
}
```
**Errors:** 
- 400: 参数错误
- 500: 服务内部错误

## Controllers

### DataCollectionController

**职责**: 管理数据采集服务的启动、停止和状态监控

**主要方法**:
- `StartCollection()`: 启动数据采集
- `StopCollection()`: 停止数据采集
- `GetStatus()`: 获取采集状态
- `UpdateConfig()`: 更新采集配置
- `GetHealth()`: 健康检查

### PriceChangeController

**职责**: 处理价格变化率数据的查询和管理

**主要方法**:
- `GetPriceChanges()`: 查询价格变化率数据
- `CalculateChangeRate()`: 计算价格变化率
- `GetAnomalies()`: 获取异常波动数据

### MetricsController

**职责**: 收集和提供性能指标数据

**主要方法**:
- `GetMetrics()`: 获取性能指标
- `CollectMetrics()`: 收集指标数据
- `ExportMetrics()`: 导出指标数据

## Purpose

### 数据采集管理

- **服务控制**: 提供启动、停止、重启数据采集服务的接口
- **配置管理**: 支持动态调整采集参数和配置
- **状态监控**: 实时监控采集服务运行状态

### 数据查询服务

- **历史数据**: 提供价格变化率历史数据查询
- **实时状态**: 提供当前采集状态和性能指标
- **异常检测**: 支持异常波动数据查询

### 运维支持

- **健康检查**: 提供服务健康状态检查
- **性能监控**: 提供详细的性能指标和统计
- **故障诊断**: 支持错误日志和状态查询
