-- 启用 TimescaleDB 扩展
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 输出版本信息
DO $$
DECLARE
    timescale_version TEXT;
BEGIN
    SELECT extversion INTO timescale_version 
    FROM pg_extension 
    WHERE extname = 'timescaledb';
    
    RAISE NOTICE 'TimescaleDB version: %', timescale_version;
END $$;

