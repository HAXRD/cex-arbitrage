# 项目文档索引

本目录包含项目的技术文档、实现总结和测试报告。

## 📚 文档列表

### 技术文档

#### [read_write_splitting.md](./read_write_splitting.md)
**数据库读写分离配置文档**

详细说明了 PostgreSQL 数据库读写分离的实现：
- 架构设计和路由策略
- 配置方法（开发/生产环境）
- 使用示例（自动路由、显式指定、事务）
- 监控和维护
- 性能优化建议
- 故障处理和最佳实践

**关键内容**：
- 支持多从库配置
- RandomPolicy 负载均衡
- 自动路由和故障降级
- 复制延迟监控

---

### 测试报告

#### [performance_test_report.md](./performance_test_report.md)
**性能测试报告**

包含数据存储层的全面性能测试结果：

**测试范围**：
1. **Redis 缓存性能**
   - 单个价格读写：< 1ms
   - 批量获取 100 个价格：1.15ms
   - 并发读取 100 次：4.35ms
   - ✅ 所有指标远超预期目标

2. **KlineDAO 性能**
   - 批量插入 100 条：18.87ms
   - 查询最新 100 条：2.88ms
   - 时间范围查询：1.93ms
   - 分页查询：1.05ms
   - ✅ 使用索引扫描，性能优异

3. **PriceTickDAO 性能**
   - 批量插入吞吐量：13,296 条/秒
   - 查询最新价格：< 3ms
   - 时间范围查询：< 2ms
   - ✅ 超过 5000 条/秒的目标

**测试环境**：
- PostgreSQL 15 + TimescaleDB 2.18
- Redis 7.4
- 本地开发环境

---

### 任务总结

#### [task_10_summary.md](./task_10_summary.md)
**任务 10: 读写分离配置 - 完成总结**

记录了读写分离功能的完整实现过程：

**完成内容**：
- ✅ 配置结构扩展（DatabaseConfig + ReplicaConfig）
- ✅ DBResolver 插件集成
- ✅ 读写分离逻辑实现
- ✅ 监控功能（复制状态、延迟监控）
- ✅ 8 个集成测试（100% 通过）
- ✅ 完整技术文档

**技术亮点**：
- 零侵入式设计
- 自动路由和降级
- 灵活配置（0 到多个从库）
- 监控完善

**相关文件**：
- `internal/config/config.go`
- `internal/database/read_write_splitting.go`
- `internal/database/connection.go`
- `internal/database/read_write_splitting_test.go`

---

## 📖 文档规范

### 命名规范
- 使用小写字母和下划线（snake_case）
- 格式：`<topic>_<type>.md`
  - 技术文档：`<topic>.md`（如 `read_write_splitting.md`）
  - 测试报告：`<scope>_test_report.md`（如 `performance_test_report.md`）
  - 任务总结：`task_<number>_summary.md`（如 `task_10_summary.md`）

### 文档结构
每个文档应包含：
- 清晰的标题和概述
- 目录（适用于长文档）
- 分段的详细内容
- 示例代码（如适用）
- 相关文件链接
- 完成时间和作者（如适用）

### 文档类型

#### 技术文档
记录系统架构、设计决策和实现细节：
- 概述和背景
- 架构设计
- 实现细节
- 配置说明
- 使用示例
- 最佳实践

#### 测试报告
记录测试结果和性能指标：
- 测试范围
- 测试环境
- 测试结果
- 性能指标
- 结论和建议

#### 任务总结
记录任务完成情况：
- 任务概述
- 完成的工作
- 技术亮点
- 测试结果
- 相关文件

---

## 🔄 更新记录

| 日期 | 文档 | 变更 |
|------|------|------|
| 2025-10-12 | 所有文档 | 统一命名规范，移至 docs/project/ 目录 |
| 2025-10-12 | read_write_splitting.md | 新增读写分离配置文档 |
| 2025-10-12 | performance_test_report.md | 新增性能测试报告 |
| 2025-10-12 | task_10_summary.md | 新增任务 10 完成总结 |

---

## 📝 待补充文档

建议后续添加以下文档：
- [ ] `database_schema.md` - 数据库表结构文档
- [ ] `api_design.md` - API 设计文档
- [ ] `deployment.md` - 部署指南
- [ ] `monitoring.md` - 监控和告警配置
- [ ] `troubleshooting.md` - 常见问题排查

---

**最后更新**: 2025-10-12

