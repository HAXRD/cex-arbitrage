-- 删除现有表的性能索引

-- 删除symbols表索引
DROP INDEX IF EXISTS idx_symbols_symbol;
DROP INDEX IF EXISTS idx_symbols_symbol_type;
DROP INDEX IF EXISTS idx_symbols_symbol_status;
DROP INDEX IF EXISTS idx_symbols_is_active;
DROP INDEX IF EXISTS idx_symbols_created_at;

-- 删除price_ticks表索引
DROP INDEX IF EXISTS idx_price_ticks_symbol_timestamp;
DROP INDEX IF EXISTS idx_price_ticks_timestamp;

-- 删除klines表索引
DROP INDEX IF EXISTS idx_klines_symbol_interval_timestamp;
DROP INDEX IF EXISTS idx_klines_symbol_timestamp;
