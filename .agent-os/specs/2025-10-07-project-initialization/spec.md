# Spec Requirements Document

> Spec: 项目初始化 - 搭建基础开发环境
> Created: 2025-10-07

## Overview

搭建CryptoSignal Hunter项目的基础开发环境，包括Go后端项目、React前端项目、Docker Compose本地开发环境（PostgreSQL + Redis），以及配置必要的开发工具（代码规范、热重载、API文档），为后续功能开发提供稳定的技术基础。

## User Stories

### 后端开发者快速启动开发

作为一名后端开发者，我希望能够快速搭建Go项目并启动本地开发环境，以便我可以立即开始编写业务逻辑，而无需花费时间在环境配置上。

开发者克隆代码库后，通过简单的命令（如 `docker-compose up` 和 `air`）即可启动完整的后端开发环境，包括数据库、缓存和热重载支持。代码修改后自动重新编译，无需手动重启服务，大幅提升开发效率。

### 前端开发者快速启动开发

作为一名前端开发者，我希望能够快速搭建React项目并连接后端API，以便我可以立即开始开发用户界面。

开发者通过 `pnpm install` 和 `pnpm dev` 即可启动前端开发服务器，支持热模块替换（HMR），代码保存后浏览器自动刷新。TypeScript、ESLint和Prettier已配置完成，确保代码质量。

### 团队协作规范统一

作为一名团队成员，我希望有统一的代码规范和Git提交流程，以便团队代码风格能够保持一致，并减少代码审查中的格式问题。

所有开发者在提交代码前，Git Hooks会自动运行代码格式化和lint检查，不符合规范的代码无法提交。团队成员使用相同的工具配置，确保代码库的整洁和一致性。

## Spec Scope

1. **Go后端项目结构** - 创建标准的Go项目结构，配置Go Modules依赖管理，集成Gin框架和基础中间件
2. **React前端项目结构** - 创建基于Vite的React + TypeScript项目，集成Ant Design和Tailwind CSS
3. **Docker Compose本地环境** - 配置PostgreSQL 15（含TimescaleDB扩展）和Redis 7的Docker容器，支持一键启动
4. **后端开发工具配置** - 配置golangci-lint代码检查、air热重载工具、Swagger API文档生成
5. **前端开发工具配置** - 配置ESLint、Prettier、Husky和lint-staged，实现代码规范自动化

## Out of Scope

- AWS生产环境配置（后续Phase部署时处理）
- CI/CD Pipeline配置（后续阶段添加）
- 具体业务功能实现（如API接口、数据库表结构设计）
- 前端路由和状态管理的具体实现（框架搭好即可，具体实现在功能开发中）

## Expected Deliverable

1. 开发者可以通过 `docker-compose up -d` 启动PostgreSQL和Redis，通过 `air` 启动后端服务并支持热重载
2. 开发者可以通过 `pnpm install && pnpm dev` 启动前端开发服务器，浏览器访问看到默认页面
3. Git提交时自动运行代码格式化和lint检查，不符合规范的代码会被阻止提交
4. 后端访问 `/swagger/index.html` 可以看到Swagger API文档界面（即使暂无接口）
5. 项目包含完整的README.md文档，说明环境要求、安装步骤和开发命令

