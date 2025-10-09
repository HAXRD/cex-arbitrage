-- 回滚 TimescaleDB 数据压缩和保留策略

-- 删除 klines 表的策略
SELECT remove_retention_policy('klines', if_exists => true);
SELECT remove_compression_policy('klines', if_exists => true);

-- 禁用 klines 表的压缩
ALTER TABLE klines SET (timescaledb.compress = false);

-- 删除 price_ticks 表的策略
SELECT remove_retention_policy('price_ticks', if_exists => true);
SELECT remove_compression_policy('price_ticks', if_exists => true);

-- 禁用 price_ticks 表的压缩
ALTER TABLE price_ticks SET (timescaledb.compress = false);

-- 输出回滚信息
DO $$
BEGIN
    RAISE NOTICE '✓ TimescaleDB policies removed successfully';
END $$;

