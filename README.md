# CryptoSignal Hunter

> 加密货币合约交易信号捕捉系统

**CryptoSignal Hunter** 是一个专为个人交易者设计的加密货币合约交易信号捕捉系统，能够实时监控并捕捉由突发消息面引起的短期价格异常波动，从而把握合约交易套利机会。

## ✨ 核心特性

-  **实时监控** - 监控BitGet交易所上百个合约交易对的实时价格
-  **智能识别** - 自动识别符合预设条件的价格异常波动
-  **可视化界面** - 直观的Web界面，实时展示市场动态和交易信号
-  **灵活配置** - 支持自定义过滤条件和告警阈值
-  **历史回测** - 基于历史数据验证交易策略的可行性

## 🛠 技术栈

### 后端

- **Golang 1.21+** - 高性能后端服务
- **Gin** - Web框架
- **PostgreSQL 15 + TimescaleDB** - 时序数据存储
- **Redis 7** - 实时数据缓存
- **Zap** - 结构化日志
- **Swagger** - API文档

### 前端

- **React 18+** - UI框架
- **TypeScript 5+** - 类型安全
- **Vite 5+** - 快速构建工具
- **Ant Design 5+** - UI组件库
- **Tailwind CSS 3+** - 样式框架
- **Zustand** - 状态管理
- **Lightweight Charts** - 图表库

### 开发工具

- **Air** - Go热重载
- **golangci-lint** - Go代码检查
- **ESLint + Prettier** - 前端代码规范
- **Husky + lint-staged** - Git提交前检查
- **Docker Compose** - 本地开发环境

## 📋 环境要求

- **Go** 1.21+
- **Node.js** 18+
- **pnpm** 8+
- **Docker** & **Docker Compose**
- **Git**

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone https://github.com/haxrd/cryptosignal-hunter.git
cd cex-arbitrage
```

### 2. 启动数据库服务

```bash
# 启动PostgreSQL和Redis
docker-compose up -d

# 验证服务状态
docker-compose ps
```

### 3. 启动后端服务

```bash
cd backend

# 安装依赖
go mod download

# 启动开发服务器（热重载）
air

# 或者使用Makefile
make run
```

后端服务将在 `http://localhost:8080` 启动

- **健康检查**: http://localhost:8080/health
- **API文档**: http://localhost:8080/swagger/index.html

### 4. 启动前端服务

```bash
cd frontend

# 安装依赖
pnpm install

# 启动开发服务器
pnpm dev
```

前端服务将在 `http://localhost:3000` 启动，浏览器会自动打开。

## 📖 开发命令

### 后端命令

```bash
cd backend

make run          # 启动服务（热重载）
make build        # 编译二进制文件
make lint         # 运行代码检查
make swagger      # 生成Swagger文档
make test         # 运行测试
make clean        # 清理构建产物
make fmt          # 格式化代码
make help         # 显示帮助信息
```

### 前端命令

```bash
cd frontend

pnpm dev          # 启动开发服务器
pnpm build        # 生产构建
pnpm lint         # ESLint检查
pnpm format       # Prettier格式化
pnpm preview      # 预览生产构建
```

### Docker命令

```bash
# 启动所有服务
docker-compose up -d

# 停止所有服务
docker-compose down

# 查看服务日志
docker-compose logs -f

# 查看PostgreSQL日志
docker-compose logs -f postgres

# 查看Redis日志
docker-compose logs -f redis

# 进入PostgreSQL容器
docker exec -it cryptosignal-postgres psql -U postgres -d cryptosignal

# 进入Redis容器
docker exec -it cryptosignal-redis redis-cli
```

## 📁 项目结构

```
cex-arbitrage/
├── backend/                 # Go后端服务
│   ├── cmd/                 # 应用入口
│   │   └── server/          
│   │       └── main.go      
│   ├── internal/            # 内部包
│   │   ├── api/             # API路由和处理器
│   │   ├── config/          # 配置管理
│   │   ├── middleware/      # 中间件
│   │   └── models/          # 数据模型
│   ├── pkg/                 # 公共包
│   ├── docs/                # Swagger文档
│   ├── config.yaml          # 配置文件
│   ├── .air.toml            # Air配置
│   ├── .golangci.yml        # Linter配置
│   ├── Makefile             # 构建脚本
│   └── go.mod               
├── frontend/                # React前端
│   ├── src/                 
│   │   ├── api/             # API调用
│   │   ├── components/      # 可复用组件
│   │   ├── layouts/         # 布局组件
│   │   ├── pages/           # 页面组件
│   │   ├── store/           # 状态管理
│   │   ├── styles/          # 全局样式
│   │   ├── types/           # 类型定义
│   │   ├── utils/           # 工具函数
│   │   ├── App.tsx          
│   │   └── main.tsx         
│   ├── public/              
│   ├── .eslintrc.cjs        # ESLint配置
│   ├── .prettierrc          # Prettier配置
│   ├── tailwind.config.js   # Tailwind配置
│   ├── vite.config.ts       # Vite配置
│   └── package.json         
├── scripts/                 # 脚本文件
│   └── init-db.sql          # 数据库初始化
├── .husky/                  # Git Hooks
│   └── pre-commit           
├── docker-compose.yml       # Docker Compose配置
└── README.md                

```

## ❓ 常见问题 (FAQ)

### 1. 无法连接数据库

**问题**: 后端启动时提示无法连接数据库

**解决方案**:
```bash
# 检查Docker服务是否运行
docker-compose ps

# 查看PostgreSQL日志
docker-compose logs postgres

# 重启数据库服务
docker-compose restart postgres
```

### 2. 前端启动失败

**问题**: 前端依赖安装失败或启动报错

**解决方案**:
```bash
# 清理node_modules和lockfile
cd frontend
rm -rf node_modules pnpm-lock.yaml

# 重新安装依赖
pnpm install

# 如果仍然失败，尝试清理pnpm缓存
pnpm store prune
pnpm install
```

### 3. 端口被占用

**问题**: 启动服务时提示端口已被占用

**解决方案**:
```bash
# 查找占用端口的进程（以8080为例）
lsof -i :8080

# 终止进程（替换<PID>为实际进程ID）
kill -9 <PID>

# 或者修改配置文件使用其他端口
# backend/config.yaml: server.port
# frontend/vite.config.ts: server.port
```

### 4. Swagger文档无法访问

**问题**: 访问 `/swagger/index.html` 返回404

**解决方案**:
```bash
cd backend

# 重新生成Swagger文档
make swagger

# 或手动生成
swag init -g cmd/server/main.go -o docs

# 重启后端服务
make run
```

### 5. Git提交被拒绝

**问题**: 提交代码时被pre-commit hook阻止

**解决方案**:
```bash
# 运行代码格式化
cd frontend
pnpm format

# 修复lint错误
pnpm lint --fix

# 再次尝试提交
git commit -m "your message"
```

## 📝 开发规范

### Git提交规范

```
feat: 新功能
fix: 修复bug
docs: 文档更新
style: 代码格式调整
refactor: 代码重构
test: 测试相关
chore: 构建/工具链相关
```

### 代码规范

- 后端代码遵循 `golangci-lint` 规则
- 前端代码遵循 `ESLint` + `Prettier` 规则
- 所有代码提交前会自动运行lint检查
- 确保代码通过所有检查后再提交

## 📄 许可证

MIT License

## 🤝 贡献

欢迎提交Issue和Pull Request！

---

**开发状态**: 🚧 开发环境已就绪，核心功能开发中...

