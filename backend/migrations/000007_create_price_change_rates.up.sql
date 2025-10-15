-- 创建价格变化率表
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

-- 添加注释
COMMENT ON TABLE price_change_rates IS '价格变化率表，存储不同时间窗口的价格变化率数据';
COMMENT ON COLUMN price_change_rates.symbol IS '交易对符号，如 BTCUSDT';
COMMENT ON COLUMN price_change_rates.timestamp IS '数据时间戳';
COMMENT ON COLUMN price_change_rates.window_size IS '时间窗口大小：1m, 5m, 15m';
COMMENT ON COLUMN price_change_rates.change_rate IS '价格变化率百分比';
COMMENT ON COLUMN price_change_rates.price_before IS '变化前价格';
COMMENT ON COLUMN price_change_rates.price_after IS '变化后价格';
