# 数据库模式

这是针对 @.agent-os/specs/2025-10-17-rest-api-development/spec.md 中详细规范的数据库模式实现文档

## 变更

### 新增表

#### monitoring_configs 表
用于存储用户监控配置信息

```sql
CREATE TABLE monitoring_configs (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    filters JSONB NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 索引和约束

#### 索引
```sql
-- 监控配置表索引
CREATE INDEX idx_monitoring_configs_name ON monitoring_configs(name);
CREATE INDEX idx_monitoring_configs_is_default ON monitoring_configs(is_default);
CREATE INDEX idx_monitoring_configs_created_at ON monitoring_configs(created_at);

-- JSONB字段索引（用于高效查询配置内容）
CREATE INDEX idx_monitoring_configs_filters_gin ON monitoring_configs USING GIN(filters);
```

#### 约束
```sql
-- 确保配置名称唯一
ALTER TABLE monitoring_configs ADD CONSTRAINT uk_monitoring_configs_name UNIQUE (name);

-- 确保只有一个默认配置
CREATE UNIQUE INDEX uk_monitoring_configs_default ON monitoring_configs(is_default) WHERE is_default = TRUE;
```

### 修改现有表

#### symbols 表扩展
为支持API查询需求，添加必要的索引：

```sql
-- 为现有字段添加索引
CREATE INDEX IF NOT EXISTS idx_symbols_symbol ON symbols(symbol);
CREATE INDEX IF NOT EXISTS idx_symbols_symbol_type ON symbols(symbol_type);
CREATE INDEX IF NOT EXISTS idx_symbols_symbol_status ON symbols(symbol_status);
CREATE INDEX IF NOT EXISTS idx_symbols_is_active ON symbols(is_active);
CREATE INDEX IF NOT EXISTS idx_symbols_created_at ON symbols(created_at);
```

#### price_ticks 表扩展
为支持实时价格查询，添加索引：

```sql
-- 为价格数据查询添加复合索引
CREATE INDEX IF NOT EXISTS idx_price_ticks_symbol_timestamp ON price_ticks(symbol, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_price_ticks_timestamp ON price_ticks(timestamp DESC);
```

#### klines 表扩展
为支持K线数据查询，添加索引：

```sql
-- 为K线数据查询添加复合索引
CREATE INDEX IF NOT EXISTS idx_klines_symbol_interval_timestamp ON klines(symbol, interval, timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_klines_symbol_timestamp ON klines(symbol, timestamp DESC);
```

## 规范说明

### 监控配置表设计理由

1. **JSONB字段存储配置**：
   - 使用JSONB类型存储filters字段，支持灵活的配置结构
   - 利用PostgreSQL的JSONB索引功能，提高查询性能
   - 便于后续扩展配置项，无需修改表结构

2. **默认配置约束**：
   - 通过唯一索引确保只有一个默认配置
   - 简化前端默认配置的获取逻辑

3. **时间戳字段**：
   - 使用TIMESTAMP WITH TIME ZONE确保时区一致性
   - 便于配置版本管理和审计

### 性能考虑

1. **索引策略**：
   - 为常用查询字段添加索引
   - 使用复合索引优化多条件查询
   - JSONB字段使用GIN索引支持复杂查询

2. **查询优化**：
   - 分页查询使用LIMIT和OFFSET
   - 时间范围查询使用索引优化
   - 避免全表扫描的查询模式

### 数据完整性规则

1. **配置名称唯一性**：
   - 防止重复配置名称
   - 便于配置管理和识别

2. **默认配置唯一性**：
   - 确保系统只有一个默认配置
   - 简化前端逻辑处理

3. **JSONB数据验证**：
   - 在应用层验证JSONB字段的数据结构
   - 确保配置数据的完整性和有效性
