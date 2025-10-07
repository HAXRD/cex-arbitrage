# 技术栈

## 技术架构概览

本项目采用高性能、轻量化的技术栈，专注于实时数据处理和可视化展示。遵循奥卡姆剃刀原则，避免过度设计，优先保证核心功能的稳定性和性能。

## 后端技术

### 应用框架
- **Golang 1.21+**
  - 选择理由：高性能、原生并发支持、适合处理大量实时数据流
  - 使用框架：Gin（轻量级Web框架）+ Gorilla WebSocket
  - 用途：API服务、WebSocket实时推送、数据采集和处理

### 数据存储

#### 主数据库
- **PostgreSQL 15+ with TimescaleDB 扩展**
  - 选择理由：成熟稳定、TimescaleDB专为时序数据优化、适合存储历史价格数据
  - 用途：
    - 历史价格数据存储（K线数据、Tick数据）
    - 用户配置和策略存储
    - 回测结果存储
  - 表设计：
    - `price_ticks`（时序表）：存储实时价格数据
    - `klines`（时序表）：存储K线数据（1m, 5m, 15m等）
    - `symbols`：交易对配置
    - `strategies`：用户策略配置
    - `backtest_results`：回测结果

#### 缓存层
- **Redis 7+**
  - 选择理由：高性能、支持多种数据结构
  - 用途：
    - 实时价格缓存（最新价格、24h涨跌幅）
    - WebSocket连接管理
    - 限流控制
    - 实时指标计算缓存

### 第三方集成
- **BitGet API**
  - 合约价格数据获取
  - WebSocket实时行情订阅

## 前端技术

### JavaScript框架
- **React 18+**
  - 选择理由：生态成熟、组件化开发、性能优异
  - 状态管理：Zustand（轻量级状态管理）
  - 路由：React Router v6

### 构建工具
- **Vite 5+**
  - 选择理由：快速的开发服务器、优化的生产构建
  - 配置：ESBuild + Rollup

### UI组件库
- **Ant Design 5+**
  - 选择理由：专业的企业级UI组件库、丰富的组件、良好的TypeScript支持
  - 用途：数据表格、表单、布局、弹窗等

### 图表库
- **Lightweight Charts (TradingView)**
  - 选择理由：轻量、高性能、专为金融图表设计
  - 用途：K线图、实时价格图表

### 样式方案
- **Tailwind CSS 3+**
  - 选择理由：实用优先、快速开发、体积小
  - 配置：与Ant Design结合使用

### 实时通信
- **WebSocket (native)**
  - 用途：接收后端实时价格推送
  - 自动重连机制

### 开发语言
- **TypeScript 5+**
  - 选择理由：类型安全、更好的IDE支持、减少运行时错误

## 开发工具

### 版本控制
- **Git + GitHub**
  - 代码仓库：待初始化
  - 分支策略：main（生产）、develop（开发）、feature/*（特性）

### 包管理器
- **后端：** Go Modules
- **前端：** pnpm
  - 选择理由：比npm/yarn更快、节省磁盘空间

### 代码规范
- **后端：** golangci-lint
- **前端：** ESLint + Prettier
- **Git Hooks：** Husky + lint-staged

### API文档
- **Swagger/OpenAPI 3.0**
  - 工具：swaggo/swag
  - 用途：自动生成API文档

## 部署架构

### 开发环境
- **本地开发**
  - Docker Compose：统一开发环境（PostgreSQL + Redis）
  - 前端：Vite Dev Server
  - 后端：直接运行Go应用（支持热重载：air）

### 生产环境（AWS）

#### 计算资源
- **后端服务：** AWS EC2 (t3.medium)
  - 或使用 AWS ECS（容器化部署）
  - 使用 Systemd 或 Docker 管理服务

#### 数据库
- **PostgreSQL：** AWS RDS for PostgreSQL
  - 实例类型：db.t3.medium（初期）
  - 启用 TimescaleDB 扩展
  - 自动备份

#### 缓存
- **Redis：** AWS ElastiCache for Redis
  - 节点类型：cache.t3.micro（初期）

#### 前端托管
- **AWS S3 + CloudFront**
  - S3存储静态文件
  - CloudFront CDN加速

#### 网络
- **AWS VPC**
  - 公有子网：API网关、前端
  - 私有子网：数据库、Redis

#### 域名与SSL
- **域名：** Route 53（待配置）
- **SSL证书：** AWS Certificate Manager

### 部署策略
- **容器化：** Docker + Docker Compose
- **镜像仓库：** AWS ECR 或 Docker Hub
- **CI/CD：** GitHub Actions
  - 自动测试
  - 自动构建镜像
  - 自动部署（初期手动部署）

## 监控与日志

### 日志
- **后端：** Zap（结构化日志）
- **日志聚合：** 初期本地文件，后期考虑 AWS CloudWatch

### 监控
- **系统监控：** 初期手动检查，后期考虑 Prometheus + Grafana
- **告警：** 初期邮件告警

## 数据流架构

```
BitGet WebSocket API
        ↓
[Go 数据采集服务]
        ↓
    [Redis缓存] → [WebSocket Server] → [前端实时展示]
        ↓
[TimescaleDB存储]
        ↓
 [历史数据查询/回测]
```

## 技术选型原则

1. **性能优先**：选择高性能方案（Go、TimescaleDB、Redis）以支持上百个币种的实时监控
2. **简单至上**：遵循奥卡姆剃刀原则，避免过度设计（单体架构，按需扩展）
3. **成熟稳定**：选择经过验证的技术栈，降低开发风险
4. **MVP导向**：优先实现核心功能，非必要功能后置（如高级告警、自动交易等）
5. **可扩展性**：预留扩展空间，但不提前优化（如微服务化可后期考虑）

## 预估资源成本（AWS月度）

**开发/测试阶段（本地）：** $0

**生产环境（MVP阶段）：**
- EC2 (t3.medium): ~$30/月
- RDS (db.t3.medium): ~$50/月
- ElastiCache (cache.t3.micro): ~$15/月
- S3 + CloudFront: ~$5/月
- 其他（流量、备份等）: ~$10/月
- **总计：** ~$110/月

**注：** 初期以本地开发测试为主，待MVP验证通过后再部署到AWS生产环境。

