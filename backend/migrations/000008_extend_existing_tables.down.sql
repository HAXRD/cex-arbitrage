-- 回滚 price_ticks 表扩展
ALTER TABLE price_ticks DROP COLUMN IF EXISTS collection_source;
ALTER TABLE price_ticks DROP COLUMN IF EXISTS collection_latency;
ALTER TABLE price_ticks DROP COLUMN IF EXISTS is_anomaly;

-- 删除相关索引
DROP INDEX IF EXISTS idx_price_ticks_collection_source;
DROP INDEX IF EXISTS idx_price_ticks_is_anomaly;

-- 回滚 symbols 表扩展
ALTER TABLE symbols DROP COLUMN IF EXISTS collection_enabled;
ALTER TABLE symbols DROP COLUMN IF EXISTS collection_priority;
ALTER TABLE symbols DROP COLUMN IF EXISTS last_collected_at;

-- 删除相关索引
DROP INDEX IF EXISTS idx_symbols_collection_enabled;
DROP INDEX IF EXISTS idx_symbols_collection_priority;
