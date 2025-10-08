# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-08-data-storage-design/spec.md

## Technical Requirements

### PostgreSQL表结构

#### 1. symbols表（交易对配置）
- `id` (BIGSERIAL PRIMARY KEY): 自增主键
- `symbol` (VARCHAR(50) UNIQUE NOT NULL): 交易对名称（如BTCUSDT）
- `base_coin` (VARCHAR(20) NOT NULL): 基础币种
- `quote_coin` (VARCHAR(20) NOT NULL): 计价币种
- `contract_type` (VARCHAR(20) NOT NULL): 合约类型（perpetual等）
- `min_trade_num` (DECIMAL(20,8)): 最小交易数量
- `price_precision` (INT): 价格精度
- `volume_precision` (INT): 数量精度
- `status` (VARCHAR(20) DEFAULT 'active'): 状态（active/inactive）
- `created_at` (TIMESTAMPTZ DEFAULT NOW()): 创建时间
- `updated_at` (TIMESTAMPTZ DEFAULT NOW()): 更新时间
- **索引**: `idx_symbol` (symbol), `idx_status` (status)

#### 2. price_ticks表（实时价格数据 - TimescaleDB超表）
- `time` (TIMESTAMPTZ NOT NULL): 时间戳（分区键）
- `symbol_id` (BIGINT NOT NULL REFERENCES symbols(id)): 交易对ID
- `symbol` (VARCHAR(50) NOT NULL): 交易对名称（冗余字段，提升查询性能）
- `last_price` (DECIMAL(20,8) NOT NULL): 最新价格
- `bid_price` (DECIMAL(20,8)): 买一价
- `ask_price` (DECIMAL(20,8)): 卖一价
- `bid_size` (DECIMAL(20,8)): 买一量
- `ask_size` (DECIMAL(20,8)): 卖一量
- `volume_24h` (DECIMAL(30,8)): 24小时成交量
- `turnover_24h` (DECIMAL(30,8)): 24小时成交额
- `change_24h` (DECIMAL(10,4)): 24小时涨跌幅（百分比）
- `high_24h` (DECIMAL(20,8)): 24小时最高价
- `low_24h` (DECIMAL(20,8)): 24小时最低价
- **主键**: (time, symbol_id)
- **索引**: `idx_price_ticks_symbol_time` (symbol_id, time DESC)
- **TimescaleDB配置**:
  - 分区间隔: 1天
  - 压缩策略: 7天后压缩
  - 数据保留策略: 30天后删除

#### 3. klines表（K线数据 - TimescaleDB超表）
- `time` (TIMESTAMPTZ NOT NULL): K线开始时间（分区键）
- `symbol_id` (BIGINT NOT NULL REFERENCES symbols(id)): 交易对ID
- `symbol` (VARCHAR(50) NOT NULL): 交易对名称
- `interval` (VARCHAR(10) NOT NULL): K线周期（1m/5m/15m/1h/4h/1d）
- `open` (DECIMAL(20,8) NOT NULL): 开盘价
- `high` (DECIMAL(20,8) NOT NULL): 最高价
- `low` (DECIMAL(20,8) NOT NULL): 最低价
- `close` (DECIMAL(20,8) NOT NULL): 收盘价
- `base_volume` (DECIMAL(30,8) NOT NULL): 交易币成交量
- `quote_volume` (DECIMAL(30,8) NOT NULL): 计价币成交量
- **主键**: (time, symbol_id, interval)
- **索引**: `idx_klines_symbol_interval_time` (symbol_id, interval, time DESC)
- **TimescaleDB配置**:
  - 分区间隔: 7天
  - 压缩策略: 14天后压缩
  - 数据保留策略: 30天后删除

### Redis缓存设计

#### 1. 实时价格缓存
- **键格式**: `price:{symbol}` (例: `price:BTCUSDT`)
- **数据结构**: Hash
- **字段**:
  - `lastPrice`: 最新价格
  - `bidPrice`: 买一价
  - `askPrice`: 卖一价
  - `volume24h`: 24小时成交量
  - `change24h`: 24小时涨跌幅
  - `updateTime`: 更新时间戳
- **过期时间**: 60秒（自动刷新）

#### 2. 交易对列表缓存
- **键格式**: `symbols:active`
- **数据结构**: Set
- **内容**: 所有活跃交易对的symbol列表
- **过期时间**: 300秒（5分钟）

#### 3. 价格变化率缓存（实时指标计算）
- **键格式**: `change:{symbol}:{window}` (例: `change:BTCUSDT:1m`)
- **数据结构**: String（JSON格式）
- **字段**:
  - `priceChange`: 价格变化量
  - `priceChangePercent`: 价格变化率（百分比）
  - `startPrice`: 窗口起始价格
  - `endPrice`: 窗口结束价格
  - `startTime`: 窗口起始时间
  - `endTime`: 窗口结束时间
- **支持窗口**: 1m, 5m, 15m, 30m
- **过期时间**: 根据窗口大小设置（1m→60s, 5m→300s等）

#### 4. 连接管理
- **键格式**: `ws:connections`
- **数据结构**: Hash
- **用途**: 记录WebSocket活跃连接数和订阅信息

### 数据访问层（DAO）设计

#### 1. 数据库连接配置
```go
type DatabaseConfig struct {
    // 主库（写）
    MasterDSN string
    // 从库（读）
    SlaveDSN  string
    // 连接池配置
    MaxOpenConns    int // 最大连接数（默认25）
    MaxIdleConns    int // 最大空闲连接数（默认5）
    ConnMaxLifetime time.Duration // 连接最大生命周期（默认5分钟）
    ConnMaxIdleTime time.Duration // 连接最大空闲时间（默认1分钟）
}
```

#### 2. DAO接口定义

**SymbolDAO**:
- `CreateSymbol(symbol *Symbol) error`: 创建交易对
- `GetSymbolByName(name string) (*Symbol, error)`: 根据名称查询
- `GetSymbolByID(id int64) (*Symbol, error)`: 根据ID查询
- `ListActiveSymbols() ([]*Symbol, error)`: 获取所有活跃交易对
- `UpdateSymbol(symbol *Symbol) error`: 更新交易对信息
- `BatchCreateSymbols(symbols []*Symbol) error`: 批量创建

**PriceTickDAO**:
- `InsertTick(tick *PriceTick) error`: 插入单条价格记录
- `BatchInsertTicks(ticks []*PriceTick) error`: 批量插入（优化性能）
- `GetLatestTick(symbolID int64) (*PriceTick, error)`: 获取最新价格
- `GetTicksByTimeRange(symbolID int64, start, end time.Time) ([]*PriceTick, error)`: 按时间范围查询
- `GetTicksWithPagination(symbolID int64, limit, offset int) ([]*PriceTick, error)`: 分页查询

**KlineDAO**:
- `InsertKline(kline *Kline) error`: 插入单条K线
- `BatchInsertKlines(klines []*Kline) error`: 批量插入
- `GetKlines(symbolID int64, interval string, start, end time.Time) ([]*Kline, error)`: 查询K线数据
- `GetLatestKline(symbolID int64, interval string) (*Kline, error)`: 获取最新K线

#### 3. 缓存层接口（RedisDAO）

**RedisPriceCache**:
- `SetPrice(symbol string, price *PriceData) error`: 设置实时价格
- `GetPrice(symbol string) (*PriceData, error)`: 获取实时价格
- `BatchSetPrices(prices map[string]*PriceData) error`: 批量设置
- `SetPriceChange(symbol, window string, change *PriceChange) error`: 设置价格变化率
- `GetPriceChange(symbol, window string) (*PriceChange, error)`: 获取价格变化率
- `SetActiveSymbols(symbols []string) error`: 设置活跃交易对列表
- `GetActiveSymbols() ([]string, error)`: 获取活跃交易对列表
- `ClearCache(pattern string) error`: 清空缓存

### 性能要求

- **写入性能**: 支持每秒500+条price_tick记录写入
- **查询性能**: 
  - Redis查询: < 10ms
  - PostgreSQL单表查询: < 50ms
  - PostgreSQL JOIN查询: < 100ms
- **连接池**: 主库25连接，从库50连接（读多写少）
- **批量操作**: 优先使用批量插入（100条/批次）减少数据库往返

### 事务管理

- 交易对创建需要事务保证
- 价格数据写入无需事务（允许部分失败）
- 使用`database/sql`包的`Begin()`, `Commit()`, `Rollback()`
- 读写分离：写操作使用master，读操作使用slave

### 错误处理

- 数据库连接错误: 自动重试3次，间隔1秒
- 主键冲突: 返回明确的错误信息
- 外键约束违反: 返回业务错误
- 超时错误: 设置合理的查询超时（5秒）

## External Dependencies

无需新增外部依赖。使用已有的技术栈：

- **database/sql** (Go标准库): 数据库连接
- **github.com/lib/pq**: PostgreSQL驱动
- **github.com/go-redis/redis/v8**: Redis客户端
- **TimescaleDB**: PostgreSQL扩展（已在tech-stack中确定）

