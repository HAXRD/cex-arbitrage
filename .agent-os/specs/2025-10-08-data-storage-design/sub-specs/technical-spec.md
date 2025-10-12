# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-08-data-storage-design/spec.md

## Technical Requirements

### 1. PostgreSQL 数据库配置

#### 版本要求
- PostgreSQL 15+ with TimescaleDB 2.13+ 扩展
- 连接参数：
  - `max_connections`: 200（支持高并发）
  - `shared_buffers`: 256MB（内存缓冲）
  - `effective_cache_size`: 1GB
  - `work_mem`: 16MB

#### 字符集和时区
- 字符集：UTF-8
- 时区：UTC（所有时间戳统一使用 UTC）

### 2. TimescaleDB 时序优化配置

#### 超表配置
- **分片间隔（Chunk Time Interval）**：7天
  - `price_ticks` 表：7天分片
  - `klines` 表：7天分片
- **分片策略**：按 `timestamp` 字段自动分片

#### 数据压缩策略
```sql
-- 7天以上的数据自动压缩
ALTER TABLE price_ticks SET (
  timescaledb.compress,
  timescaledb.compress_segmentby = 'symbol',
  timescaledb.compress_orderby = 'timestamp DESC'
);

SELECT add_compression_policy('price_ticks', INTERVAL '7 days');
```

#### 数据保留策略
```sql
-- 自动删除30天以上的数据
SELECT add_retention_policy('price_ticks', INTERVAL '30 days');
SELECT add_retention_policy('klines', INTERVAL '30 days');
```

### 3. Redis 缓存配置

#### 版本要求
- Redis 7+
- 内存限制：512MB（初期）
- 淘汰策略：`allkeys-lru`（最近最少使用）

#### 持久化配置
- RDB：每5分钟保存一次（防止重启丢失数据）
- AOF：关闭（实时数据可从数据库恢复）

#### 缓存键设计规范
- **命名规则**：`namespace:type:identifier`
- **TTL 策略**：
  - 实时价格：60秒
  - 实时指标：60秒
  - 交易对列表：300秒（5分钟）
  - WebSocket 连接：心跳超时时间（90秒）

### 4. 数据访问层（DAO）技术要求

#### ORM 选择
- 使用 **GORM** v2（Go ORM 库）
- 优势：成熟稳定、支持批量操作、自动迁移、插件扩展

#### DAO 接口设计原则
- 每个表对应一个 DAO 接口
- 支持上下文（Context）传递，便于超时控制和请求追踪
- 所有数据库错误统一包装为自定义错误类型
- 返回值使用指针类型，避免大对象拷贝

#### 批量操作优化
- 批量插入：单次最多1000条
- 批量查询：使用 `IN` 查询，单次最多100个ID
- 分页查询：默认每页50条，最大200条

### 5. 读写分离架构

#### 主从配置
- **主库（Master）**：处理所有写操作（INSERT、UPDATE、DELETE）
- **从库（Slave）**：处理所有读操作（SELECT）
- **复制延迟监控**：最大允许延迟5秒

#### GORM 读写分离配置
```go
import "gorm.io/plugin/dbresolver"

db.Use(dbresolver.Register(dbresolver.Config{
    Sources:  []gorm.Dialector{postgres.Open(masterDSN)},  // 主库
    Replicas: []gorm.Dialector{postgres.Open(slaveDSN)},   // 从库
    Policy:   dbresolver.RandomPolicy{},                    // 随机选择从库
}))
```

### 6. 连接池配置

#### PostgreSQL 连接池（GORM）
```go
sqlDB, _ := db.DB()
sqlDB.SetMaxOpenConns(100)          // 最大打开连接数
sqlDB.SetMaxIdleConns(10)           // 最大空闲连接数
sqlDB.SetConnMaxLifetime(time.Hour) // 连接最大生命周期
sqlDB.SetConnMaxIdleTime(10 * time.Minute) // 空闲连接超时
```

#### Redis 连接池（go-redis）
```go
redis.NewClient(&redis.Options{
    PoolSize:     50,                // 连接池大小
    MinIdleConns: 10,                // 最小空闲连接
    MaxRetries:   3,                 // 最大重试次数
    PoolTimeout:  4 * time.Second,   // 连接池超时
    IdleTimeout:  5 * time.Minute,   // 空闲连接超时
})
```

### 7. 数据一致性保证

#### 写入策略
- **先写数据库，后写缓存**：确保数据持久化优先
- **异步批量写入**：实时 Ticker 数据先缓存，每5秒批量写入数据库
- **事务支持**：关键操作使用数据库事务

#### 缓存更新策略
- **Cache-Aside 模式**：读取时先查缓存，缓存未命中再查数据库
- **主动更新**：数据写入数据库后，主动更新 Redis 缓存
- **缓存穿透保护**：使用布隆过滤器防止查询不存在的数据

### 8. 性能指标要求

#### 数据库性能
- 单表查询响应时间：< 100ms（90分位）
- 批量插入性能：> 5000 条/秒
- 时间范围查询：查询1天K线数据 < 200ms

#### 缓存性能
- Redis 读写延迟：< 5ms（平均）
- 缓存命中率：> 95%
- 并发连接数：支持1000+并发

#### 连接池性能
- 连接获取时间：< 10ms
- 连接池利用率：60-80%（避免过度占用）

### 9. 监控和日志

#### 数据库监控指标
- 慢查询日志：记录超过100ms的查询
- 连接池状态：打开连接数、空闲连接数
- 数据库大小：每日统计表空间占用

#### 缓存监控指标
- 内存使用率
- 缓存命中率
- 键空间统计

#### 日志记录
- 所有 DAO 操作记录结构化日志（使用 Zap）
- 包含：操作类型、表名、耗时、影响行数、错误信息
- 慢查询单独记录 WARN 级别日志

### 10. 错误处理

#### 自定义错误类型
```go
type DBError struct {
    Op        string    // 操作类型（SELECT、INSERT等）
    Table     string    // 表名
    Err       error     // 原始错误
    Timestamp time.Time // 错误时间
}
```

#### 错误分类
- **网络错误**：连接超时、连接被拒绝 → 自动重试
- **数据错误**：唯一键冲突、外键约束 → 返回业务错误
- **系统错误**：磁盘满、内存不足 → 告警并降级

### 11. 数据迁移和版本管理

#### 迁移工具
- 使用 **golang-migrate** 管理数据库迁移
- 迁移文件命名：`YYYYMMDDHHMMSS_description.up.sql` / `.down.sql`
- 支持向上迁移和回滚

#### 版本控制
- 数据库表版本通过迁移文件管理
- 每次表结构变更必须创建新的迁移文件
- 生产环境迁移需要审批和备份

## External Dependencies

### 新增依赖包

1. **gorm.io/gorm** - Go ORM 库
   - **版本**：v1.25.5
   - **用途**：数据库操作封装、模型定义、迁移管理
   - **Justification**：成熟稳定的 ORM 库，支持 PostgreSQL 和读写分离，简化数据库操作代码

2. **gorm.io/driver/postgres** - GORM PostgreSQL 驱动
   - **版本**：v1.5.4
   - **用途**：GORM 的 PostgreSQL 适配器
   - **Justification**：GORM 官方支持的 PostgreSQL 驱动

3. **gorm.io/plugin/dbresolver** - GORM 读写分离插件
   - **版本**：v1.4.7
   - **用途**：实现主从读写分离
   - **Justification**：GORM 官方插件，配置简单，支持多从库负载均衡

4. **github.com/redis/go-redis/v9** - Redis 客户端
   - **版本**：v9.3.0
   - **用途**：Redis 连接、缓存操作、连接池管理
   - **Justification**：Go 社区最流行的 Redis 客户端，支持 Redis 7+ 新特性

5. **github.com/golang-migrate/migrate/v4** - 数据库迁移工具
   - **版本**：v4.16.2
   - **用途**：数据库版本管理和迁移
   - **Justification**：支持多种数据库，提供 CLI 和 Go API，生产环境广泛使用

6. **github.com/lib/pq** - PostgreSQL 底层驱动（GORM 依赖）
   - **版本**：v1.10.9
   - **用途**：PostgreSQL 连接底层实现
   - **Justification**：Pure Go 实现，无需 CGO，稳定可靠
