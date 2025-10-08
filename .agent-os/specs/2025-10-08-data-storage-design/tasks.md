# Spec Tasks

## 任务清单

基于规范 `2025-10-08-data-storage-design` 的实施任务。

---

- [ ] 1. 数据库迁移脚本开发
  - [ ] 1.1 创建数据库迁移目录结构（backend/migrations/）
  - [ ] 1.2 编写 001_create_symbols_table.sql 迁移脚本
  - [ ] 1.3 编写 002_create_price_ticks_table.sql 迁移脚本
  - [ ] 1.4 编写 003_create_klines_table.sql 迁移脚本
  - [ ] 1.5 编写 004_create_indexes.sql 迁移脚本
  - [ ] 1.6 编写 005_configure_timescaledb_policies.sql 迁移脚本
  - [ ] 1.7 编写对应的回滚脚本（rollback_001.sql 至 rollback_005.sql）
  - [ ] 1.8 验证所有迁移脚本的SQL语法正确性

- [ ] 2. 数据库迁移工具集成
  - [ ] 2.1 安装 golang-migrate/migrate 包
  - [ ] 2.2 创建数据库配置文件（database.yaml）
  - [ ] 2.3 编写迁移工具初始化代码（internal/db/migrate.go）
  - [ ] 2.4 实现 migrate up 命令（执行所有待执行迁移）
  - [ ] 2.5 实现 migrate down 命令（回滚迁移）
  - [ ] 2.6 实现 migrate version 命令（查看当前版本）
  - [ ] 2.7 编写迁移工具的单元测试
  - [ ] 2.8 验证迁移工具能够成功执行所有迁移脚本

- [ ] 3. 数据模型（Models）定义
  - [ ] 3.1 创建 internal/models 目录
  - [ ] 3.2 定义 Symbol 结构体（对应 symbols 表）
  - [ ] 3.3 定义 PriceTick 结构体（对应 price_ticks 表）
  - [ ] 3.4 定义 Kline 结构体（对应 klines 表）
  - [ ] 3.5 添加模型验证方法（Validate）
  - [ ] 3.6 添加模型转换方法（ToJSON/FromJSON）
  - [ ] 3.7 编写模型单元测试
  - [ ] 3.8 验证所有模型字段与数据库表字段一致

- [ ] 4. 数据库连接管理
  - [ ] 4.1 创建 internal/db 目录
  - [ ] 4.2 实现 DatabaseConfig 配置结构
  - [ ] 4.3 实现数据库连接初始化函数（NewDatabase）
  - [ ] 4.4 实现连接池配置（MaxOpenConns、MaxIdleConns等）
  - [ ] 4.5 实现读写分离支持（Master/Slave连接）
  - [ ] 4.6 实现数据库健康检查（Ping）
  - [ ] 4.7 实现优雅关闭（Close）
  - [ ] 4.8 编写数据库连接测试，验证连接池和读写分离工作正常

- [ ] 5. SymbolDAO 实现
  - [ ] 5.1 创建 internal/dao/symbol_dao.go
  - [ ] 5.2 定义 SymbolDAO 接口
  - [ ] 5.3 实现 CreateSymbol 方法
  - [ ] 5.4 实现 GetSymbolByName 方法
  - [ ] 5.5 实现 GetSymbolByID 方法
  - [ ] 5.6 实现 ListActiveSymbols 方法
  - [ ] 5.7 实现 UpdateSymbol 方法
  - [ ] 5.8 实现 BatchCreateSymbols 方法
  - [ ] 5.9 编写 SymbolDAO 单元测试（使用 testify/suite）
  - [ ] 5.10 验证所有 SymbolDAO 方法通过测试

- [ ] 6. PriceTickDAO 实现
  - [ ] 6.1 创建 internal/dao/price_tick_dao.go
  - [ ] 6.2 定义 PriceTickDAO 接口
  - [ ] 6.3 实现 InsertTick 方法
  - [ ] 6.4 实现 BatchInsertTicks 方法（批量插入优化）
  - [ ] 6.5 实现 GetLatestTick 方法
  - [ ] 6.6 实现 GetTicksByTimeRange 方法
  - [ ] 6.7 实现 GetTicksWithPagination 方法
  - [ ] 6.8 编写 PriceTickDAO 单元测试
  - [ ] 6.9 性能测试：验证批量插入性能达到 500+ 条/秒
  - [ ] 6.10 验证所有 PriceTickDAO 方法通过测试

- [ ] 7. KlineDAO 实现
  - [ ] 7.1 创建 internal/dao/kline_dao.go
  - [ ] 7.2 定义 KlineDAO 接口
  - [ ] 7.3 实现 InsertKline 方法
  - [ ] 7.4 实现 BatchInsertKlines 方法
  - [ ] 7.5 实现 GetKlines 方法（按时间范围查询）
  - [ ] 7.6 实现 GetLatestKline 方法
  - [ ] 7.7 编写 KlineDAO 单元测试
  - [ ] 7.8 性能测试：验证查询24小时K线数据耗时 < 100ms
  - [ ] 7.9 验证所有 KlineDAO 方法通过测试

- [ ] 8. Redis 缓存层实现
  - [ ] 8.1 创建 internal/cache 目录
  - [ ] 8.2 实现 RedisConfig 配置结构
  - [ ] 8.3 实现 Redis 连接初始化（NewRedisClient）
  - [ ] 8.4 定义 PriceData 和 PriceChange 数据结构
  - [ ] 8.5 实现 RedisPriceCache 接口
  - [ ] 8.6 实现 SetPrice 和 GetPrice 方法
  - [ ] 8.7 实现 BatchSetPrices 方法
  - [ ] 8.8 实现 SetPriceChange 和 GetPriceChange 方法
  - [ ] 8.9 实现 SetActiveSymbols 和 GetActiveSymbols 方法
  - [ ] 8.10 实现 ClearCache 方法
  - [ ] 8.11 编写 Redis 缓存层单元测试（使用 miniredis 模拟）
  - [ ] 8.12 性能测试：验证 Redis 查询响应时间 < 10ms
  - [ ] 8.13 验证所有 Redis 缓存方法通过测试

- [ ] 9. 集成测试和性能验证
  - [ ] 9.1 创建 internal/dao/integration_test.go
  - [ ] 9.2 编写 SymbolDAO 集成测试（真实数据库）
  - [ ] 9.3 编写 PriceTickDAO 集成测试
  - [ ] 9.4 编写 KlineDAO 集成测试
  - [ ] 9.5 编写 Redis 缓存集成测试
  - [ ] 9.6 编写批量插入性能测试（500+ 条/秒）
  - [ ] 9.7 编写查询性能测试（PostgreSQL < 100ms）
  - [ ] 9.8 编写 Redis 性能测试（< 10ms）
  - [ ] 9.9 验证 TimescaleDB 超表创建成功
  - [ ] 9.10 验证数据压缩和保留策略配置生效
  - [ ] 9.11 验证所有集成测试通过

- [ ] 10. 数据初始化和文档完善
  - [ ] 10.1 创建 seed_symbols.sql 初始化脚本
  - [ ] 10.2 从 BitGet API 获取真实交易对列表
  - [ ] 10.3 批量导入交易对数据到数据库
  - [ ] 10.4 验证交易对数据导入成功（至少 50+ 交易对）
  - [ ] 10.5 编写数据库操作文档（README_DATABASE.md）
  - [ ] 10.6 编写 DAO 使用示例代码
  - [ ] 10.7 编写性能优化建议文档
  - [ ] 10.8 更新项目主 README，添加数据库设置说明
  - [ ] 10.9 验证所有交付成果（5 个预期交付项）

---

## 任务说明

**任务总数：** 10 个主要任务，共 88 个子任务

**预计工作量：** S（2-3 天）

**依赖关系：**
- 任务 1（迁移脚本）是基础，必须先完成
- 任务 2（迁移工具）依赖任务 1
- 任务 3（数据模型）可以与任务 1-2 并行开发
- 任务 4（连接管理）是 DAO 层的基础
- 任务 5-7（DAO 实现）依赖任务 3 和 4
- 任务 8（Redis 缓存）可以与任务 5-7 并行开发
- 任务 9（集成测试）在所有 DAO 完成后进行
- 任务 10（数据初始化）在测试通过后进行

**技术栈：**
- PostgreSQL 15+ with TimescaleDB
- Redis 7+
- golang-migrate/migrate
- database/sql + github.com/lib/pq
- github.com/go-redis/redis/v8
- testify/suite（测试框架）
- miniredis（Redis 模拟）

**验证标准：**
按照规范中的 "Expected Deliverable" 部分，所有 5 项交付成果必须能够成功验证：
1. SQL 迁移脚本成功创建所有表和 TimescaleDB 配置
2. DAO 接口支持高性能数据操作（500+ 条/秒写入）
3. Redis 缓存响应时间 < 10ms
4. 数据库查询性能满足要求（< 100ms）
5. 数据保留策略正常工作（30 天自动压缩和清理）

