# 故障排查指南

## 概述

本文档提供了 CryptoSignal Hunter 系统的故障排查指南，帮助快速定位和解决常见问题。

## 系统架构检查

### 1. 服务状态检查

#### 检查所有服务状态
```bash
# 检查Docker服务状态
docker-compose ps

# 检查服务健康状态
docker-compose exec postgres pg_isready -U postgres
docker-compose exec redis redis-cli ping
```

#### 检查端口占用
```bash
# 检查关键端口
netstat -tlnp | grep -E ":(5432|6379|8080|3000)"

# 检查进程
ps aux | grep -E "(postgres|redis|go|node)"
```

### 2. 日志检查

#### 应用日志
```bash
# 后端日志
docker-compose logs -f backend

# 数据库日志
docker-compose logs -f postgres

# Redis日志
docker-compose logs -f redis
```

#### 系统日志
```bash
# 系统日志
journalctl -u docker
dmesg | tail -50

# 网络日志
ss -tuln
```

## 数据库故障排查

### 1. 连接问题

#### 问题：无法连接数据库
```bash
# 症状
ERROR: connection refused
FATAL: password authentication failed

# 诊断步骤
1. 检查PostgreSQL服务状态
docker-compose exec postgres pg_ctl status

2. 检查连接配置
docker-compose exec postgres psql -U postgres -c "SELECT version();"

3. 检查网络连接
telnet localhost 5432

# 解决方案
1. 重启PostgreSQL服务
docker-compose restart postgres

2. 检查密码配置
docker-compose exec postgres psql -U postgres -c "ALTER USER postgres PASSWORD 'password';"

3. 检查防火墙设置
sudo ufw allow 5432
```

#### 问题：连接池耗尽
```bash
# 症状
ERROR: sorry, too many clients already
FATAL: remaining connection slots are reserved

# 诊断步骤
1. 检查当前连接数
docker-compose exec postgres psql -U postgres -c "SELECT count(*) FROM pg_stat_activity;"

2. 检查最大连接数
docker-compose exec postgres psql -U postgres -c "SHOW max_connections;"

# 解决方案
1. 增加最大连接数
docker-compose exec postgres psql -U postgres -c "ALTER SYSTEM SET max_connections = 200;"

2. 优化连接池配置
# 在应用配置中调整
MaxOpenConns: 50
MaxIdleConns: 10
```

### 2. 迁移问题

#### 问题：迁移失败
```bash
# 症状
ERROR: migration failed
FATAL: extension "timescaledb" does not exist

# 诊断步骤
1. 检查迁移状态
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" version

2. 检查TimescaleDB扩展
docker-compose exec postgres psql -U postgres -d cryptosignal -c "SELECT * FROM pg_extension WHERE extname = 'timescaledb';"

# 解决方案
1. 重新安装TimescaleDB
docker-compose down
docker-compose build --no-cache postgres
docker-compose up -d postgres

2. 手动安装扩展
docker-compose exec postgres psql -U postgres -d cryptosignal -c "CREATE EXTENSION IF NOT EXISTS timescaledb;"

3. 强制回滚迁移
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" force 0
```

#### 问题：迁移版本不一致
```bash
# 症状
ERROR: migration version mismatch
FATAL: dirty database version

# 诊断步骤
1. 检查迁移历史
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" version

2. 检查schema_migrations表
docker-compose exec postgres psql -U postgres -d cryptosignal -c "SELECT * FROM schema_migrations;"

# 解决方案
1. 强制设置版本
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" force 4

2. 清理迁移状态
docker-compose exec postgres psql -U postgres -d cryptosignal -c "DELETE FROM schema_migrations WHERE dirty = true;"
```

### 3. 性能问题

#### 问题：查询慢
```bash
# 症状
查询响应时间 > 1秒
数据库CPU使用率高

# 诊断步骤
1. 检查慢查询
docker-compose exec postgres psql -U postgres -d cryptosignal -c "SELECT query, mean_time, calls FROM pg_stat_statements ORDER BY mean_time DESC LIMIT 10;"

2. 检查索引使用
docker-compose exec postgres psql -U postgres -d cryptosignal -c "SELECT schemaname, tablename, indexname, idx_scan, idx_tup_read FROM pg_stat_user_indexes;"

# 解决方案
1. 创建缺失索引
CREATE INDEX idx_price_ticks_symbol_timestamp ON price_ticks (symbol, timestamp DESC);

2. 优化查询语句
EXPLAIN ANALYZE SELECT * FROM price_ticks WHERE symbol = 'BTCUSDT' AND timestamp > NOW() - INTERVAL '1 hour';
```

#### 问题：写入性能差
```bash
# 症状
写入延迟高
数据库锁等待

# 诊断步骤
1. 检查锁等待
docker-compose exec postgres psql -U postgres -d cryptosignal -c "SELECT * FROM pg_locks WHERE NOT granted;"

2. 检查WAL配置
docker-compose exec postgres psql -U postgres -d cryptosignal -c "SHOW wal_level;"

# 解决方案
1. 优化WAL配置
ALTER SYSTEM SET wal_level = replica;
ALTER SYSTEM SET max_wal_size = '2GB';
ALTER SYSTEM SET checkpoint_completion_target = 0.9;

2. 使用批量插入
INSERT INTO price_ticks (symbol, bid_price, ask_price, timestamp) VALUES ...;
```

## Redis故障排查

### 1. 连接问题

#### 问题：Redis连接失败
```bash
# 症状
ERROR: connection refused
FATAL: redis: connection pool exhausted

# 诊断步骤
1. 检查Redis服务状态
docker-compose exec redis redis-cli ping

2. 检查连接配置
docker-compose exec redis redis-cli info clients

# 解决方案
1. 重启Redis服务
docker-compose restart redis

2. 检查网络配置
telnet localhost 6379
```

#### 问题：内存不足
```bash
# 症状
ERROR: OOM command not allowed
FATAL: memory limit exceeded

# 诊断步骤
1. 检查内存使用
docker-compose exec redis redis-cli info memory

2. 检查键数量
docker-compose exec redis redis-cli dbsize

# 解决方案
1. 增加内存限制
# 在docker-compose.yml中调整
redis:
  command: redis-server --maxmemory 2gb --maxmemory-policy allkeys-lru

2. 清理过期键
docker-compose exec redis redis-cli --scan --pattern "*" | xargs redis-cli del
```

### 2. 性能问题

#### 问题：缓存命中率低
```bash
# 症状
缓存命中率 < 80%
大量缓存未命中

# 诊断步骤
1. 检查缓存统计
docker-compose exec redis redis-cli info stats

2. 检查键过期
docker-compose exec redis redis-cli --scan --pattern "*" | head -10 | xargs -I {} redis-cli ttl {}

# 解决方案
1. 调整TTL策略
# 增加缓存时间
TTLRealTimePrice = 300 * time.Second

2. 优化缓存键设计
# 使用更合理的键名
key := fmt.Sprintf("price:%s:%d", symbol, timestamp.Unix()/60)
```

#### 问题：Redis响应慢
```bash
# 症状
Redis延迟 > 10ms
客户端超时

# 诊断步骤
1. 检查慢查询
docker-compose exec redis redis-cli slowlog get 10

2. 检查内存碎片
docker-compose exec redis redis-cli info memory | grep fragmentation

# 解决方案
1. 优化数据结构
# 使用Hash代替String
HSET price:BTCUSDT bid 50000 ask 50010

2. 使用Pipeline
# 批量操作减少网络往返
pipe := client.Pipeline()
for _, key := range keys {
    pipe.Get(key)
}
pipe.Exec()
```

## 应用故障排查

### 1. 启动问题

#### 问题：应用启动失败
```bash
# 症状
FATAL: failed to start server
ERROR: configuration error

# 诊断步骤
1. 检查配置文件
cat backend/config.yaml

2. 检查环境变量
env | grep -E "(DB_|REDIS_)"

3. 检查依赖服务
docker-compose ps

# 解决方案
1. 修复配置文件
# 检查YAML语法
yamllint backend/config.yaml

2. 设置环境变量
export DB_HOST=localhost
export DB_PORT=5432
export REDIS_HOST=localhost
export REDIS_PORT=6379
```

#### 问题：依赖注入失败
```bash
# 症状
ERROR: dependency injection failed
FATAL: service not found

# 诊断步骤
1. 检查服务注册
grep -r "NewService" backend/internal/

2. 检查依赖关系
go mod graph | grep -E "(database|cache)"

# 解决方案
1. 检查导入路径
import "github.com/haxrd/cryptosignal-hunter/internal/database"

2. 重新构建
go mod tidy
go build ./cmd/server
```

### 2. 运行时问题

#### 问题：内存泄漏
```bash
# 症状
内存使用持续增长
应用响应变慢

# 诊断步骤
1. 检查内存使用
go tool pprof http://localhost:6060/debug/pprof/heap

2. 检查goroutine泄漏
go tool pprof http://localhost:6060/debug/pprof/goroutine

# 解决方案
1. 修复资源泄漏
defer conn.Close()
defer file.Close()

2. 使用对象池
var pool = sync.Pool{New: func() interface{} { return &Object{} }}
```

#### 问题：死锁
```bash
# 症状
应用无响应
goroutine阻塞

# 诊断步骤
1. 检查goroutine状态
go tool pprof http://localhost:6060/debug/pprof/goroutine

2. 检查锁竞争
go tool pprof http://localhost:6060/debug/pprof/mutex

# 解决方案
1. 避免嵌套锁
// 错误
mu1.Lock()
mu2.Lock()
// 正确
mu1.Lock()
// 处理逻辑
mu1.Unlock()
mu2.Lock()
```

### 3. 网络问题

#### 问题：API请求失败
```bash
# 症状
HTTP 500 Internal Server Error
连接超时

# 诊断步骤
1. 检查API端点
curl -v http://localhost:8080/health

2. 检查日志
tail -f backend/logs/app.log

# 解决方案
1. 检查路由配置
// 确保路由正确注册
r.GET("/health", healthHandler)

2. 检查中间件
// 确保中间件正确配置
r.Use(middleware.Logger())
```

## 监控和诊断

### 1. 系统监控

#### 资源监控
```bash
# CPU使用率
top -p $(pgrep -f "cryptosignal")

# 内存使用
ps aux | grep cryptosignal

# 磁盘使用
df -h
du -sh /var/lib/docker/volumes/
```

#### 网络监控
```bash
# 网络连接
netstat -tlnp | grep -E ":(5432|6379|8080)"

# 网络延迟
ping -c 5 localhost
```

### 2. 应用监控

#### 性能监控
```bash
# 使用pprof
go tool pprof http://localhost:6060/debug/pprof/profile

# 内存分析
go tool pprof http://localhost:6060/debug/pprof/heap

# 阻塞分析
go tool pprof http://localhost:6060/debug/pprof/block
```

#### 日志分析
```bash
# 错误日志
grep "ERROR" backend/logs/app.log | tail -20

# 慢查询日志
grep "SLOW" backend/logs/app.log | tail -20

# 统计日志
grep "STATS" backend/logs/app.log | tail -20
```

## 常见问题解决方案

### 1. 数据库问题

#### 问题：表不存在
```sql
-- 检查表是否存在
SELECT table_name FROM information_schema.tables WHERE table_schema = 'public';

-- 重新创建表
migrate -path ./migrations -database "postgres://postgres:password@localhost:5432/cryptosignal?sslmode=disable" up
```

#### 问题：权限不足
```sql
-- 检查用户权限
SELECT * FROM pg_user WHERE usename = 'postgres';

-- 授予权限
GRANT ALL PRIVILEGES ON DATABASE cryptosignal TO postgres;
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO postgres;
```

### 2. 缓存问题

#### 问题：键不存在
```bash
# 检查键是否存在
docker-compose exec redis redis-cli exists "cryptosignal:latest_price:BTCUSDT"

# 检查键过期时间
docker-compose exec redis redis-cli ttl "cryptosignal:latest_price:BTCUSDT"
```

#### 问题：数据类型错误
```bash
# 检查键类型
docker-compose exec redis redis-cli type "cryptosignal:latest_price:BTCUSDT"

# 删除错误类型的键
docker-compose exec redis redis-cli del "cryptosignal:latest_price:BTCUSDT"
```

### 3. 应用问题

#### 问题：配置错误
```bash
# 检查配置文件
cat backend/config.yaml

# 验证配置
go run ./cmd/server --config-check
```

#### 问题：端口冲突
```bash
# 查找占用端口的进程
lsof -i :8080

# 终止进程
kill -9 <PID>

# 或修改配置使用其他端口
```

## 预防措施

### 1. 监控告警

#### 设置监控指标
```yaml
# prometheus.yml
scrape_configs:
  - job_name: 'cryptosignal'
    static_configs:
      - targets: ['localhost:8080']
    metrics_path: '/metrics'
    scrape_interval: 15s
```

#### 配置告警规则
```yaml
# alert.rules
groups:
  - name: cryptosignal
    rules:
      - alert: HighErrorRate
        expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High error rate detected"
```

### 2. 备份策略

#### 数据库备份
```bash
# 创建备份
pg_dump -h localhost -U postgres -d cryptosignal > backup_$(date +%Y%m%d_%H%M%S).sql

# 恢复备份
psql -h localhost -U postgres -d cryptosignal < backup_20240101_120000.sql
```

#### 配置备份
```bash
# 备份配置文件
cp backend/config.yaml config_backup_$(date +%Y%m%d).yaml

# 备份迁移文件
tar -czf migrations_backup_$(date +%Y%m%d).tar.gz backend/migrations/
```

### 3. 测试验证

#### 健康检查
```bash
# 数据库健康检查
curl http://localhost:8080/health/db

# Redis健康检查
curl http://localhost:8080/health/redis

# 应用健康检查
curl http://localhost:8080/health
```

#### 功能测试
```bash
# 运行集成测试
go test -v ./integration_test.go

# 运行性能测试
go test -bench=. ./benchmark_test.go
```

## 联系支持

### 1. 收集信息

在寻求帮助时，请提供以下信息：

1. **系统信息**
   - 操作系统版本
   - Docker版本
   - Go版本

2. **错误信息**
   - 完整的错误日志
   - 错误发生的时间
   - 错误发生前的操作

3. **配置信息**
   - 配置文件内容
   - 环境变量设置
   - 服务状态

### 2. 日志收集

```bash
# 收集系统日志
journalctl -u docker > system.log

# 收集应用日志
docker-compose logs > app.log

# 收集配置信息
cat backend/config.yaml > config.yaml
env > environment.txt
```

### 3. 问题报告

请按照以下格式报告问题：

```
**问题描述**
简要描述问题

**重现步骤**
1. 执行的操作
2. 期望的结果
3. 实际的结果

**环境信息**
- 操作系统: Ubuntu 20.04
- Docker版本: 20.10.7
- Go版本: 1.21.0

**错误日志**
[粘贴相关错误日志]

**配置文件**
[粘贴相关配置文件]
```
