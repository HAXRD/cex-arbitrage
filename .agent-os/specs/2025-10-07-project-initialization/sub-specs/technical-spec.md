# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-07-project-initialization/spec.md

## Technical Requirements

### 后端项目结构

**Go版本要求：** 1.21+

**项目目录结构：**
```
backend/
├── cmd/
│   └── server/
│       └── main.go           # 应用入口
├── internal/
│   ├── api/                  # API路由和处理器
│   │   ├── router.go
│   │   └── handlers/
│   ├── config/               # 配置管理
│   │   └── config.go
│   ├── middleware/           # 中间件
│   │   ├── cors.go
│   │   └── logger.go
│   └── models/               # 数据模型（预留）
├── pkg/                      # 可复用的公共包（预留）
├── docs/                     # Swagger生成的文档
├── .air.toml                 # air热重载配置
├── .golangci.yml             # golangci-lint配置
├── go.mod
├── go.sum
└── Makefile                  # 常用命令封装
```

**核心依赖：**
- `github.com/gin-gonic/gin` - Web框架
- `github.com/swaggo/gin-swagger` - Swagger集成
- `github.com/swaggo/swag` - Swagger文档生成
- `github.com/spf13/viper` - 配置管理
- `go.uber.org/zap` - 结构化日志
- `github.com/lib/pq` - PostgreSQL驱动（预装，暂不使用）
- `github.com/redis/go-redis/v9` - Redis客户端（预装，暂不使用）

**Gin基础配置：**
- 启用CORS中间件（支持前端跨域请求）
- 启用请求日志中间件（使用Zap）
- 启用Recovery中间件（panic恢复）
- 配置健康检查端点：`GET /health`
- 配置Swagger文档端点：`GET /swagger/*`

**配置文件：** `config.yaml`
```yaml
server:
  port: 8080
  mode: debug  # debug/release

database:
  host: localhost
  port: 5432
  user: postgres
  password: postgres
  dbname: cryptosignal
  sslmode: disable

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0
```

**热重载配置（.air.toml）：**
- 监听 `.go` 文件变化
- 排除 `vendor/`、`tmp/`、`docs/` 目录
- 延迟重启时间：100ms
- 自动清理旧进程

**代码规范配置（.golangci.yml）：**
启用以下linters：
- `gofmt` - 代码格式化
- `goimports` - import排序
- `govet` - Go静态分析
- `errcheck` - 错误检查
- `staticcheck` - 静态分析
- `unused` - 未使用代码检测
- `gosimple` - 代码简化建议

---

### 前端项目结构

**Node版本要求：** 18+
**包管理器：** pnpm 8+

**项目目录结构：**
```
frontend/
├── public/                   # 静态资源
├── src/
│   ├── api/                  # API调用封装
│   ├── assets/               # 图片、字体等资源
│   ├── components/           # 可复用组件
│   ├── layouts/              # 布局组件（预留）
│   ├── pages/                # 页面组件（预留）
│   ├── store/                # Zustand状态管理（预留）
│   ├── styles/               # 全局样式
│   │   └── globals.css
│   ├── types/                # TypeScript类型定义
│   ├── utils/                # 工具函数
│   ├── App.tsx               # 应用根组件
│   ├── main.tsx              # 应用入口
│   └── vite-env.d.ts
├── .eslintrc.cjs             # ESLint配置
├── .prettierrc               # Prettier配置
├── .prettierignore
├── tailwind.config.js        # Tailwind配置
├── postcss.config.js         # PostCSS配置
├── tsconfig.json             # TypeScript配置
├── tsconfig.node.json
├── vite.config.ts            # Vite配置
└── package.json
```

**核心依赖：**
- `react` / `react-dom` ^18.3.0
- `react-router-dom` ^6.26.0 - 路由
- `antd` ^5.20.0 - UI组件库
- `zustand` ^4.5.0 - 状态管理
- `lightweight-charts` ^4.1.0 - 图表库
- `axios` ^1.7.0 - HTTP客户端
- `dayjs` ^1.11.0 - 日期处理

**开发依赖：**
- `vite` ^5.4.0
- `@vitejs/plugin-react` ^4.3.0
- `typescript` ^5.5.0
- `tailwindcss` ^3.4.0
- `eslint` ^8.57.0
- `prettier` ^3.3.0
- `husky` ^9.1.0
- `lint-staged` ^15.2.0

**Vite配置要点：**
- 开发服务器端口：3000
- 代理配置：`/api` 代理到 `http://localhost:8080`
- 自动打开浏览器
- 支持环境变量（`.env.development`、`.env.production`）

**ESLint配置：**
- 继承 `eslint:recommended`
- 继承 `plugin:@typescript-eslint/recommended`
- 继承 `plugin:react-hooks/recommended`
- 规则：禁用 `console.log`（仅警告）

**Prettier配置：**
```json
{
  "semi": true,
  "singleQuote": true,
  "tabWidth": 2,
  "trailingComma": "es5",
  "printWidth": 100
}
```

**Tailwind配置：**
- 与Ant Design主题兼容
- 自定义颜色变量（预留）
- 响应式断点保持默认

**Husky + lint-staged配置：**
- Pre-commit Hook：运行 `lint-staged`
- `lint-staged` 规则：
  - `*.{ts,tsx}`: `eslint --fix`, `prettier --write`
  - `*.{css,json,md}`: `prettier --write`

**初始App.tsx内容：**
- 展示一个简单的欢迎页面
- 使用Ant Design的 `<Button>` 和 `<Card>` 组件验证UI库正常工作
- 显示"CryptoSignal Hunter - 开发环境已就绪"

---

### Docker Compose配置

**文件位置：** 项目根目录 `docker-compose.yml`

**服务定义：**

1. **PostgreSQL 15 + TimescaleDB**
   - 镜像：`timescale/timescaledb:latest-pg15`
   - 端口：`5432:5432`
   - 环境变量：
     - `POSTGRES_USER=postgres`
     - `POSTGRES_PASSWORD=postgres`
     - `POSTGRES_DB=cryptosignal`
   - 数据卷：`postgres_data:/var/lib/postgresql/data`
   - 健康检查：`pg_isready -U postgres`
   - 初始化脚本：`./scripts/init-db.sql`（启用TimescaleDB扩展）

2. **Redis 7**
   - 镜像：`redis:7-alpine`
   - 端口：`6379:6379`
   - 数据卷：`redis_data:/data`
   - 健康检查：`redis-cli ping`
   - 持久化：启用AOF

**数据卷：**
- `postgres_data` - PostgreSQL数据持久化
- `redis_data` - Redis数据持久化

**网络：**
- 创建自定义网络 `cryptosignal-network`
- 所有服务连接到该网络

**初始化脚本（scripts/init-db.sql）：**
```sql
-- 启用TimescaleDB扩展
CREATE EXTENSION IF NOT EXISTS timescaledb;

-- 验证扩展
SELECT extname, extversion FROM pg_extension WHERE extname = 'timescaledb';
```

---

### 文档要求

**README.md 必须包含：**

1. 项目简介
2. 技术栈列表
3. 环境要求（Go 1.21+, Node 18+, Docker, pnpm）
4. 快速开始：
   - 克隆代码
   - 启动数据库：`docker-compose up -d`
   - 启动后端：`cd backend && go mod download && air`
   - 启动前端：`cd frontend && pnpm install && pnpm dev`
5. 开发命令说明
6. 项目结构说明
7. 常见问题（FAQ）

**后端Makefile常用命令：**
- `make run` - 启动服务（使用air）
- `make build` - 编译二进制
- `make lint` - 运行golangci-lint
- `make swagger` - 生成Swagger文档
- `make test` - 运行测试（预留）

**前端package.json scripts：**
- `pnpm dev` - 启动开发服务器
- `pnpm build` - 生产构建
- `pnpm lint` - ESLint检查
- `pnpm format` - Prettier格式化
- `pnpm preview` - 预览生产构建

---

### 性能和质量标准

- **后端启动时间：** < 3秒
- **前端热更新响应：** < 1秒
- **代码覆盖率目标：** 暂不要求（后续阶段添加测试）
- **Docker Compose启动时间：** < 30秒（首次拉取镜像除外）
- **代码规范覆盖率：** 100%（所有代码必须通过lint检查才能提交）

---

## 外部依赖

所有依赖已在上述技术要求中详细列出，无额外的外部依赖需求。所有工具和库均为开源且成熟稳定的技术栈。

