# Facet — 功能开关管理平台

[![License](https://img.shields.io/badge/license-AGPLv3-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Node](https://img.shields.io/badge/Node-20+-339933?logo=node.js)](https://nodejs.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Facet 是一个开源的功能开关（Feature Flag）管理平台——LaunchDarkly 的自托管替代品。使用 Go 和 React 构建，提供集中式仪表盘，方便跨环境、租户和用户管理功能开关，支持安全渐进式发布和定向功能发布，无需供应商锁定。

## 为什么选择 Facet？

- **自托管 & 开源** — 完全掌控自己的数据。无需第三方 API 调用，无按席位计费。
- **原生多租户** — 专为需要为团队或客户隔离功能开关的平台而设计。
- **异步评估日志** — Redis 支持的管道，永不阻塞功能开关评估热路径。
- **SDK 无关** — 简单的 REST API，可使用任何 HTTP 客户端。官方提供 Go、Java 和 JavaScript 示例。

## 界面预览

### 租户管理

![租户管理](pictures/%E7%A7%9F%E6%88%B7%E7%AE%A1%E7%90%86.png)
![创建租户](pictures/%E5%88%9B%E5%BB%BA%E7%A7%9F%E6%88%B7.png)

### 用户管理

![用户列表](pictures/%E7%94%A8%E6%88%B7%E5%88%97%E8%A1%A8.png)
![创建用户](pictures/%E5%88%9B%E5%BB%BA%E7%94%A8%E6%88%B7.png)

### 环境管理

![环境列表](pictures/%E7%8E%AF%E5%A2%83%E5%88%97%E8%A1%A8.png)
![创建环境](pictures/%E5%88%9B%E5%BB%BA%E4%B8%80%E4%B8%AA%E7%8E%AF%E5%A2%83.png)

### 功能开关

![开关列表](pictures/%E5%BC%80%E5%85%B3%E5%88%97%E8%A1%A8.png)
![创建开关](pictures/%E5%88%9B%E5%BB%BA%E4%B8%80%E4%B8%AA%E5%BC%80%E5%85%B3.png)

### 系统通知

![系统通知列表](pictures/%E7%B3%BB%E7%BB%9F%E9%80%9A%E7%9F%A5%E5%88%97%E8%A1%A8.png)
![创建系统通知](pictures/%E5%88%9B%E5%BB%BA%E4%B8%80%E4%B8%AA%E7%B3%BB%E7%BB%9F%E9%80%9A%E7%9F%A5.png)

## 功能特性

- **多租户架构** — 为不同团队或客户提供隔离环境
- **环境管理** — 开发、测试、生产或任意自定义环境
- **功能开关 CRUD** — 创建、更新、删除、开关功能开关及定向规则
- **用户定向** — 基于账号和用户 ID 的白名单规则
- **基于角色的访问控制** — 系统管理员、租户管理员、租户观察者三种角色
- **验证码 & JWT 认证** — 安全的登录流程
- **系统通知** — 向所有租户广播公告
- **评估日志** — Redis → PostgreSQL 的异步日志处理管道
- **RESTful API** — 统一的 `{code, message, data}` JSON 响应格式
- **SDK 示例** — [Go、Java、JavaScript](facet-backend/docs/)（Node.js 和浏览器）

## 技术栈

| 层 | 技术 |
|-------|-----------|
| 后端 | Go, Gin, pgx/v5, Redis (go-redis/v9), Snowflake ID |
| 前端 | React 19, MUI v7, React Router v7, Zustand, Axios |
| 数据库 | PostgreSQL 17 |
| 缓存 | Redis 7 |
| 部署 | Docker Compose |

## 快速开始

```bash
# 克隆仓库
git clone https://github.com/anomalyco/facet.git
cd facet

# 启动所有服务（PostgreSQL、Redis、后端、前端）
docker compose up -d

# 访问控制台
# 前端: http://localhost
# 后端 API: http://localhost:8080
```

**默认登录：** 启动后系统没有预置用户账号。需要通过数据库或 API 创建初始管理员用户。详见后端开发文档。

## 项目结构

```
facet/
├── facet-backend/                 # Go 后端
│   ├── cmd/server/                # 入口
│   ├── internal/
│   │   ├── api/                   # HTTP 处理器、中间件、路由
│   │   ├── cache/                 # Redis 缓存层
│   │   ├── domain/                # 领域类型与接口
│   │   ├── repository/            # 数据访问层（PostgreSQL）
│   │   ├── server/                # 依赖注入与生命周期
│   │   ├── service/               # 业务逻辑
│   │   └── worker/                # 后台工作进程
│   ├── internal/repository/postgres/migrations/  # 数据库迁移
│   └── docs/                      # API 文档与 SDK 示例
├── facet-front/                   # React 前端
│   ├── website/                   # 营销官网
│   └── src/
│       ├── api/                   # API 客户端模块
│       ├── layouts/               # 布局组件
│       ├── router/                # 路由定义
│       ├── stores/                # Zustand 状态管理
│       ├── utils/                 # 工具与格式化函数
│       └── views/                 # 页面组件
├── docker-compose.yml             # Docker Compose 配置
└── LICENSE
```

## 配置

关键环境变量（完整列表见 `.env.example`）：

| 变量 | 说明 |
|----------|-------------|
| `POSTGRES_DSN` | PostgreSQL 连接字符串 |
| `JWT_SECRET_KEY` | JWT 签名密钥 |
| `JWT_EXPIRES_IN_HOURS` | Token 过期时间（默认: 24） |
| `REDIS_ADDRESS` | Redis 服务器地址 |
| `REDIS_PASSWORD` | Redis 密码 |
| `LOGGER_LEVEL` | 日志级别（debug, info, warn, error） |

## API 文档

完整的 API 参考见 [API_DOC.md](facet-backend/API_DOC.md)。服务器运行时可访问 Swagger UI：`/swagger/index.html`（后端端口）。

基础路径: `/api/v1`

- `GET /evaluate` — SDK 评估接口（无需认证）
- `/manage/auth/*` — 登录与验证码
- `/manage/*` — 租户管理（JWT 认证）
- `/system/*` — 系统管理（需要 system_admin 角色）

## SDK 示例

开箱即用的 SDK 客户端示例位于 [`facet-backend/docs/example/`](facet-backend/docs/example/)：

| 语言 | 文件 | 最低版本 |
|----------|------|-----------------|
| Go | [`after_1_18.go`](facet-backend/docs/example/go/after_1_18.go) | 1.18+ |
| Java | [`Java17Demo.java`](facet-backend/docs/example/java/Java17Demo.java) | 17+ |
| JavaScript (Node.js) | [`after_node_18.js`](facet-backend/docs/example/JavaScript/after_node_18.js) | 18+ |
| JavaScript (浏览器) | [`web.js`](facet-backend/docs/example/JavaScript/web.js) | 现代浏览器 |

## 开发

### 前置条件

- Go 1.25+
- Node.js 20+
- PostgreSQL 17
- Redis 7

### 本地搭建

```bash
# 启动依赖服务
docker compose up -d postgres redis

# 后端
cd facet-backend
cp .env.example .env
go run ./cmd/server/main.go

# 前端（另开终端）
cd facet-front
npm install
npm run dev
```

## 贡献

欢迎贡献代码！请先阅读 [CONTRIBUTING.md](CONTRIBUTING.md)。

## 许可证

[AGPL v3](LICENSE)
