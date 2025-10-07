-- ============================================
-- CryptoSignal Hunter 数据库初始化脚本
-- 用途：启用TimescaleDB扩展和基础配置
-- ============================================

-- 启用TimescaleDB扩展
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 验证扩展是否安装成功
SELECT extname, extversion 
FROM pg_extension 
WHERE extname = 'timescaledb';

-- 设置默认时区为UTC
SET timezone = 'UTC';

-- 输出确认信息
DO $$
BEGIN
  RAISE NOTICE 'TimescaleDB 初始化完成！';
  RAISE NOTICE '数据库已就绪，可以开始开发。';
END $$;

