# Database Schema

This is the database schema implementation for the spec detailed in @.agent-os/specs/2025-10-07-project-initialization/spec.md

## 说明

在项目初始化阶段，**暂不创建具体的业务数据表**。本阶段仅负责：

1. 启用 TimescaleDB 扩展
2. 验证数据库连接正常
3. 为后续功能开发做好准备

具体的数据库表结构（如 `symbols`、`price_ticks`、`klines` 等）将在后续的"数据存储设计"规范中详细定义和实现。

## 初始化脚本

**文件位置：** `scripts/init-db.sql`

```sql
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

-- 创建应用专用Schema（可选，保持使用public即可）
-- CREATE SCHEMA IF NOT EXISTS cryptosignal;

-- 设置默认时区为UTC
SET timezone = 'UTC';

-- 输出确认信息
DO $$
BEGIN
  RAISE NOTICE 'TimescaleDB 初始化完成！';
  RAISE NOTICE '数据库已就绪，可以开始开发。';
END $$;
```

## Docker Compose集成

在 `docker-compose.yml` 中，PostgreSQL服务需要挂载初始化脚本：

```yaml
services:
  postgres:
    image: timescale/timescaledb:latest-pg15
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init.sql
```

**工作原理：**
- Docker容器首次启动时，会自动执行 `/docker-entrypoint-initdb.d/` 目录下的所有 `.sql` 和 `.sh` 文件
- `init-db.sql` 会在数据库创建后自动运行
- 后续启动不会重复执行（仅首次运行）

## 验证步骤

开发者可通过以下命令验证TimescaleDB是否正确安装：

```bash
# 进入PostgreSQL容器
docker exec -it <container_name> psql -U postgres -d cryptosignal

# 执行验证查询
SELECT extname, extversion FROM pg_extension WHERE extname = 'timescaledb';

# 预期输出：
#    extname    | extversion
# --------------+------------
#  timescaledb  | 2.x.x
```

## 数据库连接信息

**开发环境默认配置：**
- Host: `localhost`
- Port: `5432`
- Database: `cryptosignal`
- User: `postgres`
- Password: `postgres`
- SSLMode: `disable`

**连接字符串示例（Go）：**
```go
dsn := "host=localhost port=5432 user=postgres password=postgres dbname=cryptosignal sslmode=disable"
```

## 未来扩展

后续规范中将创建以下表结构（预告）：

- `symbols` - 交易对配置表
- `price_ticks` - 实时价格数据（TimescaleDB hypertable）
- `klines` - K线数据（TimescaleDB hypertable）
- `strategies` - 用户策略配置表
- `backtest_results` - 回测结果表

这些表的详细设计将在"数据存储设计"规范中说明。

