# Spec Tasks

## 任务清单

基于规范 `2025-10-08-data-storage-design` 的实施任务。

---

- [x] 1. PostgreSQL 数据库和 TimescaleDB 配置 ✅ **已完成**
  - [x] 1.1 配置 Docker Compose 添加 PostgreSQL 15+ 和 TimescaleDB 扩展
  - [x] 1.2 配置数据库连接参数（max_connections、shared_buffers 等）
  - [x] 1.3 创建数据库迁移工具配置（golang-migrate）
  - [x] 1.4 编写第一个迁移文件：启用 TimescaleDB 扩展
  - [x] 1.5 验证 TimescaleDB 扩展安装成功
  - [x] 1.6 配置数据库连接字符串和环境变量

- [x] 2. 数据库表结构创建 ✅ **已完成**
  - [x] 2.1 编写迁移文件：创建 symbols 表（包含所有 BitGet Symbol 字段）
  - [x] 2.2 编写迁移文件：创建 price_ticks 时序表
  - [x] 2.3 编写迁移文件：创建 klines 时序表
  - [x] 2.4 配置 price_ticks 为 TimescaleDB 超表（7天分片）
  - [x] 2.5 配置 klines 为 TimescaleDB 超表（7天分片）
  - [x] 2.6 创建必要的索引（symbols.is_active, price_ticks, klines）
  - [x] 2.7 执行数据库迁移，验证表创建成功
  - [x] 2.8 验证超表配置正确（查询 timescaledb_information.hypertables）

- [x] 3. TimescaleDB 数据压缩和保留策略 ✅ **已完成**
  - [x] 3.1 编写迁移文件：配置 price_ticks 表压缩策略（7天后压缩）
  - [x] 3.2 编写迁移文件：配置 klines 表压缩策略（7天后压缩）
  - [x] 3.3 配置 price_ticks 表数据保留策略（30天自动删除）
  - [x] 3.4 配置 klines 表数据保留策略（30天自动删除）
  - [x] 3.5 验证压缩策略配置成功（查询 timescaledb_information.jobs）
  - [x] 3.6 验证保留策略配置成功（查询 timescaledb_information.jobs）
  - [x] 3.7 编写测试数据生成脚本（生成7天以上的测试数据）
  - [x] 3.8 验证压缩策略执行（手动触发或等待自动执行）

- [x] 4. GORM 数据访问层（DAO）基础架构 ✅ **已完成**
  - [x] 4.1 安装依赖包（gorm.io/gorm, gorm.io/driver/postgres）
  - [x] 4.2 创建数据库连接管理模块（internal/database/connection.go）
  - [x] 4.3 配置 GORM 连接池（最大连接数、空闲连接数、连接超时）
  - [x] 4.4 定义 Symbol、PriceTick、Kline 数据模型（internal/models/）
  - [x] 4.5 实现自定义数据库错误类型（internal/database/errors.go）
  - [x] 4.6 实现数据库健康检查函数
  - [x] 4.7 编写数据库连接测试（测试连接、连接池、健康检查）
  - [x] 4.8 验证所有测试通过

- [x] 5. SymbolDAO 实现 ✅ **已完成**
  - [x] 5.1 编写 SymbolDAO 单元测试（使用 testify 和内存数据库）
  - [x] 5.2 实现 SymbolDAO 接口定义（Create, GetBySymbol, List, Update, Delete）
  - [x] 5.3 实现 Create 方法（插入单个交易对，处理唯一约束冲突）
  - [x] 5.4 实现 CreateBatch 方法（批量插入，单次最多1000条）
  - [x] 5.5 实现 GetBySymbol 方法（根据 symbol 精确查询）
  - [x] 5.6 实现 List 方法（查询所有交易对，支持 is_active 过滤）
  - [x] 5.7 实现 Update 方法（更新交易对信息）
  - [x] 5.8 实现 Upsert 方法（存在则更新，不存在则插入）
  - [x] 5.9 验证所有单元测试通过
  - [x] 5.10 编写集成测试（真实数据库）并验证通过

- [x] 6. KlineDAO 实现 ✅ **已完成**
  - [x] 6.1 编写 KlineDAO 单元测试
  - [x] 6.2 实现 KlineDAO 接口定义（Create, CreateBatch, GetByRange, GetLatest）
  - [x] 6.3 实现 Create 方法（插入单条K线，处理唯一约束冲突）
  - [x] 6.4 实现 CreateBatch 方法（批量插入，使用 ON CONFLICT DO NOTHING）
  - [x] 6.5 实现 GetByRange 方法（时间范围查询，支持分页）
  - [x] 6.6 实现 GetLatest 方法（查询最新N条K线）
  - [x] 6.7 实现 GetBySymbolAndGranularity 方法（按交易对和周期查询）
  - [x] 6.8 实现查询性能优化（确保使用索引，EXPLAIN ANALYZE）
  - [x] 6.9 验证所有单元测试通过
  - [x] 6.10 编写集成测试并验证通过

- [x] 7. TickerDAO 实现 ✅ **已完成**
  - [x] 7.1 编写 TickerDAO 单元测试
  - [x] 7.2 实现 TickerDAO 接口定义（Create, CreateBatch, GetByRange, GetLatest）
  - [x] 7.3 实现 Create 方法（插入单条 Ticker 数据）
  - [x] 7.4 实现 CreateBatch 方法（批量插入，优化写入性能）
  - [x] 7.5 实现 GetLatest 方法（查询指定交易对的最新 Ticker）
  - [x] 7.6 实现 GetByRange 方法（时间范围查询）
  - [x] 7.7 实现 GetLatestMultiple 方法（批量查询多个交易对的最新价格）
  - [x] 7.8 验证所有单元测试通过
  - [x] 7.9 编写集成测试并验证通过

- [x] 8. Redis 缓存基础架构 ✅ **已完成**
  - [x] 8.1 配置 Docker Compose 添加 Redis 7+（设置内存限制、淘汰策略）
  - [x] 8.2 安装 go-redis 依赖包（github.com/redis/go-redis/v9）
  - [x] 8.3 创建 Redis 连接管理模块（internal/cache/connection.go）
  - [x] 8.4 配置 Redis 连接池（PoolSize、MinIdleConns、超时参数）
  - [x] 8.5 定义缓存键命名规范（constants.go：CacheKeyPrice 等）
  - [x] 8.6 实现 Redis 健康检查函数
  - [x] 8.7 编写 Redis 连接测试
  - [x] 8.8 验证所有测试通过

- [ ] 9. Redis 缓存操作实现
  - [ ] 9.1 编写 PriceCache 单元测试（使用 miniredis 模拟）
  - [ ] 9.2 实现 PriceCache 接口（Set, Get, GetMultiple, Delete）
  - [ ] 9.3 实现 SetPrice 方法（缓存实时价格，Hash 结构，TTL 60秒）
  - [ ] 9.4 实现 GetPrice 方法（获取单个交易对价格）
  - [ ] 9.5 实现 GetMultiplePrices 方法（批量获取价格，使用 Pipeline）
  - [ ] 9.6 实现 SetMetrics 方法（缓存实时指标）
  - [ ] 9.7 实现 GetMetrics 方法（获取实时指标）
  - [ ] 9.8 实现 SetActiveSymbols 方法（缓存活跃交易对列表，Set 结构）
  - [ ] 9.9 实现 GetActiveSymbols 方法（获取活跃交易对列表）
  - [ ] 9.10 验证所有单元测试通过
  - [ ] 9.11 编写集成测试（真实 Redis）并验证通过

- [ ] 10. 读写分离配置（可选，开发环境暂不实现）
  - [ ] 10.1 安装 GORM dbresolver 插件（gorm.io/plugin/dbresolver）
  - [ ] 10.2 配置主库连接（写操作）
  - [ ] 10.3 配置从库连接（读操作）
  - [ ] 10.4 实现读写分离策略（随机选择从库）
  - [ ] 10.5 配置复制延迟监控
  - [ ] 10.6 编写读写分离测试（验证写操作到主库，读操作到从库）
  - [ ] 10.7 验证所有测试通过

- [ ] 11. 性能测试和优化
  - [ ] 11.1 编写批量插入性能测试（price_ticks 表，目标 > 5000条/秒）
  - [ ] 11.2 编写时间范围查询性能测试（查询1天K线，目标 < 200ms）
  - [ ] 11.3 编写 Redis 缓存性能测试（读写延迟，目标 < 5ms）
  - [ ] 11.4 编写并发查询测试（100+ 并发，验证连接池稳定）
  - [ ] 11.5 使用 EXPLAIN ANALYZE 分析慢查询
  - [ ] 11.6 优化索引策略（如有必要）
  - [ ] 11.7 验证所有性能指标达标
  - [ ] 11.8 生成性能测试报告

- [ ] 12. 监控和日志集成
  - [ ] 12.1 在所有 DAO 方法中添加结构化日志（使用 Zap）
  - [ ] 12.2 记录慢查询日志（超过 100ms 的查询）
  - [ ] 12.3 实现数据库连接池监控（定期记录连接状态）
  - [ ] 12.4 实现 Redis 内存监控（记录内存使用率）
  - [ ] 12.5 实现缓存命中率统计
  - [ ] 12.6 配置日志级别（开发环境 Debug，生产环境 Info）
  - [ ] 12.7 验证日志输出格式和内容
  - [ ] 12.8 验证监控指标正确记录

- [ ] 13. 数据一致性和错误处理
  - [ ] 13.1 实现数据库事务封装函数（WithTransaction）
  - [ ] 13.2 实现缓存更新策略（先写数据库，后写缓存）
  - [ ] 13.3 实现缓存穿透保护（布隆过滤器或空值缓存）
  - [ ] 13.4 实现重试机制（网络错误自动重试）
  - [ ] 13.5 编写错误处理测试（数据库连接失败、Redis 连接失败等）
  - [ ] 13.6 编写事务测试（验证 ACID 特性）
  - [ ] 13.7 编写缓存一致性测试
  - [ ] 13.8 验证所有测试通过

- [ ] 14. 集成测试和验证
  - [ ] 14.1 编写端到端测试：保存交易对 → 查询 → 验证数据一致
  - [ ] 14.2 编写端到端测试：批量保存K线 → 时间范围查询 → 验证数据
  - [ ] 14.3 编写端到端测试：保存 Ticker → 缓存 → 查询缓存 → 验证一致性
  - [ ] 14.4 验证 TimescaleDB 压缩策略（插入7天前数据，手动触发压缩）
  - [ ] 14.5 验证数据保留策略（插入30天前数据，验证自动删除）
  - [ ] 14.6 验证连接池在高并发下的稳定性（1000+ 并发请求）
  - [ ] 14.7 验证 Redis 缓存 TTL 正确过期
  - [ ] 14.8 生成测试覆盖率报告（目标 > 80%）
  - [ ] 14.9 验证所有 Expected Deliverable 达成

- [ ] 15. 文档和示例代码
  - [ ] 15.1 编写数据库迁移使用文档（migrate up/down 命令）
  - [ ] 15.2 编写 DAO 使用示例代码
  - [ ] 15.3 编写 Redis 缓存使用示例代码
  - [ ] 15.4 更新 README.md（数据库配置、迁移步骤）
  - [ ] 15.5 生成 ER 图（数据库表关系图）
  - [ ] 15.6 编写性能优化建议文档
  - [ ] 15.7 编写故障排查指南
  - [ ] 15.8 验证文档完整性和准确性

---

## 任务说明

**任务总数：** 15 个主要任务，共 156 个子任务

**预计工作量：** S（2-3 天）

**依赖关系：**
- 任务 1-3 是数据库基础，必须先完成
- 任务 4 是 DAO 基础架构，依赖任务 1-3
- 任务 5-7（各个 DAO 实现）依赖任务 4，可以并行开发
- 任务 8-9（Redis 缓存）可以与任务 5-7 并行开发
- 任务 10（读写分离）是可选任务，开发环境可暂不实现
- 任务 11-13（性能测试、监控、错误处理）依赖任务 5-9
- 任务 14（集成测试）依赖所有功能完成
- 任务 15（文档）贯穿整个开发过程

**技术栈：**
- PostgreSQL 15+ with TimescaleDB 2.13+
- GORM v1.25.5
- go-redis v9.3.0
- golang-migrate v4.16.2
- testify (测试框架)
- miniredis (Redis 模拟)

**验证标准：**
按照规范中的 "Expected Deliverable" 部分，所有 5 项交付成果必须能够成功验证：
1. PostgreSQL 数据库创建成功，包含3个表和 TimescaleDB 配置
2. DAO 接口可以正常读写数据，数据一致性验证通过
3. TimescaleDB 压缩和保留策略验证生效
4. Redis 缓存读写正常，TTL 过期策略工作正常
5. 读写分离配置生效，连接池在高并发下稳定（开发环境可选）
