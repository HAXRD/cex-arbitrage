-- 测试数据生成脚本
-- 用于测试 TimescaleDB 压缩和保留策略

-- ========================================
-- 生成 price_ticks 测试数据
-- ========================================

-- 插入8天前的数据（应该被压缩）
INSERT INTO price_ticks (symbol, timestamp, last_price, base_volume, quote_volume)
SELECT 
    'TEST_' || (i % 5)::text,                                    -- 5个测试交易对
    NOW() - INTERVAL '8 days' + (i || ' minutes')::interval,     -- 8天前开始
    40000 + (random() * 5000)::numeric(20,8),                    -- 随机价格 40000-45000
    (random() * 100)::numeric(30,8),                             -- 随机交易量
    (random() * 4000000)::numeric(30,8)                          -- 随机计价币交易量
FROM generate_series(1, 1000) i;

-- 插入6天前的数据（未达到压缩时间）
INSERT INTO price_ticks (symbol, timestamp, last_price, base_volume, quote_volume)
SELECT 
    'TEST_' || (i % 5)::text,
    NOW() - INTERVAL '6 days' + (i || ' minutes')::interval,
    40000 + (random() * 5000)::numeric(20,8),
    (random() * 100)::numeric(30,8),
    (random() * 4000000)::numeric(30,8)
FROM generate_series(1, 1000) i;

-- 插入最近的数据
INSERT INTO price_ticks (symbol, timestamp, last_price, base_volume, quote_volume)
SELECT 
    'TEST_' || (i % 5)::text,
    NOW() - (i || ' minutes')::interval,
    40000 + (random() * 5000)::numeric(20,8),
    (random() * 100)::numeric(30,8),
    (random() * 4000000)::numeric(30,8)
FROM generate_series(1, 500) i;

-- ========================================
-- 生成 klines 测试数据
-- ========================================

-- 插入8天前的K线数据（应该被压缩）
INSERT INTO klines (symbol, timestamp, granularity, open, high, low, close, base_volume, quote_volume)
SELECT 
    'TEST_' || (i % 5)::text,
    NOW() - INTERVAL '8 days' + (i || ' minutes')::interval,
    '1m',
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    (random() * 100)::numeric(30,8),
    (random() * 4000000)::numeric(30,8)
FROM generate_series(1, 1000) i;

-- 插入6天前的K线数据（未达到压缩时间）
INSERT INTO klines (symbol, timestamp, granularity, open, high, low, close, base_volume, quote_volume)
SELECT 
    'TEST_' || (i % 5)::text,
    NOW() - INTERVAL '6 days' + (i || ' minutes')::interval,
    '1m',
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    (random() * 100)::numeric(30,8),
    (random() * 4000000)::numeric(30,8)
FROM generate_series(1, 1000) i;

-- 插入最近的K线数据
INSERT INTO klines (symbol, timestamp, granularity, open, high, low, close, base_volume, quote_volume)
SELECT 
    'TEST_' || (i % 5)::text,
    NOW() - (i || ' minutes')::interval,
    '1m',
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    40000 + (random() * 5000)::numeric(20,8),
    (random() * 100)::numeric(30,8),
    (random() * 4000000)::numeric(30,8)
FROM generate_series(1, 500) i;

-- ========================================
-- 验证数据插入
-- ========================================

DO $$
DECLARE
    price_ticks_count INT;
    klines_count INT;
    price_ticks_chunks INT;
    klines_chunks INT;
BEGIN
    SELECT COUNT(*) INTO price_ticks_count FROM price_ticks;
    SELECT COUNT(*) INTO klines_count FROM klines;
    
    SELECT COUNT(*) INTO price_ticks_chunks 
    FROM timescaledb_information.chunks 
    WHERE hypertable_name = 'price_ticks';
    
    SELECT COUNT(*) INTO klines_chunks 
    FROM timescaledb_information.chunks 
    WHERE hypertable_name = 'klines';
    
    RAISE NOTICE '✓ Test data generated successfully';
    RAISE NOTICE '  - price_ticks: % rows, % chunks', price_ticks_count, price_ticks_chunks;
    RAISE NOTICE '  - klines: % rows, % chunks', klines_count, klines_chunks;
    RAISE NOTICE '';
    RAISE NOTICE 'To test compression:';
    RAISE NOTICE '  1. Run: CALL run_job(<job_id>); -- for compression policy';
    RAISE NOTICE '  2. Check: SELECT * FROM timescaledb_information.chunks WHERE is_compressed = true;';
END $$;

