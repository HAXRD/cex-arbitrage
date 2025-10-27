# Technical Specification

This is the technical specification for the spec detailed in @.agent-os/specs/2025-10-25-frontend-framework/spec.md

## Technical Requirements

### 项目初始化
- 使用Vite 5+创建React 18 + TypeScript项目
- 配置ESBuild和Rollup构建工具
- 设置开发服务器和热重载功能
- 配置生产环境优化构建

### UI框架集成
- 集成Ant Design 5+组件库，配置主题和国际化
- 集成Tailwind CSS 3+，配置与Ant Design的兼容性
- 建立统一的设计系统，包括颜色、字体、间距等规范
- 实现响应式布局，支持桌面端和移动端

### 状态管理
- 使用Zustand进行轻量级状态管理
- 实现全局状态和组件级状态的合理分离
- 支持状态持久化（localStorage/sessionStorage）
- 实现状态订阅和更新机制

### 路由配置
- 使用React Router v6配置页面路由
- 实现路由守卫和权限控制基础结构
- 支持动态路由和嵌套路由
- 配置路由懒加载和代码分割

### WebSocket集成
- 实现WebSocket客户端连接管理
- 支持自动重连和心跳检测
- 实现消息订阅和取消订阅机制
- 处理连接状态和错误处理

### 图表组件
- 集成TradingView Lightweight Charts
- 实现K线图组件，支持多种时间周期
- 实现实时价格图表组件
- 支持图表交互（缩放、平移、十字线等）

### 开发工具配置
- 配置ESLint和Prettier代码格式化
- 设置Husky和lint-staged Git钩子
- 配置TypeScript严格模式
- 设置路径别名和模块解析

### 性能优化
- 实现组件懒加载和代码分割
- 配置图片和资源优化
- 实现虚拟滚动（用于大量数据列表）
- 优化WebSocket数据处理性能

## External Dependencies

### 核心依赖
- **React 18+** - 前端框架
- **TypeScript 5+** - 类型安全
- **Vite 5+** - 构建工具
- **Ant Design 5+** - UI组件库
- **Tailwind CSS 3+** - 样式框架

### 状态管理
- **Zustand 4+** - 轻量级状态管理
- **React Router v6** - 路由管理

### 图表库
- **TradingView Lightweight Charts** - 金融图表组件

### WebSocket
- **原生WebSocket API** - 实时通信

### 开发工具
- **ESLint** - 代码检查
- **Prettier** - 代码格式化
- **Husky** - Git钩子
- **lint-staged** - 暂存区代码检查

### 构建工具
- **pnpm** - 包管理器
- **@types/node** - Node.js类型定义
- **@types/react** - React类型定义
- **@types/react-dom** - React DOM类型定义

### 工具库
- **dayjs** - 日期处理
- **lodash-es** - 工具函数库
- **clsx** - 条件类名处理
