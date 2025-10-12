-- 配置 TimescaleDB 数据压缩和保留策略

-- ========================================
-- 配置 price_ticks 表压缩策略
-- ========================================

-- 启用压缩
ALTER TABLE price_ticks SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol',
    timescaledb.compress_orderby = 'timestamp DESC'
);

-- 添加压缩策略（7天后压缩）
SELECT add_compression_policy('price_ticks', INTERVAL '7 days');

-- 添加数据保留策略（30天后删除）
SELECT add_retention_policy('price_ticks', INTERVAL '30 days');

-- ========================================
-- 配置 klines 表压缩策略
-- ========================================

-- 启用压缩
ALTER TABLE klines SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol, granularity',
    timescaledb.compress_orderby = 'timestamp DESC'
);

-- 添加压缩策略（7天后压缩）
SELECT add_compression_policy('klines', INTERVAL '7 days');

-- 添加数据保留策略（30天后删除）
SELECT add_retention_policy('klines', INTERVAL '30 days');

-- ========================================
-- 验证策略配置
-- ========================================

-- 输出策略信息
DO $$
DECLARE
    compression_jobs INT;
    retention_jobs INT;
BEGIN
    SELECT COUNT(*) INTO compression_jobs 
    FROM timescaledb_information.jobs 
    WHERE proc_name = 'policy_compression';
    
    SELECT COUNT(*) INTO retention_jobs 
    FROM timescaledb_information.jobs 
    WHERE proc_name = 'policy_retention';
    
    RAISE NOTICE 'Compression policies configured: %', compression_jobs;
    RAISE NOTICE 'Retention policies configured: %', retention_jobs;
    RAISE NOTICE '✓ TimescaleDB policies configured successfully';
    RAISE NOTICE '  - Compression: 7 days old data will be compressed';
    RAISE NOTICE '  - Retention: 30 days old data will be deleted';
END $$;

