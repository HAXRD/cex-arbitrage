-- 为现有表添加性能索引

-- symbols表索引优化
CREATE INDEX IF NOT EXISTS idx_symbols_symbol ON symbols(symbol);
CREATE INDEX IF NOT EXISTS idx_symbols_symbol_type ON symbols(symbol_type);
CREATE INDEX IF NOT EXISTS idx_symbols_symbol_status ON symbols(symbol_status);
CREATE INDEX IF NOT EXISTS idx_symbols_is_active ON symbols(is_active);
CREATE INDEX IF NOT EXISTS idx_symbols_created_at ON symbols(created_at);

-- price_ticks表索引优化
CREATE INDEX IF NOT EXISTS idx_price_ticks_symbol_timestamp ON price_ticks(symbol, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_price_ticks_timestamp ON price_ticks(timestamp DESC);

-- klines表索引优化
CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval_timestamp ON klines(symbol, interval, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_klines_symbol_timestamp ON klines(symbol, timestamp DESC);
