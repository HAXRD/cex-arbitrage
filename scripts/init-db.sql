-- 数据库初始化脚本
-- CryptoSignal Hunter - PostgreSQL + TimescaleDB 初始化

-- 创建数据库（如果不存在）
-- 注意：在 Docker 初始化时，POSTGRES_DB 已经创建了数据库

-- 设置时区为 UTC
SET timezone = 'UTC';

-- 输出初始化信息
DO $$
BEGIN
    RAISE NOTICE 'Database initialization started';
    RAISE NOTICE 'Current database: %', current_database();
    RAISE NOTICE 'Current user: %', current_user();
    RAISE NOTICE 'Timezone: %', current_setting('timezone');
END $$;

-- 数据库迁移将在应用启动时通过 golang-migrate 执行
-- 这个脚本只负责基础初始化
