# Spec Tasks

## 任务清单

基于规范 `2025-10-07-project-initialization` 的实施任务。

---

- [x] 1. 创建后端项目结构和基础配置
  - [x] 1.1 初始化Go项目目录结构（backend/cmd、internal、pkg等）
  - [x] 1.2 创建 `go.mod` 并添加核心依赖（Gin、Zap、Viper等）
  - [x] 1.3 实现配置管理模块（config/config.go + config.yaml）
  - [x] 1.4 实现基础中间件（CORS、Logger、Recovery）
  - [x] 1.5 配置Gin路由和健康检查端点（GET /health）
  - [x] 1.6 集成Swagger文档生成（gin-swagger + swaggo）
  - [x] 1.7 创建 `.air.toml` 配置热重载工具
  - [x] 1.8 创建 `.golangci.yml` 配置代码检查工具
  - [x] 1.9 创建 `Makefile` 封装常用命令
  - [x] 1.10 验证后端服务能够正常启动并访问 `/health` 和 `/swagger/index.html`

- [x] 2. 创建前端项目结构和基础配置
  - [x] 2.1 使用Vite创建React + TypeScript项目（frontend/）
  - [x] 2.2 安装核心依赖（React Router、Ant Design、Zustand等）
  - [x] 2.3 配置Tailwind CSS（tailwind.config.js + postcss.config.js）
  - [x] 2.4 配置Vite（代理设置、端口3000、环境变量支持）
  - [x] 2.5 配置TypeScript（tsconfig.json 严格模式）
  - [x] 2.6 配置ESLint（.eslintrc.cjs，启用推荐规则）
  - [x] 2.7 配置Prettier（.prettierrc + .prettierignore）
  - [x] 2.8 配置Husky + lint-staged（pre-commit自动格式化）
  - [x] 2.9 创建基础目录结构（src/api、components、pages、store等）
  - [x] 2.10 实现初始App.tsx（展示欢迎页面和Ant Design组件验证）
  - [x] 2.11 配置package.json scripts（dev、build、lint、format）
  - [x] 2.12 验证前端开发服务器能够正常启动并显示欢迎页面

- [x] 3. 配置Docker Compose本地开发环境
  - [x] 3.1 创建项目根目录的 `docker-compose.yml` 文件
  - [x] 3.2 配置PostgreSQL 15 + TimescaleDB服务（端口5432、环境变量、数据卷）
  - [x] 3.3 配置Redis 7服务（端口6379、数据卷、AOF持久化）
  - [x] 3.4 创建自定义网络 `cryptosignal-network`
  - [x] 3.5 创建数据库初始化脚本 `scripts/init-db.sql`（启用TimescaleDB扩展）
  - [x] 3.6 配置健康检查（pg_isready、redis-cli ping）
  - [x] 3.7 验证 `docker-compose up -d` 能够成功启动所有服务
  - [x] 3.8 验证能够连接到PostgreSQL并查询TimescaleDB扩展信息
  - [x] 3.9 验证能够连接到Redis并执行基本操作

- [x] 4. 配置开发工具和代码规范
  - [x] 4.1 安装并配置后端热重载工具air（确保文件变更自动重启）
  - [x] 4.2 配置golangci-lint（启用gofmt、govet、errcheck等linter）
  - [x] 4.3 配置前端Git Hooks（Husky初始化和pre-commit hook）
  - [x] 4.4 测试代码规范自动化（提交不符合规范的代码应被阻止）
  - [x] 4.5 验证后端修改代码后air自动重新编译
  - [x] 4.6 验证前端修改代码后HMR正常工作
  - [x] 4.7 运行 `make lint` 验证后端代码检查工作正常
  - [x] 4.8 运行 `pnpm lint` 验证前端代码检查工作正常

- [x] 5. 创建项目文档和最终验证
  - [x] 5.1 创建项目根目录 `README.md`（包含项目简介、技术栈、环境要求）
  - [x] 5.2 在README中添加快速开始指南（克隆、启动数据库、启动后端、启动前端）
  - [x] 5.3 在README中添加开发命令说明（后端和前端常用命令）
  - [x] 5.4 在README中添加项目结构说明
  - [x] 5.5 在README中添加常见问题（FAQ）
  - [x] 5.6 创建 `.gitignore` 文件（排除node_modules、vendor、dist等）
  - [x] 5.7 创建 `.env.example` 文件（前后端环境变量模板）
  - [x] 5.8 执行完整的端到端验证（按README步骤从零开始启动整个开发环境）
  - [x] 5.9 验证所有交付成果（5个预期交付项）
  - [x] 5.10 更新路线图标记"项目初始化"任务为已完成

---

## 任务说明

**任务总数：** 5个主要任务，共47个子任务

**预计工作量：** XS（1天以内）

**依赖关系：**
- 任务1和任务2可以并行执行
- 任务3依赖任务1（数据库连接信息）
- 任务4依赖任务1和任务2
- 任务5依赖所有前置任务完成

**验证标准：**
按照规范中的"Expected Deliverable"部分，所有5项交付成果必须能够成功验证。

