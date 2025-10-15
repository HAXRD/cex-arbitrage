-- 创建数据采集配置表
CREATE TABLE data_collection_configs (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true,
    collection_interval INTEGER NOT NULL DEFAULT 1000, -- 采集间隔(毫秒)
    price_change_threshold DECIMAL(10,4) NOT NULL DEFAULT 0.01, -- 价格变化阈值
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX idx_data_collection_configs_symbol ON data_collection_configs(symbol);
CREATE INDEX idx_data_collection_configs_active ON data_collection_configs(is_active) WHERE is_active = true;

-- 添加注释
COMMENT ON TABLE data_collection_configs IS '数据采集配置表，存储每个交易对的采集参数';
COMMENT ON COLUMN data_collection_configs.symbol IS '交易对符号，如 BTCUSDT';
COMMENT ON COLUMN data_collection_configs.is_active IS '是否启用采集';
COMMENT ON COLUMN data_collection_configs.collection_interval IS '采集间隔，单位毫秒';
COMMENT ON COLUMN data_collection_configs.price_change_threshold IS '价格变化阈值，用于异常检测';
