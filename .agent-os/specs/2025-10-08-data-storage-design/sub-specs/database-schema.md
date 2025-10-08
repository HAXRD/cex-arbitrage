# Database Schema

This is the database schema implementation for the spec detailed in @.agent-os/specs/2025-10-08-data-storage-design/spec.md

## 迁移脚本概览

数据库迁移采用版本化管理，按照以下顺序执行：

1. `001_create_symbols_table.sql` - 创建交易对表
2. `002_create_price_ticks_table.sql` - 创建实时价格表并配置TimescaleDB
3. `003_create_klines_table.sql` - 创建K线表并配置TimescaleDB
4. `004_create_indexes.sql` - 创建索引
5. `005_configure_timescaledb_policies.sql` - 配置TimescaleDB策略（压缩、保留）

## 迁移脚本

### 001_create_symbols_table.sql

```sql
-- 创建交易对配置表
CREATE TABLE IF NOT EXISTS symbols (
    id BIGSERIAL PRIMARY KEY,
    symbol VARCHAR(50) UNIQUE NOT NULL,
    base_coin VARCHAR(20) NOT NULL,
    quote_coin VARCHAR(20) NOT NULL,
    contract_type VARCHAR(20) NOT NULL DEFAULT 'perpetual',
    min_trade_num DECIMAL(20, 8) NOT NULL DEFAULT 0.01,
    price_precision INT NOT NULL DEFAULT 8,
    volume_precision INT NOT NULL DEFAULT 8,
    status VARCHAR(20) NOT NULL DEFAULT 'active',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 添加注释
COMMENT ON TABLE symbols IS '交易对配置表';
COMMENT ON COLUMN symbols.symbol IS '交易对名称（如BTCUSDT）';
COMMENT ON COLUMN symbols.base_coin IS '基础币种';
COMMENT ON COLUMN symbols.quote_coin IS '计价币种';
COMMENT ON COLUMN symbols.contract_type IS '合约类型（perpetual永续等）';
COMMENT ON COLUMN symbols.status IS '状态（active活跃/inactive停用）';

-- 创建更新时间触发器
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_symbols_updated_at BEFORE UPDATE ON symbols
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
```

**设计说明**:
- `symbol`字段设置UNIQUE约束，防止重复
- `status`字段用于软删除，保留历史数据
- 自动更新`updated_at`字段，便于追踪变更
- 价格和数量精度字段用于前端格式化显示

### 002_create_price_ticks_table.sql

```sql
-- 创建实时价格数据表（普通表）
CREATE TABLE IF NOT EXISTS price_ticks (
    time TIMESTAMPTZ NOT NULL,
    symbol_id BIGINT NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    last_price DECIMAL(20, 8) NOT NULL,
    bid_price DECIMAL(20, 8),
    ask_price DECIMAL(20, 8),
    bid_size DECIMAL(20, 8),
    ask_size DECIMAL(20, 8),
    volume_24h DECIMAL(30, 8),
    turnover_24h DECIMAL(30, 8),
    change_24h DECIMAL(10, 4),
    high_24h DECIMAL(20, 8),
    low_24h DECIMAL(20, 8),
    PRIMARY KEY (time, symbol_id),
    CONSTRAINT fk_price_ticks_symbol FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);

-- 添加注释
COMMENT ON TABLE price_ticks IS '实时价格数据表（TimescaleDB超表）';
COMMENT ON COLUMN price_ticks.time IS '时间戳（分区键）';
COMMENT ON COLUMN price_ticks.symbol IS '交易对名称（冗余字段，提升查询性能）';
COMMENT ON COLUMN price_ticks.last_price IS '最新成交价';
COMMENT ON COLUMN price_ticks.change_24h IS '24小时涨跌幅（百分比）';

-- 转换为TimescaleDB超表
SELECT create_hypertable('price_ticks', 'time', 
    chunk_time_interval => INTERVAL '1 day',
    if_not_exists => TRUE
);

-- 创建复合索引（优化按symbol和时间查询）
CREATE INDEX IF NOT EXISTS idx_price_ticks_symbol_time 
    ON price_ticks (symbol_id, time DESC);

CREATE INDEX IF NOT EXISTS idx_price_ticks_symbol_name 
    ON price_ticks (symbol, time DESC);
```

**设计说明**:
- `time`和`symbol_id`作为复合主键，保证唯一性
- `symbol`字段冗余存储，避免JOIN查询，提升性能
- 使用TimescaleDB的`create_hypertable`转换为超表
- 分区间隔设置为1天，平衡查询性能和分区管理
- 复合索引支持高效的按交易对和时间范围查询

### 003_create_klines_table.sql

```sql
-- 创建K线数据表（普通表）
CREATE TABLE IF NOT EXISTS klines (
    time TIMESTAMPTZ NOT NULL,
    symbol_id BIGINT NOT NULL,
    symbol VARCHAR(50) NOT NULL,
    interval VARCHAR(10) NOT NULL,
    open DECIMAL(20, 8) NOT NULL,
    high DECIMAL(20, 8) NOT NULL,
    low DECIMAL(20, 8) NOT NULL,
    close DECIMAL(20, 8) NOT NULL,
    base_volume DECIMAL(30, 8) NOT NULL,
    quote_volume DECIMAL(30, 8) NOT NULL,
    PRIMARY KEY (time, symbol_id, interval),
    CONSTRAINT fk_klines_symbol FOREIGN KEY (symbol_id) REFERENCES symbols(id) ON DELETE CASCADE
);

-- 添加注释
COMMENT ON TABLE klines IS 'K线数据表（TimescaleDB超表）';
COMMENT ON COLUMN klines.time IS 'K线开始时间（分区键）';
COMMENT ON COLUMN klines.interval IS 'K线周期（1m/5m/15m/1h/4h/1d）';
COMMENT ON COLUMN klines.base_volume IS '交易币成交量';
COMMENT ON COLUMN klines.quote_volume IS '计价币成交量（USDT）';

-- 转换为TimescaleDB超表
SELECT create_hypertable('klines', 'time', 
    chunk_time_interval => INTERVAL '7 days',
    if_not_exists => TRUE
);

-- 创建复合索引（优化按symbol、interval和时间查询）
CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval_time 
    ON klines (symbol_id, interval, time DESC);

CREATE INDEX IF NOT EXISTS idx_klines_symbol_name_interval 
    ON klines (symbol, interval, time DESC);

-- 添加CHECK约束，限制interval取值
ALTER TABLE klines ADD CONSTRAINT check_interval 
    CHECK (interval IN ('1m', '5m', '15m', '30m', '1h', '4h', '1d'));
```

**设计说明**:
- 三字段复合主键（time, symbol_id, interval），确保同一时间、同一交易对、同一周期的K线唯一
- 分区间隔设置为7天，K线数据相比价格tick更稀疏
- CHECK约束限制interval字段取值，避免脏数据
- 复合索引支持按交易对和周期的高效查询

### 004_create_indexes.sql

```sql
-- symbols表索引
CREATE INDEX IF NOT EXISTS idx_symbols_symbol ON symbols(symbol);
CREATE INDEX IF NOT EXISTS idx_symbols_status ON symbols(status);
CREATE INDEX IF NOT EXISTS idx_symbols_created_at ON symbols(created_at DESC);

-- price_ticks表额外索引
CREATE INDEX IF NOT EXISTS idx_price_ticks_time ON price_ticks(time DESC);

-- klines表额外索引  
CREATE INDEX IF NOT EXISTS idx_klines_time ON klines(time DESC);
CREATE INDEX IF NOT EXISTS idx_klines_interval ON klines(interval);

-- 分析表统计信息
ANALYZE symbols;
ANALYZE price_ticks;
ANALYZE klines;
```

**设计说明**:
- `idx_symbols_symbol`: 支持按交易对名称快速查询
- `idx_symbols_status`: 支持按状态过滤活跃交易对
- `idx_price_ticks_time`: 支持按时间范围查询所有交易对
- `idx_klines_interval`: 支持按周期查询K线
- `ANALYZE`更新统计信息，优化查询计划

### 005_configure_timescaledb_policies.sql

```sql
-- price_ticks表：配置数据压缩策略（7天后压缩）
SELECT add_compression_policy('price_ticks', INTERVAL '7 days', if_not_exists => TRUE);

-- price_ticks表：配置数据保留策略（30天后删除）
SELECT add_retention_policy('price_ticks', INTERVAL '30 days', if_not_exists => TRUE);

-- klines表：配置数据压缩策略（14天后压缩）
SELECT add_compression_policy('klines', INTERVAL '14 days', if_not_exists => TRUE);

-- klines表：配置数据保留策略（30天后删除）
SELECT add_retention_policy('klines', INTERVAL '30 days', if_not_exists => TRUE);

-- 启用自动分区管理
SELECT set_chunk_time_interval('price_ticks', INTERVAL '1 day');
SELECT set_chunk_time_interval('klines', INTERVAL '7 days');

-- 查看策略配置
SELECT * FROM timescaledb_information.jobs;
```

**设计说明**:
- **压缩策略**: price_ticks 7天后压缩（数据量大），klines 14天后压缩（数据量相对小）
- **保留策略**: 统一保留30天，符合"近一个月"要求
- **压缩比例**: TimescaleDB压缩通常可达10:1到20:1，显著节省存储空间
- **自动执行**: 策略由TimescaleDB后台自动执行，无需手动干预

## 回滚脚本

### rollback_005.sql
```sql
SELECT remove_compression_policy('price_ticks', if_exists => TRUE);
SELECT remove_retention_policy('price_ticks', if_exists => TRUE);
SELECT remove_compression_policy('klines', if_exists => TRUE);
SELECT remove_retention_policy('klines', if_exists => TRUE);
```

### rollback_004.sql
```sql
DROP INDEX IF EXISTS idx_symbols_symbol;
DROP INDEX IF EXISTS idx_symbols_status;
DROP INDEX IF EXISTS idx_symbols_created_at;
DROP INDEX IF EXISTS idx_price_ticks_time;
DROP INDEX IF EXISTS idx_klines_time;
DROP INDEX IF EXISTS idx_klines_interval;
```

### rollback_003.sql
```sql
DROP INDEX IF EXISTS idx_klines_symbol_interval_time;
DROP INDEX IF EXISTS idx_klines_symbol_name_interval;
DROP TABLE IF EXISTS klines CASCADE;
```

### rollback_002.sql
```sql
DROP INDEX IF EXISTS idx_price_ticks_symbol_time;
DROP INDEX IF EXISTS idx_price_ticks_symbol_name;
DROP TABLE IF EXISTS price_ticks CASCADE;
```

### rollback_001.sql
```sql
DROP TRIGGER IF EXISTS update_symbols_updated_at ON symbols;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS symbols CASCADE;
```

## 初始化数据

### seed_symbols.sql

```sql
-- 插入示例交易对数据（从BitGet API获取后可批量导入）
INSERT INTO symbols (symbol, base_coin, quote_coin, contract_type, min_trade_num, price_precision, volume_precision) VALUES
('BTCUSDT', 'BTC', 'USDT', 'perpetual', 0.001, 1, 3),
('ETHUSDT', 'ETH', 'USDT', 'perpetual', 0.01, 2, 2),
('BNBUSDT', 'BNB', 'USDT', 'perpetual', 0.01, 2, 2)
ON CONFLICT (symbol) DO NOTHING;
```

## 数据迁移工具

建议使用Go编写的迁移工具，支持：
- 按序执行迁移脚本
- 记录迁移版本（可创建`schema_migrations`表）
- 支持回滚操作
- 验证迁移结果

推荐工具：
- **golang-migrate/migrate**: 成熟的Go迁移工具
- **goose**: 轻量级迁移工具

## 性能测试SQL

### 1. 测试写入性能（批量插入）
```sql
-- 插入10000条price_tick记录，测试耗时
EXPLAIN ANALYZE
INSERT INTO price_ticks (time, symbol_id, symbol, last_price, volume_24h)
SELECT 
    NOW() - (n || ' seconds')::INTERVAL as time,
    1 as symbol_id,
    'BTCUSDT' as symbol,
    50000 + (random() * 1000) as last_price,
    1000000 as volume_24h
FROM generate_series(1, 10000) n;
```

### 2. 测试查询性能（按symbol和时间范围）
```sql
-- 查询指定交易对最近1小时的价格数据
EXPLAIN ANALYZE
SELECT * FROM price_ticks
WHERE symbol_id = 1
  AND time >= NOW() - INTERVAL '1 hour'
ORDER BY time DESC;
```

### 3. 测试JOIN性能
```sql
-- JOIN查询交易对和价格数据
EXPLAIN ANALYZE
SELECT s.symbol, p.last_price, p.change_24h
FROM symbols s
JOIN price_ticks p ON s.id = p.symbol_id
WHERE s.status = 'active'
  AND p.time >= NOW() - INTERVAL '5 minutes'
ORDER BY p.change_24h DESC
LIMIT 10;
```

### 4. 检查表大小和压缩效果
```sql
-- 查看表占用空间
SELECT 
    hypertable_name,
    pg_size_pretty(hypertable_size(format('%I.%I', hypertable_schema, hypertable_name))) as table_size,
    pg_size_pretty(indexes_size(format('%I.%I', hypertable_schema, hypertable_name))) as indexes_size
FROM timescaledb_information.hypertables;

-- 查看压缩状态
SELECT chunk_name, 
       pg_size_pretty(before_compression_total_bytes) as before,
       pg_size_pretty(after_compression_total_bytes) as after,
       round((1 - after_compression_total_bytes::numeric / before_compression_total_bytes) * 100, 2) as compression_ratio
FROM timescaledb_information.compressed_chunk_stats;
```

## 数据库性能优化建议

1. **连接池**: 使用连接池避免频繁建立连接
2. **批量操作**: 优先使用批量插入而非逐条插入
3. **EXPLAIN ANALYZE**: 定期分析慢查询，优化索引
4. **VACUUM**: 定期执行VACUUM和ANALYZE，维护统计信息
5. **分区管理**: TimescaleDB自动管理分区，无需手动干预
6. **监控**: 使用pg_stat_statements扩展监控SQL性能

