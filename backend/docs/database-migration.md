# 数据库迁移使用文档

## 概述

本项目使用 `golang-migrate` 工具进行数据库版本管理，支持 PostgreSQL 和 TimescaleDB 扩展。

## 环境要求

- PostgreSQL 15+
- TimescaleDB 2.13+
- golang-migrate 4.16.2+

## 安装 golang-migrate

### macOS
```bash
brew install golang-migrate
```

### Linux
```bash
curl -L https://packagecloud.io/golang-migrate/migrate/gpgkey | apt-key add -
echo "deb https://packagecloud.io/golang-migrate/migrate/ubuntu/ $(lsb_release -sc) main" > /etc/apt/sources.list.d/migrate.list
apt-get update
apt-get install -y migrate
```

### Windows
```bash
choco install golang-migrate
```

## 配置环境变量

创建 `.env` 文件或设置环境变量：

```bash
# 数据库连接配置
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=cryptosignal
export DB_USER=postgres
export DB_PASSWORD=your_password
export DB_SSLMODE=disable

# 迁移文件路径
export MIGRATIONS_PATH=./migrations
```

## 基本命令

### 查看迁移状态
```bash
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" version
```

### 执行迁移（升级到最新版本）
```bash
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" up
```

### 回滚迁移（回滚到上一个版本）
```bash
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" down 1
```

### 回滚到指定版本
```bash
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" goto 2
```

### 强制设置版本（谨慎使用）
```bash
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" force 3
```

## 迁移文件说明

### 000001_enable_timescaledb
- **功能**: 启用 TimescaleDB 扩展
- **依赖**: 需要 TimescaleDB 已安装
- **回滚**: 禁用 TimescaleDB 扩展

### 000002_create_symbols_table
- **功能**: 创建 symbols 表
- **包含**: 所有 BitGet API 字段，索引配置
- **回滚**: 删除 symbols 表

### 000003_create_timeseries_tables
- **功能**: 创建时序表（price_ticks, klines）
- **包含**: TimescaleDB 超表配置，7天分片
- **回滚**: 删除时序表

### 000004_configure_timescaledb_policies
- **功能**: 配置压缩和保留策略
- **包含**: 7天压缩策略，30天保留策略
- **回滚**: 删除所有策略

## 开发环境快速开始

### 1. 启动数据库
```bash
# 使用 Docker Compose
docker-compose up -d postgres redis
```

### 2. 执行迁移
```bash
# 进入 backend 目录
cd backend

# 执行所有迁移
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" up
```

### 3. 验证迁移
```bash
# 检查当前版本
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" version

# 验证 TimescaleDB 扩展
psql -h localhost -U postgres -d cryptosignal -c "SELECT * FROM timescaledb_information.hypertables;"
```

## 生产环境部署

### 1. 备份现有数据
```bash
pg_dump -h localhost -U postgres -d cryptosignal > backup_$(date +%Y%m%d_%H%M%S).sql
```

### 2. 执行迁移
```bash
# 在生产环境执行迁移
migrate -path ./migrations -database "postgres://postgres:password@prod-host:5432/cryptosignal?sslmode=require" up
```

### 3. 验证迁移结果
```bash
# 检查表结构
psql -h prod-host -U postgres -d cryptosignal -c "\dt"

# 检查 TimescaleDB 配置
psql -h prod-host -U postgres -d cryptosignal -c "SELECT * FROM timescaledb_information.jobs;"
```

## 故障排查

### 迁移失败
```bash
# 查看迁移历史
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" version

# 强制设置版本（谨慎使用）
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" force 3
```

### 连接问题
```bash
# 测试数据库连接
psql -h localhost -U postgres -d cryptosignal -c "SELECT version();"

# 检查 TimescaleDB 扩展
psql -h localhost -U postgres -d cryptosignal -c "SELECT * FROM pg_extension WHERE extname = 'timescaledb';"
```

### 权限问题
```bash
# 确保用户有足够权限
psql -h localhost -U postgres -d cryptosignal -c "GRANT ALL PRIVILEGES ON DATABASE cryptosignal TO postgres;"
```

## 最佳实践

1. **备份**: 执行迁移前务必备份数据
2. **测试**: 在测试环境先验证迁移
3. **版本控制**: 迁移文件必须纳入版本控制
4. **回滚计划**: 准备回滚方案
5. **监控**: 监控迁移执行过程

## 常用脚本

### 检查迁移状态脚本
```bash
#!/bin/bash
# check_migration_status.sh

DB_URL="postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable"
MIGRATIONS_PATH="./migrations"

echo "当前迁移版本:"
migrate -path $MIGRATIONS_PATH -database $DB_URL version

echo -e "\n迁移历史:"
migrate -path $MIGRATIONS_PATH -database $DB_URL version
```

### 自动迁移脚本
```bash
#!/bin/bash
# auto_migrate.sh

DB_URL="postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable"
MIGRATIONS_PATH="./migrations"

echo "开始执行数据库迁移..."
migrate -path $MIGRATIONS_PATH -database $DB_URL up

if [ $? -eq 0 ]; then
    echo "迁移执行成功"
    migrate -path $MIGRATIONS_PATH -database $DB_URL version
else
    echo "迁移执行失败"
    exit 1
fi
```

## 注意事项

1. **TimescaleDB 依赖**: 确保 TimescaleDB 扩展已正确安装
2. **权限要求**: 需要 CREATE EXTENSION 权限
3. **数据备份**: 生产环境迁移前务必备份
4. **版本兼容**: 确保迁移文件与数据库版本兼容
5. **回滚测试**: 定期测试回滚功能
