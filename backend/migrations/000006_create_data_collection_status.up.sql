-- 创建数据采集状态表
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

-- 创建索引
CREATE UNIQUE INDEX idx_data_collection_status_symbol ON data_collection_status(symbol);
CREATE INDEX idx_data_collection_status_connection ON data_collection_status(connection_status);

-- 添加注释
COMMENT ON TABLE data_collection_status IS '数据采集状态表，监控每个交易对的采集状态';
COMMENT ON COLUMN data_collection_status.symbol IS '交易对符号，如 BTCUSDT';
COMMENT ON COLUMN data_collection_status.last_collected_at IS '最后采集时间';
COMMENT ON COLUMN data_collection_status.collection_count IS '累计采集次数';
COMMENT ON COLUMN data_collection_status.error_count IS '累计错误次数';
COMMENT ON COLUMN data_collection_status.connection_status IS '连接状态：connected, disconnected, error';
