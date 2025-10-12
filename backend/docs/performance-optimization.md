# 性能优化建议文档

## 概述

本文档提供了 CryptoSignal Hunter 系统的性能优化建议，涵盖数据库、缓存、应用层和系统级的优化策略。

## 数据库性能优化

### 1. PostgreSQL 配置优化

#### 内存配置
```sql
-- 共享缓冲区（建议设置为系统内存的25%）
shared_buffers = '2GB'

-- 工作内存（用于排序和哈希操作）
work_mem = '64MB'

-- 维护工作内存（用于VACUUM、CREATE INDEX等）
maintenance_work_mem = '512MB'

-- 有效缓存大小（用于查询规划器）
effective_cache_size = '6GB'
```

#### 连接配置
```sql
-- 最大连接数
max_connections = 200

-- 共享预加载库
shared_preload_libraries = 'timescaledb'

-- 日志配置
log_statement = 'mod'
log_min_duration_statement = 1000
log_line_prefix = '%t [%p]: [%l-1] user=%u,db=%d,app=%a,client=%h '
```

### 2. TimescaleDB 优化

#### 分片策略优化
```sql
-- 根据数据量调整分片大小
-- 高频数据：1天分片
SELECT create_hypertable('price_ticks', 'timestamp', 
    chunk_time_interval => INTERVAL '1 day');

-- 低频数据：7天分片
SELECT create_hypertable('klines', 'timestamp',
    chunk_time_interval => INTERVAL '7 days');
```

#### 压缩策略优化
```sql
-- 根据查询模式调整压缩时间
-- 实时查询：3天后压缩
SELECT add_compression_policy('price_ticks', INTERVAL '3 days');

-- 历史分析：7天后压缩
SELECT add_compression_policy('klines', INTERVAL '7 days');
```

#### 索引优化
```sql
-- 创建复合索引
CREATE INDEX idx_price_ticks_symbol_time ON price_ticks (symbol, timestamp DESC);
CREATE INDEX idx_klines_symbol_granularity_time ON klines (symbol, granularity, timestamp DESC);

-- 部分索引（只对活跃数据建索引）
CREATE INDEX idx_price_ticks_active ON price_ticks (symbol, timestamp DESC) 
WHERE timestamp > NOW() - INTERVAL '7 days';
```

### 3. 查询优化

#### 时间范围查询优化
```sql
-- 使用时间范围限制
SELECT * FROM price_ticks 
WHERE symbol = 'BTCUSDT' 
  AND timestamp >= NOW() - INTERVAL '1 hour'
  AND timestamp <= NOW()
ORDER BY timestamp DESC
LIMIT 100;

-- 使用分页查询
SELECT * FROM klines 
WHERE symbol = 'BTCUSDT' 
  AND granularity = '1h'
  AND timestamp >= '2024-01-01'
  AND timestamp <= '2024-01-31'
ORDER BY timestamp DESC
LIMIT 50 OFFSET 0;
```

#### 批量操作优化
```sql
-- 使用批量插入
INSERT INTO price_ticks (symbol, bid_price, ask_price, timestamp) 
VALUES 
  ('BTCUSDT', 50000, 50010, NOW()),
  ('ETHUSDT', 3000, 3010, NOW()),
  ('ADAUSDT', 0.5, 0.51, NOW())
ON CONFLICT DO NOTHING;

-- 使用COPY命令（大批量数据）
COPY price_ticks (symbol, bid_price, ask_price, timestamp) 
FROM '/path/to/data.csv' 
WITH (FORMAT csv, HEADER true);
```

## 缓存性能优化

### 1. Redis 配置优化

#### 内存配置
```bash
# redis.conf
# 最大内存（建议设置为系统内存的50%）
maxmemory 4gb

# 内存淘汰策略
maxmemory-policy allkeys-lru

# 启用内存压缩
hash-max-ziplist-entries 512
hash-max-ziplist-value 64
list-max-ziplist-size -2
set-max-intset-entries 512
```

#### 网络配置
```bash
# 连接池配置
tcp-keepalive 60
timeout 300

# 客户端配置
tcp-backlog 511
```

### 2. 缓存策略优化

#### 缓存键设计
```go
// 使用有意义的键名
const (
    KeyLatestPrice = "cryptosignal:latest_price:%s"      // %s = symbol
    KeyPriceHistory = "cryptosignal:price_history:%s:%s"  // %s = symbol, time
    KeyActiveSymbols = "cryptosignal:active_symbols"
    KeyMetrics = "cryptosignal:metrics:%s"              // %s = symbol
)
```

#### TTL 策略优化
```go
// 根据数据特性设置不同的TTL
const (
    TTLRealTimePrice = 60 * time.Second    // 实时价格：60秒
    TTLPriceHistory = 300 * time.Second     // 价格历史：5分钟
    TTLActiveSymbols = 600 * time.Second    // 活跃交易对：10分钟
    TTLMetrics = 300 * time.Second          // 指标数据：5分钟
)
```

#### 批量操作优化
```go
// 使用Pipeline减少网络往返
func (c *priceCacheImpl) SetMultiplePrices(ctx context.Context, prices []*PriceData) error {
    pipe := c.client.Pipeline()
    
    for _, price := range prices {
        key := BuildLatestPriceKey(price.Symbol)
        data, _ := json.Marshal(price)
        pipe.Set(ctx, key, data, TTLRealTimePrice)
    }
    
    _, err := pipe.Exec(ctx)
    return err
}
```

## 应用层性能优化

### 1. 数据库连接池优化

#### GORM 连接池配置
```go
// 连接池配置
db.SetMaxOpenConns(100)        // 最大打开连接数
db.SetMaxIdleConns(20)         // 最大空闲连接数
db.SetConnMaxLifetime(time.Hour) // 连接最大生存时间
db.SetConnMaxIdleTime(30 * time.Minute) // 空闲连接最大时间
```

#### 连接池监控
```go
// 定期监控连接池状态
func (c *Connection) LogConnectionPoolStats() {
    stats := c.db.Stats()
    c.logger.Info("Database connection pool stats",
        zap.Int("max_open_connections", stats.MaxOpenConnections),
        zap.Int("open_connections", stats.OpenConnections),
        zap.Int("in_use", stats.InUse),
        zap.Int("idle", stats.Idle),
        zap.Int64("wait_count", stats.WaitCount),
        zap.Duration("wait_duration", stats.WaitDuration),
    )
}
```

### 2. 并发处理优化

#### Goroutine 池
```go
// 使用worker pool模式
type WorkerPool struct {
    workers    int
    jobQueue   chan Job
    quit       chan bool
}

func (wp *WorkerPool) Start() {
    for i := 0; i < wp.workers; i++ {
        go wp.worker()
    }
}

func (wp *WorkerPool) worker() {
    for {
        select {
        case job := <-wp.jobQueue:
            job.Process()
        case <-wp.quit:
            return
        }
    }
}
```

#### 批量处理
```go
// 批量处理价格数据
func (s *PriceService) ProcessBatch(prices []PriceData) error {
    const batchSize = 100
    
    for i := 0; i < len(prices); i += batchSize {
        end := i + batchSize
        if end > len(prices) {
            end = len(prices)
        }
        
        batch := prices[i:end]
        if err := s.processBatch(batch); err != nil {
            return err
        }
    }
    
    return nil
}
```

### 3. 内存优化

#### 对象池
```go
// 使用sync.Pool复用对象
var priceDataPool = sync.Pool{
    New: func() interface{} {
        return &PriceData{}
    },
}

func GetPriceData() *PriceData {
    return priceDataPool.Get().(*PriceData)
}

func PutPriceData(p *PriceData) {
    p.Reset()
    priceDataPool.Put(p)
}
```

#### 内存监控
```go
// 定期监控内存使用
func MonitorMemory() {
    var m runtime.MemStats
    runtime.ReadMemStats(&m)
    
    log.Printf("Memory usage: Alloc=%d KB, Sys=%d KB, NumGC=%d",
        m.Alloc/1024, m.Sys/1024, m.NumGC)
}
```

## 系统级性能优化

### 1. 操作系统优化

#### 文件描述符限制
```bash
# 增加文件描述符限制
ulimit -n 65536

# 永久设置
echo "* soft nofile 65536" >> /etc/security/limits.conf
echo "* hard nofile 65536" >> /etc/security/limits.conf
```

#### 网络优化
```bash
# TCP参数优化
echo 'net.core.somaxconn = 65535' >> /etc/sysctl.conf
echo 'net.ipv4.tcp_max_syn_backlog = 65535' >> /etc/sysctl.conf
echo 'net.core.netdev_max_backlog = 5000' >> /etc/sysctl.conf
sysctl -p
```

### 2. 监控和调优

#### 性能监控
```go
// 使用Prometheus监控
var (
    dbQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "db_query_duration_seconds",
            Help: "Database query duration",
        },
        []string{"operation", "table"},
    )
    
    cacheHitRate = prometheus.NewGaugeVec(
        prometheus.GaugeOpts{
            Name: "cache_hit_rate",
            Help: "Cache hit rate",
        },
        []string{"cache_type"},
    )
)
```

#### 性能分析
```go
// 使用pprof进行性能分析
import _ "net/http/pprof"

func main() {
    go func() {
        log.Println(http.ListenAndServe("localhost:6060", nil))
    }()
    
    // 应用代码...
}
```

## 性能测试

### 1. 数据库性能测试

#### 写入性能测试
```go
func BenchmarkPriceTickInsert(b *testing.B) {
    db := setupTestDB()
    defer db.Close()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            priceTick := generateRandomPriceTick()
            db.Create(priceTick)
        }
    })
}
```

#### 查询性能测试
```go
func BenchmarkPriceTickQuery(b *testing.B) {
    db := setupTestDB()
    defer db.Close()
    
    // 准备测试数据
    setupTestData(db, 10000)
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        var prices []PriceTick
        db.Where("symbol = ? AND timestamp > ?", "BTCUSDT", time.Now().Add(-time.Hour)).
           Find(&prices)
    }
}
```

### 2. 缓存性能测试

#### Redis 性能测试
```go
func BenchmarkRedisSet(b *testing.B) {
    client := setupRedisClient()
    defer client.Close()
    
    b.ResetTimer()
    b.RunParallel(func(pb *testing.PB) {
        for pb.Next() {
            key := fmt.Sprintf("test:key:%d", rand.Intn(1000))
            value := generateRandomData()
            client.Set(ctx, key, value, time.Minute)
        }
    })
}
```

## 性能调优检查清单

### 数据库调优
- [ ] PostgreSQL 配置参数优化
- [ ] TimescaleDB 分片策略调整
- [ ] 索引策略优化
- [ ] 查询语句优化
- [ ] 连接池配置调整

### 缓存调优
- [ ] Redis 内存配置优化
- [ ] 缓存键设计优化
- [ ] TTL 策略调整
- [ ] 批量操作优化
- [ ] 缓存穿透保护

### 应用调优
- [ ] 连接池配置优化
- [ ] 并发处理优化
- [ ] 内存使用优化
- [ ] 批量处理优化
- [ ] 错误处理优化

### 系统调优
- [ ] 操作系统参数优化
- [ ] 网络配置优化
- [ ] 监控系统部署
- [ ] 性能测试执行
- [ ] 调优效果验证

## 性能指标目标

### 数据库性能
- 写入性能: > 10,000 条/秒
- 查询延迟: < 100ms (P95)
- 连接池利用率: < 80%
- 慢查询比例: < 1%

### 缓存性能
- 缓存命中率: > 95%
- 缓存延迟: < 5ms (P95)
- 内存使用率: < 80%
- 键过期率: < 5%

### 应用性能
- API 响应时间: < 200ms (P95)
- 并发处理能力: > 1000 QPS
- 内存使用率: < 80%
- CPU 使用率: < 70%

## 故障排查

### 性能问题诊断
1. 检查数据库慢查询日志
2. 分析缓存命中率
3. 监控系统资源使用
4. 检查网络延迟
5. 分析应用日志

### 常见性能问题
1. 数据库连接池耗尽
2. 缓存内存不足
3. 网络延迟过高
4. 磁盘I/O瓶颈
5. 内存泄漏

### 解决方案
1. 调整连接池配置
2. 增加缓存内存
3. 优化网络配置
4. 使用SSD存储
5. 修复内存泄漏
