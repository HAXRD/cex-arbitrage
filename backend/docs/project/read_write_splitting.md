# 读写分离配置文档

## 概述

本文档描述了数据存储层的读写分离实现，支持主从数据库架构，提高系统的可扩展性和读取性能。

## 架构设计

### 1. 基本架构

```
┌─────────────┐
│ Application │
└──────┬──────┘
       │
       │ (GORM + DBResolver)
       │
       ├─────────────┬─────────────┐
       │             │             │
   ┌───▼───┐    ┌───▼───┐    ┌───▼───┐
   │Primary│    │Replica│    │Replica│
   │ (写)  │    │ (读)  │    │ (读)  │
   └───────┘    └───────┘    └───────┘
```

### 2. 路由策略

- **写操作**: 所有的 `INSERT`, `UPDATE`, `DELETE` 操作自动路由到主库
- **读操作**: 所有的 `SELECT` 操作默认路由到从库（随机选择）
- **事务**: 事务中的所有操作（包括读）都路由到主库，保证一致性

## 实现细节

### 1. 核心组件

#### 1.1 配置结构

```go
// DatabaseConfig 数据库配置
type DatabaseConfig struct {
    Host            string          `mapstructure:"host"`
    Port            int             `mapstructure:"port"`
    User            string          `mapstructure:"user"`
    Password        string          `mapstructure:"password"`
    DBName          string          `mapstructure:"dbname"`
    SSLMode         string          `mapstructure:"sslmode"`
    MaxOpenConns    int             `mapstructure:"max_open_conns"`
    MaxIdleConns    int             `mapstructure:"max_idle_conns"`
    ConnMaxLifetime int             `mapstructure:"conn_max_lifetime"`
    ConnMaxIdleTime int             `mapstructure:"conn_max_idle_time"`
    Replicas        []ReplicaConfig `mapstructure:"replicas"` // 从库配置
}

// ReplicaConfig 从库配置
type ReplicaConfig struct {
    Host     string `mapstructure:"host"`
    Port     int    `mapstructure:"port"`
    User     string `mapstructure:"user"`
    Password string `mapstructure:"password"`
    DBName   string `mapstructure:"dbname"`
    SSLMode  string `mapstructure:"sslmode"`
}
```

#### 1.2 读写分离设置

使用 `gorm.io/plugin/dbresolver` 插件实现:

```go
func setupReadWriteSplitting(db *gorm.DB, cfg *config.DatabaseConfig, log *zap.Logger) error {
    if len(cfg.Replicas) == 0 {
        return nil
    }

    // 构建从库 DSN 列表
    replicaDSNs := make([]gorm.Dialector, 0, len(cfg.Replicas))
    for _, replica := range cfg.Replicas {
        dsn := replica.GetReplicaDSN()
        replicaDSNs = append(replicaDSNs, postgres.Open(dsn))
    }

    // 配置 DBResolver 插件
    err := db.Use(dbresolver.Register(dbresolver.Config{
        Replicas: replicaDSNs,
        Policy:   dbresolver.RandomPolicy{}, // 随机选择从库
    }).
        SetConnMaxIdleTime(time.Duration(cfg.ConnMaxIdleTime) * time.Second).
        SetConnMaxLifetime(time.Duration(cfg.ConnMaxLifetime) * time.Second).
        SetMaxIdleConns(cfg.MaxIdleConns).
        SetMaxOpenConns(cfg.MaxOpenConns))

    return err
}
```

### 2. 使用示例

#### 2.1 自动路由

```go
// 写操作 - 自动路由到主库
db.Create(&symbol)
db.Updates(&symbol)
db.Delete(&symbol)

// 读操作 - 自动路由到从库
db.First(&symbol)
db.Find(&symbols)
db.Count(&count)
```

#### 2.2 显式指定数据源

```go
import "gorm.io/plugin/dbresolver"

// 强制从主库读取（适用于需要立即一致性的场景）
db.Clauses(dbresolver.Write).First(&symbol)

// 显式从从库读取
db.Clauses(dbresolver.Read).Find(&symbols)
```

#### 2.3 事务中的操作

```go
// 事务中的所有操作（包括读）都在主库执行
db.Transaction(func(tx *gorm.DB) error {
    // 写操作
    if err := tx.Create(&symbol).Error; err != nil {
        return err
    }
    
    // 读操作（也在主库）
    var result Symbol
    if err := tx.Where("symbol = ?", symbol.Symbol).First(&result).Error; err != nil {
        return err
    }
    
    return nil
})
```

## 配置示例

### 开发环境配置 (config.yaml)

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: cryptosignal
  sslmode: disable
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  conn_max_idle_time: 600
  # 开发环境可以不配置从库
  replicas: []
```

### 生产环境配置 (config.yaml)

```yaml
database:
  host: primary.db.example.com
  port: 5432
  user: postgres
  password: ${DB_PASSWORD}
  dbname: cryptosignal
  sslmode: require
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 3600
  conn_max_idle_time: 600
  # 配置多个从库
  replicas:
    - host: replica1.db.example.com
      port: 5432
      user: postgres
      password: ${DB_PASSWORD}
      dbname: cryptosignal
      sslmode: require
    - host: replica2.db.example.com
      port: 5432
      user: postgres
      password: ${DB_PASSWORD}
      dbname: cryptosignal
      sslmode: require
```

## 监控和维护

### 1. 复制延迟监控

```go
// MonitorReplicationLag 监控复制延迟
func MonitorReplicationLag(ctx context.Context, db *gorm.DB, log *zap.Logger) error {
    var stats []ReplicationStat
    
    err := db.WithContext(ctx).
        Clauses(dbresolver.Write). // 在主库执行
        Raw(`
            SELECT 
                application_name,
                client_addr,
                state,
                replay_lag,
                flush_lag,
                write_lag
            FROM pg_stat_replication
        `).
        Scan(&stats).Error
    
    // 记录每个从库的延迟状态
    for _, stat := range stats {
        log.Info("Replication status",
            zap.String("application_name", stat.ApplicationName),
            zap.String("state", stat.State),
            zap.String("replay_lag", *stat.ReplayLag),
        )
    }
    
    return err
}
```

### 2. 复制状态检查

```go
// GetReplicationStatus 获取复制状态
func GetReplicationStatus(ctx context.Context, db *gorm.DB) (map[string]interface{}, error) {
    var info ReplicationInfo
    
    err := db.WithContext(ctx).
        Clauses(dbresolver.Write).
        Raw("SELECT COUNT(*) as slave_count FROM pg_stat_replication").
        Scan(&info).Error
    
    if err != nil {
        return nil, err
    }
    
    return map[string]interface{}{
        "slave_count": info.SlaveCount,
        "timestamp":   time.Now().UTC(),
    }, nil
}
```

## 测试覆盖

### 1. 单元测试

- ✅ 基本读写分离功能
- ✅ 默认路由行为
- ✅ 事务中的路由行为
- ✅ 批量操作的路由
- ✅ 强制指定数据源

### 2. 集成测试

- ✅ 多从库配置
- ✅ 无从库配置（降级为单库模式）
- ✅ 复制状态监控
- ✅ 复制延迟监控

## 性能优化建议

### 1. 从库数量

- 根据读写比例配置从库数量
- 建议 2-3 个从库即可满足大部分场景
- 从库过多会增加主从同步压力

### 2. 连接池配置

```go
// 主库连接池（写操作为主）
MaxOpenConns: 50
MaxIdleConns: 10

// 从库连接池（读操作较多）
MaxOpenConns: 100  // 可以适当增大
MaxIdleConns: 20
```

### 3. 负载均衡策略

当前实现使用 `RandomPolicy`（随机选择从库）:

```go
Policy: dbresolver.RandomPolicy{}
```

DBResolver 还支持其他策略:
- `RandomPolicy`: 随机选择
- 可以自定义策略实现 `LoadBalancer` 接口

## 故障处理

### 1. 从库故障

- DBResolver 会自动检测从库连接失败
- 自动降级到主库执行读操作
- 建议配置健康检查和告警

### 2. 主从延迟过大

- 对一致性要求高的查询使用 `dbresolver.Write` 强制主库读取
- 监控 `replay_lag` 指标
- 延迟超过阈值时触发告警

### 3. 主库故障转移

- 需要外部机制（如 Patroni、PgPool）处理主库故障转移
- 应用层需要重新连接到新主库

## 最佳实践

### 1. 什么时候使用读写分离？

✅ **适合场景**:
- 读操作远多于写操作（读写比 > 3:1）
- 需要横向扩展读能力
- 有明确的最终一致性要求

❌ **不适合场景**:
- 读写比例接近（< 2:1）
- 强一致性要求（所有读都需要最新数据）
- 数据量小，单库性能足够

### 2. 一致性权衡

```go
// 场景 1: 统计数据，可以容忍短暂延迟
// 使用从库读取，提高性能
count := db.Model(&Symbol{}).Count(&total)

// 场景 2: 写入后立即读取，需要强一致性
// 使用主库读取
db.Transaction(func(tx *gorm.DB) error {
    tx.Create(&symbol)  // 写入
    tx.First(&symbol)   // 立即读取（在同一事务，使用主库）
    return nil
})

// 或者显式指定主库
db.Clauses(dbresolver.Write).First(&symbol)
```

### 3. 监控指标

建议监控以下指标:
- 主从延迟（`replay_lag`）
- 主库 QPS（每秒查询数）
- 从库 QPS 分布
- 连接池使用率
- 查询响应时间

## 总结

✅ **已完成功能**:
1. 支持多从库配置
2. 自动读写分离路由
3. 随机负载均衡策略
4. 复制延迟监控
5. 复制状态查询
6. 完整的测试覆盖

🎯 **性能提升**:
- 读操作可分散到多个从库
- 减轻主库压力
- 提高系统整体吞吐量

⚠️ **注意事项**:
- 需要合理配置主从复制
- 注意一致性要求场景
- 监控主从延迟
- 做好故障预案

