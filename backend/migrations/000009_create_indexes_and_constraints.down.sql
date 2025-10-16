-- 删除索引
DROP INDEX IF EXISTS idx_data_collection_configs_updated_at;
DROP INDEX IF EXISTS idx_data_collection_configs_interval;
DROP INDEX IF EXISTS idx_data_collection_status_updated_at;
DROP INDEX IF EXISTS idx_data_collection_status_last_collected;
DROP INDEX IF EXISTS idx_data_collection_status_error_count;
DROP INDEX IF EXISTS idx_price_change_rates_created_at;
DROP INDEX IF EXISTS idx_price_change_rates_symbol_window;
DROP INDEX IF EXISTS idx_price_ticks_symbol_anomaly;
DROP INDEX IF EXISTS idx_price_ticks_timestamp_anomaly;
DROP INDEX IF EXISTS idx_symbols_priority_enabled;
DROP INDEX IF EXISTS idx_symbols_last_collected;

-- 删除约束
ALTER TABLE data_collection_configs DROP CONSTRAINT IF EXISTS chk_collection_interval_positive;
ALTER TABLE data_collection_configs DROP CONSTRAINT IF EXISTS chk_price_change_threshold_positive;
ALTER TABLE data_collection_status DROP CONSTRAINT IF EXISTS chk_collection_count_non_negative;
ALTER TABLE data_collection_status DROP CONSTRAINT IF EXISTS chk_error_count_non_negative;
ALTER TABLE data_collection_status DROP CONSTRAINT IF EXISTS chk_connection_status_valid;
ALTER TABLE price_change_rates DROP CONSTRAINT IF EXISTS chk_window_size_valid;
ALTER TABLE price_change_rates DROP CONSTRAINT IF EXISTS chk_price_before_positive;
ALTER TABLE price_change_rates DROP CONSTRAINT IF EXISTS chk_price_after_positive;
ALTER TABLE symbols DROP CONSTRAINT IF EXISTS chk_collection_priority_positive;
ALTER TABLE price_ticks DROP CONSTRAINT IF EXISTS chk_collection_latency_non_negative;
