# Database Schema

This is the database schema implementation for the spec detailed in @.agent-os/specs/2025-10-12-real-time-data-collection/spec.md

## Changes

### 新增表结构

#### 1. 数据采集配置表 (data_collection_configs)

```sql
CREATE TABLE data_collection_configs (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    collection_interval INTEGER NOT NULL DEFAULT 1000, -- 采集间隔(毫秒)
    price_change_threshold DECIMAL(10,4) NOT NULL DEFAULT 0.01, -- 价格变化阈值
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_data_collection_configs_symbol ON data_collection_configs(symbol);
CREATE INDEX idx_data_collection_configs_active ON data_collection_configs(is_active) WHERE is_active = true;
```

#### 2. 数据采集状态表 (data_collection_status)

```sql
CREATE TABLE data_collection_status (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    last_collected_at TIMESTAMP WITH TIME ZONE,
    collection_count BIGINT NOT NULL DEFAULT 0,
    error_count BIGINT NOT NULL DEFAULT 0,
    last_error_message TEXT,
    connection_status VARCHAR(20) NOT NULL DEFAULT 'disconnected', -- connected, disconnected, error
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_data_collection_status_symbol ON data_collection_status(symbol);
CREATE INDEX idx_data_collection_status_connection ON data_collection_status(connection_status);
```

#### 3. 价格变化率表 (price_change_rates)

```sql
CREATE TABLE price_change_rates (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    window_size VARCHAR(10) NOT NULL, -- 1m, 5m, 15m
    change_rate DECIMAL(10,6) NOT NULL, -- 变化率百分比
    price_before DECIMAL(20,8) NOT NULL, -- 变化前价格
    price_after DECIMAL(20,8) NOT NULL, -- 变化后价格
    volume_24h DECIMAL(20,8), -- 24小时交易量
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 创建TimescaleDB超表
SELECT create_hypertable('price_change_rates', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day');

-- 创建索引
CREATE INDEX idx_price_change_rates_symbol_timestamp ON price_change_rates(symbol, timestamp DESC);
CREATE INDEX idx_price_change_rates_window_size ON price_change_rates(window_size);
CREATE INDEX idx_price_change_rates_change_rate ON price_change_rates(change_rate);
```

### 修改现有表结构

#### 1. 扩展 price_ticks 表

```sql
-- 添加数据采集相关字段
ALTER TABLE price_ticks ADD COLUMN collection_source VARCHAR(20) DEFAULT 'websocket';
ALTER TABLE price_ticks ADD COLUMN collection_latency INTEGER; -- 采集延迟(毫秒)
ALTER TABLE price_ticks ADD COLUMN is_anomaly BOOLEAN DEFAULT false; -- 是否异常数据

-- 创建新索引
CREATE INDEX idx_price_ticks_collection_source ON price_ticks(collection_source);
CREATE INDEX idx_price_ticks_is_anomaly ON price_ticks(is_anomaly) WHERE is_anomaly = true;
```

#### 2. 扩展 symbols 表

```sql
-- 添加数据采集配置字段
ALTER TABLE symbols ADD COLUMN collection_enabled BOOLEAN DEFAULT true;
ALTER TABLE symbols ADD COLUMN collection_priority INTEGER DEFAULT 1; -- 采集优先级
ALTER TABLE symbols ADD COLUMN last_collected_at TIMESTAMP WITH TIME ZONE;

-- 创建索引
CREATE INDEX idx_symbols_collection_enabled ON symbols(collection_enabled) WHERE collection_enabled = true;
CREATE INDEX idx_symbols_collection_priority ON symbols(collection_priority);
```

## Specifications

### 数据采集配置表

- **主键**: 自增ID
- **唯一约束**: 交易对符号唯一
- **索引**: 交易对符号、活跃状态
- **字段说明**:
  - `collection_interval`: 采集间隔，支持毫秒级配置
  - `price_change_threshold`: 价格变化阈值，用于异常检测
  - `is_active`: 是否启用采集

### 数据采集状态表

- **主键**: 自增ID
- **唯一约束**: 交易对符号唯一
- **索引**: 交易对符号、连接状态
- **字段说明**:
  - `last_collected_at`: 最后采集时间
  - `collection_count`: 累计采集次数
  - `error_count`: 累计错误次数
  - `connection_status`: 连接状态枚举

### 价格变化率表

- **主键**: 自增ID
- **TimescaleDB超表**: 按时间分片，1天分片
- **索引**: 交易对+时间、窗口大小、变化率
- **字段说明**:
  - `window_size`: 时间窗口大小（1m、5m、15m）
  - `change_rate`: 变化率百分比
  - `price_before/after`: 变化前后价格

## Rationale

### 数据采集配置表设计理由

- **独立配置**: 每个交易对可以独立配置采集参数
- **灵活控制**: 支持动态启用/禁用特定交易对的采集
- **性能优化**: 通过采集间隔控制数据频率
- **异常检测**: 通过阈值配置实现异常波动检测

### 数据采集状态表设计理由

- **状态监控**: 实时监控每个交易对的采集状态
- **错误统计**: 统计采集成功率和错误率
- **故障诊断**: 记录错误信息便于问题排查
- **连接管理**: 跟踪WebSocket连接状态

### 价格变化率表设计理由

- **时序数据**: 使用TimescaleDB优化时序数据存储
- **多时间窗口**: 支持不同时间窗口的变化率计算
- **快速查询**: 通过索引优化查询性能
- **异常检测**: 通过变化率字段实现异常波动识别

### 扩展现有表设计理由

- **向后兼容**: 通过ADD COLUMN方式扩展，不影响现有数据
- **数据溯源**: 添加采集来源字段，便于数据质量分析
- **性能监控**: 添加采集延迟字段，监控数据采集性能
- **异常标记**: 添加异常数据标记，便于后续分析
