-- 创建 price_ticks 时序表（实时价格数据表）
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

-- 转换为 TimescaleDB 超表（7天分片）
SELECT create_hypertable('price_ticks', 'timestamp', 
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- 创建索引
CREATE INDEX idx_price_ticks_symbol_timestamp ON price_ticks(symbol, timestamp DESC);
CREATE INDEX idx_price_ticks_timestamp ON price_ticks(timestamp DESC);

-- 添加注释
COMMENT ON TABLE price_ticks IS '实时价格数据表（时序表），存储从 WebSocket 接收的 Ticker 数据';
COMMENT ON COLUMN price_ticks.timestamp IS '数据时间戳，来自 BitGet API 的 ts 字段';
COMMENT ON COLUMN price_ticks.last_price IS '最新成交价，核心价格字段';

-- 创建 klines 时序表（K线数据表）
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

-- 转换为 TimescaleDB 超表（7天分片）
SELECT create_hypertable('klines', 'timestamp', 
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- 创建索引
CREATE INDEX idx_klines_symbol_granularity_timestamp ON klines(symbol, granularity, timestamp DESC);
CREATE INDEX idx_klines_timestamp ON klines(timestamp DESC);
CREATE INDEX idx_klines_granularity ON klines(granularity);

-- 添加注释
COMMENT ON TABLE klines IS 'K线数据表（时序表），存储从 REST API 获取的历史K线数据';
COMMENT ON COLUMN klines.granularity IS 'K线周期：1m, 5m, 15m, 30m, 1H, 4H, 1D';
COMMENT ON COLUMN klines.timestamp IS 'K线开始时间，来自 BitGet API 的 ts 字段';

