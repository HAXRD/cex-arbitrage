# Database Schema

This is the database schema implementation for the spec detailed in @.agent-os/specs/2025-10-08-data-storage-design/spec.md

## 数据库表设计

### 1. symbols - 交易对信息表

#### 表结构

```sql
CREATE TABLE symbols (
    id                      BIGSERIAL PRIMARY KEY,
    symbol                  VARCHAR(50) NOT NULL UNIQUE,        -- 交易对名称，如 BTCUSDT
    base_coin               VARCHAR(20) NOT NULL,               -- 基础币种
    quote_coin              VARCHAR(20) NOT NULL,               -- 计价币种
    buy_limit_price_ratio   DECIMAL(10, 4),                     -- 买限价比例
    sell_limit_price_ratio  DECIMAL(10, 4),                     -- 卖限价比例
    fee_rate_up_ratio       DECIMAL(10, 4),                     -- 手续费上浮比例
    maker_fee_rate          DECIMAL(10, 6),                     -- Maker 手续费率
    taker_fee_rate          DECIMAL(10, 6),                     -- Taker 手续费率
    open_cost_up_ratio      DECIMAL(10, 4),                     -- 开仓成本上浮比例
    support_margin_coins    TEXT[],                             -- 支持保证金币种（数组）
    min_trade_num           DECIMAL(20, 8),                     -- 最小开单数量
    price_end_step          DECIMAL(20, 8),                     -- 价格步长
    volume_place            INTEGER,                            -- 数量精度
    price_place             INTEGER,                            -- 价格精度
    size_multiplier         DECIMAL(20, 8),                     -- 数量乘数
    symbol_type             VARCHAR(20),                        -- 合约类型（perpetual/delivery）
    min_trade_usdt          DECIMAL(20, 2),                     -- 最小交易数量（USDT）
    max_symbol_order_num    INTEGER,                            -- 最大持有订单数（symbol维度）
    max_product_order_num   INTEGER,                            -- 最大持有订单数（产品类型维度）
    max_position_num        DECIMAL(20, 8),                     -- 最大持仓数量
    symbol_status           VARCHAR(20),                        -- 交易对状态
    off_time                BIGINT,                             -- 下线时间（-1表示正常）
    limit_open_time         BIGINT,                             -- 可开仓时间（-1表示正常）
    delivery_time           BIGINT,                             -- 交割时间
    delivery_start_time     BIGINT,                             -- 交割开始时间
    delivery_period         VARCHAR(20),                        -- 交割周期
    launch_time             BIGINT,                             -- 上线时间
    fund_interval           INTEGER,                            -- 资金费率间隔（小时）
    min_lever               DECIMAL(10, 2),                     -- 最小杠杆
    max_lever               DECIMAL(10, 2),                     -- 最大杠杆
    pos_limit               DECIMAL(10, 4),                     -- 持仓限制
    maintain_time           BIGINT,                             -- 维护时间
    max_market_order_qty    DECIMAL(20, 8),                     -- 单笔市价单最大下单数量
    max_order_qty           DECIMAL(20, 8),                     -- 单笔限价单最大下单数量
    is_active               BOOLEAN DEFAULT TRUE,               -- 是否激活监控
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 索引
CREATE INDEX idx_symbols_is_active ON symbols(is_active) WHERE is_active = TRUE;

-- 注释
COMMENT ON TABLE symbols IS '交易对信息表，存储从 BitGet API 获取的合约交易对配置';
COMMENT ON COLUMN symbols.symbol IS '交易对唯一标识，如 BTCUSDT';
COMMENT ON COLUMN symbols.is_active IS '是否激活监控，用于控制是否订阅该交易对的实时数据';
```

#### 设计说明
- **字段映射**：直接映射 BitGet API 的 Symbol 结构体字段，使用 snake_case 命名
- **数据类型**：价格和数量字段使用 DECIMAL 保证精度，避免浮点数精度问题
- **唯一约束**：symbol 字段设置唯一约束，防止重复插入，已自动创建索引
- **数组类型**：support_margin_coins 使用 PostgreSQL 原生数组类型存储多个保证金币种
- **业务字段**：增加 is_active 字段控制是否监控该交易对
- **索引策略**：
  - symbol 字段有唯一约束，已自动创建唯一索引，无需额外索引
  - 仅保留 `is_active` 部分索引（WHERE is_active = TRUE），用于快速查询激活的交易对
  - 其他字段（base_coin、quote_coin、symbol_type、symbol_status）查询频率极低，不创建索引
  - symbols 表数据量小（预计<1000条），全表扫描性能可接受

### 2. price_ticks - 实时价格数据表（时序表）

#### 表结构

```sql
CREATE TABLE price_ticks (
    symbol                  VARCHAR(50) NOT NULL,               -- 交易对名称
    timestamp               TIMESTAMP WITH TIME ZONE NOT NULL,  -- 数据时间戳
    last_price              DECIMAL(20, 8) NOT NULL,            -- 最新成交价
    ask_price               DECIMAL(20, 8),                     -- 卖一价
    bid_price               DECIMAL(20, 8),                     -- 买一价
    bid_size                DECIMAL(20, 8),                     -- 买一量
    ask_size                DECIMAL(20, 8),                     -- 卖一量
    high_24h                DECIMAL(20, 8),                     -- 24小时最高价
    low_24h                 DECIMAL(20, 8),                     -- 24小时最低价
    change_24h              DECIMAL(10, 4),                     -- 24小时价格涨跌幅（百分比）
    base_volume             DECIMAL(30, 8),                     -- 交易币交易量
    quote_volume            DECIMAL(30, 8),                     -- 计价币交易量
    usdt_volume             DECIMAL(30, 8),                     -- USDT交易量
    open_utc                DECIMAL(20, 8),                     -- 开盘价（UTC+0时区）
    change_utc_24h          DECIMAL(10, 4),                     -- 24小时价格涨跌幅（UTC+0时区）
    index_price             DECIMAL(20, 8),                     -- 指数价格
    funding_rate            DECIMAL(10, 6),                     -- 资金费率
    holding_amount          DECIMAL(30, 8),                     -- 当前持仓量
    open_24h                DECIMAL(20, 8),                     -- 开盘价（24小时）
    mark_price              DECIMAL(20, 8),                     -- 标记价格
    delivery_start_time     BIGINT,                             -- 交割开始时间
    delivery_time           BIGINT,                             -- 交割时间
    delivery_status         VARCHAR(30),                        -- 交割状态
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- 转换为 TimescaleDB 超表（Hypertable）
SELECT create_hypertable('price_ticks', 'timestamp', 
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- 索引
CREATE INDEX idx_price_ticks_symbol_timestamp ON price_ticks(symbol, timestamp DESC);
CREATE INDEX idx_price_ticks_timestamp ON price_ticks(timestamp DESC);

-- 数据压缩策略（7天以上的数据压缩）
ALTER TABLE price_ticks SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol',
    timescaledb.compress_orderby = 'timestamp DESC'
);

SELECT add_compression_policy('price_ticks', INTERVAL '7 days');

-- 数据保留策略（自动删除30天以上的数据）
SELECT add_retention_policy('price_ticks', INTERVAL '30 days');

-- 注释
COMMENT ON TABLE price_ticks IS '实时价格数据表（时序表），存储从 WebSocket 接收的 Ticker 数据';
COMMENT ON COLUMN price_ticks.timestamp IS '数据时间戳，来自 BitGet API 的 ts 字段';
COMMENT ON COLUMN price_ticks.last_price IS '最新成交价，核心价格字段';
```

#### 设计说明
- **时序表**：使用 TimescaleDB 的 `create_hypertable` 转换为超表，优化时间序列数据查询
- **分片策略**：按时间每7天分片，平衡查询性能和管理复杂度
- **压缩策略**：7天以上数据自动压缩，节省存储空间（压缩比约5:1）
- **保留策略**：30天以上数据自动删除，避免数据无限增长
- **索引设计**：
  - 联合索引 `(symbol, timestamp DESC)`：优化按交易对查询最新数据
  - 单独索引 `timestamp DESC`：优化全局时间范围查询
- **无主键**：时序表通常不需要主键，使用 `(symbol, timestamp)` 组合唯一标识

### 3. klines - K线数据表（时序表）

#### 表结构

```sql
CREATE TABLE klines (
    symbol                  VARCHAR(50) NOT NULL,               -- 交易对名称
    timestamp               TIMESTAMP WITH TIME ZONE NOT NULL,  -- K线开始时间
    granularity             VARCHAR(10) NOT NULL,               -- K线周期（1m, 5m, 15m, 1h等）
    open                    DECIMAL(20, 8) NOT NULL,            -- 开盘价
    high                    DECIMAL(20, 8) NOT NULL,            -- 最高价
    low                     DECIMAL(20, 8) NOT NULL,            -- 最低价
    close                   DECIMAL(20, 8) NOT NULL,            -- 收盘价
    base_volume             DECIMAL(30, 8) NOT NULL,            -- 交易币成交量
    quote_volume            DECIMAL(30, 8) NOT NULL,            -- 计价币成交量
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, timestamp, granularity)                      -- 防止重复插入
);

-- 转换为 TimescaleDB 超表（Hypertable）
SELECT create_hypertable('klines', 'timestamp', 
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- 索引
CREATE INDEX idx_klines_symbol_granularity_timestamp ON klines(symbol, granularity, timestamp DESC);
CREATE INDEX idx_klines_timestamp ON klines(timestamp DESC);
CREATE INDEX idx_klines_granularity ON klines(granularity);

-- 数据压缩策略（7天以上的数据压缩）
ALTER TABLE klines SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol, granularity',
    timescaledb.compress_orderby = 'timestamp DESC'
);

SELECT add_compression_policy('klines', INTERVAL '7 days');

-- 数据保留策略（自动删除30天以上的数据）
SELECT add_retention_policy('klines', INTERVAL '30 days');

-- 注释
COMMENT ON TABLE klines IS 'K线数据表（时序表），存储从 REST API 获取的历史K线数据';
COMMENT ON COLUMN klines.granularity IS 'K线周期：1m, 5m, 15m, 30m, 1H, 4H, 1D';
COMMENT ON COLUMN klines.timestamp IS 'K线开始时间，来自 BitGet API 的 ts 字段';
```

#### 设计说明
- **唯一约束**：`(symbol, timestamp, granularity)` 组合唯一，防止重复插入相同K线
- **分片策略**：按时间每7天分片
- **压缩策略**：按 `(symbol, granularity)` 分段压缩，同一交易对、同一周期的数据压缩在一起
- **索引设计**：
  - 联合索引 `(symbol, granularity, timestamp DESC)`：优化按交易对和周期查询K线
  - 单独索引 `granularity`：支持按周期过滤查询
- **OHLCV 字段**：标准K线数据结构（开盘价、最高价、最低价、收盘价、成交量）

## 迁移文件

### 迁移文件 1：创建基础表

**文件名**：`20251008000001_create_base_tables.up.sql`

```sql
-- 启用 TimescaleDB 扩展
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 创建 symbols 表
CREATE TABLE symbols (
    id                      BIGSERIAL PRIMARY KEY,
    symbol                  VARCHAR(50) NOT NULL UNIQUE,
    base_coin               VARCHAR(20) NOT NULL,
    quote_coin              VARCHAR(20) NOT NULL,
    buy_limit_price_ratio   DECIMAL(10, 4),
    sell_limit_price_ratio  DECIMAL(10, 4),
    fee_rate_up_ratio       DECIMAL(10, 4),
    maker_fee_rate          DECIMAL(10, 6),
    taker_fee_rate          DECIMAL(10, 6),
    open_cost_up_ratio      DECIMAL(10, 4),
    support_margin_coins    TEXT[],
    min_trade_num           DECIMAL(20, 8),
    price_end_step          DECIMAL(20, 8),
    volume_place            INTEGER,
    price_place             INTEGER,
    size_multiplier         DECIMAL(20, 8),
    symbol_type             VARCHAR(20),
    min_trade_usdt          DECIMAL(20, 2),
    max_symbol_order_num    INTEGER,
    max_product_order_num   INTEGER,
    max_position_num        DECIMAL(20, 8),
    symbol_status           VARCHAR(20),
    off_time                BIGINT,
    limit_open_time         BIGINT,
    delivery_time           BIGINT,
    delivery_start_time     BIGINT,
    delivery_period         VARCHAR(20),
    launch_time             BIGINT,
    fund_interval           INTEGER,
    min_lever               DECIMAL(10, 2),
    max_lever               DECIMAL(10, 2),
    pos_limit               DECIMAL(10, 4),
    maintain_time           BIGINT,
    max_market_order_qty    DECIMAL(20, 8),
    max_order_qty           DECIMAL(20, 8),
    is_active               BOOLEAN DEFAULT TRUE,
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_symbols_is_active ON symbols(is_active) WHERE is_active = TRUE;

-- 创建 price_ticks 表
CREATE TABLE price_ticks (
    symbol                  VARCHAR(50) NOT NULL,
    timestamp               TIMESTAMP WITH TIME ZONE NOT NULL,
    last_price              DECIMAL(20, 8) NOT NULL,
    ask_price               DECIMAL(20, 8),
    bid_price               DECIMAL(20, 8),
    bid_size                DECIMAL(20, 8),
    ask_size                DECIMAL(20, 8),
    high_24h                DECIMAL(20, 8),
    low_24h                 DECIMAL(20, 8),
    change_24h              DECIMAL(10, 4),
    base_volume             DECIMAL(30, 8),
    quote_volume            DECIMAL(30, 8),
    usdt_volume             DECIMAL(30, 8),
    open_utc                DECIMAL(20, 8),
    change_utc_24h          DECIMAL(10, 4),
    index_price             DECIMAL(20, 8),
    funding_rate            DECIMAL(10, 6),
    holding_amount          DECIMAL(30, 8),
    open_24h                DECIMAL(20, 8),
    mark_price              DECIMAL(20, 8),
    delivery_start_time     BIGINT,
    delivery_time           BIGINT,
    delivery_status         VARCHAR(30),
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

SELECT create_hypertable('price_ticks', 'timestamp', 
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

CREATE INDEX idx_price_ticks_symbol_timestamp ON price_ticks(symbol, timestamp DESC);
CREATE INDEX idx_price_ticks_timestamp ON price_ticks(timestamp DESC);

-- 创建 klines 表
CREATE TABLE klines (
    symbol                  VARCHAR(50) NOT NULL,
    timestamp               TIMESTAMP WITH TIME ZONE NOT NULL,
    granularity             VARCHAR(10) NOT NULL,
    open                    DECIMAL(20, 8) NOT NULL,
    high                    DECIMAL(20, 8) NOT NULL,
    low                     DECIMAL(20, 8) NOT NULL,
    close                   DECIMAL(20, 8) NOT NULL,
    base_volume             DECIMAL(30, 8) NOT NULL,
    quote_volume            DECIMAL(30, 8) NOT NULL,
    created_at              TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(symbol, timestamp, granularity)
);

SELECT create_hypertable('klines', 'timestamp', 
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

CREATE INDEX idx_klines_symbol_granularity_timestamp ON klines(symbol, granularity, timestamp DESC);
CREATE INDEX idx_klines_timestamp ON klines(timestamp DESC);
CREATE INDEX idx_klines_granularity ON klines(granularity);
```

**回滚文件**：`20251008000001_create_base_tables.down.sql`

```sql
DROP TABLE IF EXISTS klines CASCADE;
DROP TABLE IF EXISTS price_ticks CASCADE;
DROP TABLE IF EXISTS symbols CASCADE;
DROP EXTENSION IF EXISTS timescaledb CASCADE;
```

### 迁移文件 2：配置数据压缩和保留策略

**文件名**：`20251008000002_configure_timescaledb_policies.up.sql`

```sql
-- 配置 price_ticks 压缩策略
ALTER TABLE price_ticks SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol',
    timescaledb.compress_orderby = 'timestamp DESC'
);

SELECT add_compression_policy('price_ticks', INTERVAL '7 days');

-- 配置 price_ticks 保留策略
SELECT add_retention_policy('price_ticks', INTERVAL '30 days');

-- 配置 klines 压缩策略
ALTER TABLE klines SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol, granularity',
    timescaledb.compress_orderby = 'timestamp DESC'
);

SELECT add_compression_policy('klines', INTERVAL '7 days');

-- 配置 klines 保留策略
SELECT add_retention_policy('klines', INTERVAL '30 days');
```

**回滚文件**：`20251008000002_configure_timescaledb_policies.down.sql`

```sql
-- 删除保留策略
SELECT remove_retention_policy('klines', if_exists => true);
SELECT remove_retention_policy('price_ticks', if_exists => true);

-- 删除压缩策略
SELECT remove_compression_policy('klines', if_exists => true);
SELECT remove_compression_policy('price_ticks', if_exists => true);

-- 禁用压缩
ALTER TABLE klines SET (timescaledb.compress = false);
ALTER TABLE price_ticks SET (timescaledb.compress = false);
```

## Redis 缓存键设计

### 1. 实时价格缓存

**键格式**：`cex:price:{symbol}`

**数据类型**：Hash

**字段**：
```
last_price      - 最新成交价
ask_price       - 卖一价
bid_price       - 买一价
high_24h        - 24小时最高价
low_24h         - 24小时最低价
change_24h      - 24小时涨跌幅
base_volume     - 交易币交易量
quote_volume    - 计价币交易量
timestamp       - 更新时间戳
```

**TTL**：60秒

**示例**：
```
cex:price:BTCUSDT {
    last_price: "43250.5",
    ask_price: "43251.0",
    bid_price: "43250.0",
    high_24h: "44000.0",
    low_24h: "42500.0",
    change_24h: "2.35",
    base_volume: "1234.56",
    quote_volume: "53421789.12",
    timestamp: "1696789012345"
}
```

### 2. 实时指标缓存

**键格式**：`cex:metrics:{symbol}:{window}`

**数据类型**：Hash

**字段**：
```
change_pct      - 涨跌幅（百分比）
change_value    - 涨跌值
start_price     - 起始价格
end_price       - 结束价格
high_price      - 区间最高价
low_price       - 区间最低价
volume          - 区间成交量
```

**TTL**：60秒

**示例**：
```
cex:metrics:BTCUSDT:1m {
    change_pct: "0.15",
    change_value: "65.0",
    start_price: "43185.5",
    end_price: "43250.5",
    high_price: "43260.0",
    low_price: "43180.0",
    volume: "12.34"
}
```

### 3. 交易对列表缓存

**键格式**：`cex:symbols:active`

**数据类型**：Set

**成员**：交易对 symbol 列表

**TTL**：300秒（5分钟）

**示例**：
```
cex:symbols:active {
    "BTCUSDT",
    "ETHUSDT",
    "BNBUSDT",
    ...
}
```

### 4. WebSocket 连接管理

**键格式**：`cex:ws:connections`

**数据类型**：Set

**成员**：连接ID（UUID）

**TTL**：90秒（心跳超时时间）

**示例**：
```
cex:ws:connections {
    "conn-uuid-1234-5678-abcd",
    "conn-uuid-2345-6789-bcde",
    ...
}
```

### 5. 订阅关系缓存

**键格式**：`cex:ws:subscriptions:{conn_id}`

**数据类型**：Set

**成员**：该连接订阅的交易对列表

**TTL**：90秒（随连接过期）

**示例**：
```
cex:ws:subscriptions:conn-uuid-1234-5678-abcd {
    "BTCUSDT",
    "ETHUSDT"
}
```

## 数据库关系图

```
┌─────────────────────────────────────┐
│           symbols 表                 │
│  (交易对信息，普通表)                │
│                                     │
│  - id (PK)                          │
│  - symbol (UNIQUE)                  │
│  - base_coin, quote_coin            │
│  - 费率、杠杆等配置信息              │
│  - is_active (业务字段)             │
└─────────────────────────────────────┘
                │
                │ 1:N (逻辑关联)
                │
    ┌───────────┴───────────┐
    │                       │
    ▼                       ▼
┌─────────────┐      ┌──────────────┐
│ price_ticks │      │   klines     │
│ (时序表)    │      │  (时序表)    │
│             │      │              │
│ - symbol    │      │ - symbol     │
│ - timestamp │      │ - timestamp  │
│ - last_price│      │ - granularity│
│ - 24h数据   │      │ - OHLCV      │
│ - 成交量    │      │              │
└─────────────┘      └──────────────┘
```

## 性能优化建议

### 1. 数据库层面
- **分区表**：TimescaleDB 自动按时间分片，无需手动分区
- **索引优化**：定期分析慢查询，调整索引策略
- **连接池**：合理配置连接池大小，避免连接耗尽
- **批量操作**：使用批量插入减少数据库往返次数

### 2. 缓存层面
- **预热策略**：系统启动时预加载活跃交易对数据到 Redis
- **缓存雪崩防护**：使用随机TTL，避免大量缓存同时过期
- **缓存穿透防护**：使用布隆过滤器过滤不存在的交易对查询
- **缓存击穿防护**：热点数据使用互斥锁防止并发查询数据库

### 3. 查询优化
- **时间范围查询**：始终使用索引字段 `timestamp`，并指定合理的时间范围
- **避免全表扫描**：查询时必须包含 `symbol` 或 `timestamp` 条件
- **分页查询**：大结果集使用 LIMIT 和 OFFSET 分页

### 4. 写入优化
- **异步批量写入**：实时 Ticker 数据先写缓存，批量写数据库（每5秒一次）
- **写入缓冲**：使用 Go channel 作为写入缓冲区，平滑写入峰值
- **忽略重复**：使用 `ON CONFLICT DO NOTHING` 忽略重复K线数据

## 验证和测试

### 数据库验证
```sql
-- 验证 TimescaleDB 扩展已安装
SELECT * FROM pg_extension WHERE extname = 'timescaledb';

-- 验证超表创建成功
SELECT * FROM timescaledb_information.hypertables;

-- 验证压缩策略
SELECT * FROM timescaledb_information.compression_settings;

-- 验证保留策略
SELECT * FROM timescaledb_information.jobs WHERE proc_name = 'policy_retention';

-- 查看表空间占用
SELECT 
    hypertable_name,
    pg_size_pretty(hypertable_size(format('%I.%I', hypertable_schema, hypertable_name)::regclass)) AS size
FROM timescaledb_information.hypertables;
```

### 性能测试
```sql
-- 测试查询性能（查询1天K线数据）
EXPLAIN ANALYZE
SELECT * FROM klines
WHERE symbol = 'BTCUSDT' 
    AND granularity = '1m'
    AND timestamp >= NOW() - INTERVAL '1 day'
ORDER BY timestamp DESC;

-- 测试批量插入性能
\timing
INSERT INTO price_ticks (symbol, timestamp, last_price, base_volume, quote_volume)
SELECT 
    'TEST' || i::text,
    NOW() - (i || ' seconds')::interval,
    random() * 50000,
    random() * 1000,
    random() * 50000000
FROM generate_series(1, 10000) i;
```

## 备份策略

### PostgreSQL 备份
- **全量备份**：每日凌晨3点执行 `pg_dump`
- **增量备份**：使用 WAL 归档实现增量备份
- **保留策略**：保留最近7天的全量备份

### Redis 备份
- **RDB 快照**：每5分钟自动保存
- **持久化文件**：保留最近3个 RDB 文件
- **恢复策略**：实时数据可从数据库恢复，丢失风险可接受
