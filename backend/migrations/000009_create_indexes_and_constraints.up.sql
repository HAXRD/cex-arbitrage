-- 为数据采集配置表创建额外索引
CREATE INDEX IF NOT EXISTS idx_data_collection_configs_updated_at ON data_collection_configs(updated_at);
CREATE INDEX IF NOT EXISTS idx_data_collection_configs_interval ON data_collection_configs(collection_interval);

-- 为数据采集状态表创建额外索引
CREATE INDEX IF NOT EXISTS idx_data_collection_status_updated_at ON data_collection_status(updated_at);
CREATE INDEX IF NOT EXISTS idx_data_collection_status_last_collected ON data_collection_status(last_collected_at DESC);
CREATE INDEX IF NOT EXISTS idx_data_collection_status_error_count ON data_collection_status(error_count) WHERE error_count > 0;

-- 为价格变化率表创建额外索引
CREATE INDEX IF NOT EXISTS idx_price_change_rates_created_at ON price_change_rates(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_price_change_rates_symbol_window ON price_change_rates(symbol, window_size);

-- 为扩展的 price_ticks 表创建复合索引
CREATE INDEX IF NOT EXISTS idx_price_ticks_symbol_anomaly ON price_ticks(symbol, is_anomaly) WHERE is_anomaly = true;
CREATE INDEX IF NOT EXISTS idx_price_ticks_timestamp_anomaly ON price_ticks(timestamp DESC, is_anomaly) WHERE is_anomaly = true;

-- 为扩展的 symbols 表创建复合索引
CREATE INDEX IF NOT EXISTS idx_symbols_priority_enabled ON symbols(collection_priority, collection_enabled) WHERE collection_enabled = true;
CREATE INDEX IF NOT EXISTS idx_symbols_last_collected ON symbols(last_collected_at DESC) WHERE last_collected_at IS NOT NULL;

-- 创建约束
-- 数据采集配置表约束
ALTER TABLE data_collection_configs ADD CONSTRAINT chk_collection_interval_positive CHECK (collection_interval > 0);
ALTER TABLE data_collection_configs ADD CONSTRAINT chk_price_change_threshold_positive CHECK (price_change_threshold > 0);

-- 数据采集状态表约束
ALTER TABLE data_collection_status ADD CONSTRAINT chk_collection_count_non_negative CHECK (collection_count >= 0);
ALTER TABLE data_collection_status ADD CONSTRAINT chk_error_count_non_negative CHECK (error_count >= 0);
ALTER TABLE data_collection_status ADD CONSTRAINT chk_connection_status_valid CHECK (connection_status IN ('connected', 'disconnected', 'error'));

-- 价格变化率表约束
ALTER TABLE price_change_rates ADD CONSTRAINT chk_window_size_valid CHECK (window_size IN ('1m', '5m', '15m', '1h', '4h', '1d'));
ALTER TABLE price_change_rates ADD CONSTRAINT chk_price_before_positive CHECK (price_before > 0);
ALTER TABLE price_change_rates ADD CONSTRAINT chk_price_after_positive CHECK (price_after > 0);

-- 扩展表约束
ALTER TABLE symbols ADD CONSTRAINT chk_collection_priority_positive CHECK (collection_priority > 0);
ALTER TABLE price_ticks ADD CONSTRAINT chk_collection_latency_non_negative CHECK (collection_latency IS NULL OR collection_latency >= 0);

-- 添加注释
COMMENT ON INDEX idx_data_collection_configs_updated_at IS '数据采集配置表更新时间索引';
COMMENT ON INDEX idx_data_collection_status_last_collected IS '数据采集状态表最后采集时间索引';
COMMENT ON INDEX idx_price_change_rates_symbol_window IS '价格变化率表交易对和窗口大小复合索引';
COMMENT ON INDEX idx_price_ticks_symbol_anomaly IS '价格数据表交易对和异常状态复合索引';