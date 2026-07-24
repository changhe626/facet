# 我开源了一个 LaunchDarkly 替代品：Facet

> 功能开关平台，Go + React，自托管

如果你所在的公司正在使用功能开关（Feature Flag），那你大概率听过 LaunchDarkly——功能开关领域的标杆产品。它很好用，但价格也同样"标杆"：按席位收费，量大了之后成本蹭蹭往上涨。更关键的是，你的所有评估数据都要经过第三方服务器，这对很多企业来说是个合规顾虑。

所以我写了一个开源的替代品：**Facet**。

## 不是一个简单的开关

Facet 不是一个"一个人花两周写出来的 side project"。它的目标是从第一天起就承担真实的生产流量。

整个项目的架构是这样的：

```
后端：Go 1.25 + Gin + pgx/v5 + Redis
前端：React 19 + MUI v7 + TypeScript
数据库：PostgreSQL 17
部署：Docker Compose
```

核心链路是：**SDK 请求 → 路由分发 → 规则引擎匹配 → 异步日志 → Redis → PostgreSQL 批量写入**。评估路径上不写数据库，只从 Redis 缓存读数据，所以即使流量暴增，也不会对评估延迟产生影响。

## 为什么我自己写了一个？

市面上其实已经有几个开源的 Feature Flag 方案，比如 Unleash 和 Flagsmith。但在调研之后，我发现它们在一些关键点上不太满足我的需求：

**Unleash** 功能很强，但它的评估路径依赖 PostgreSQL，在高并发场景下需要额外的缓存层来保证延迟。

**Flagsmith** 支持自托管，但它用 Django 写的，如果团队主要是 Go/Node 技术栈，出了问题不好排查和改。

我想要的是：

1. **评估路径极快** —— 从 Redis 读，不查数据库
2. **技术栈统一** —— Go + React，团队接手成本低
3. **原生多租户** —— 不是靠"项目"来模拟隔离，而是从数据模型层面就设计了租户体系
4. **评估日志不丢** —— 异步写入，即使 DB 抖动也不阻塞 SDK

Facet 就是围绕这四点设计的。

## 核心功能一览

- **多租户管理**：每个租户有独立的环境、开关、用户。适合 SaaS 平台为自己的客户提供功能开关能力。
- **环境管理**：开发、测试、生产，或者按你的业务线来分，都可以。
- **功能开关 CRUD**：支持全量发布、百分比灰度、白名单定向。
- **用户定向**：基于 accountId + userId 的白名单规则，适配大多数业务场景。
- **分段发布**：通过 Segment + Condition 实现更灵活的受众筛选。
- **基于角色的访问控制**：System Admin → Tenant Admin → Tenant Viewer，三级权限。
- **异步评估日志**：所有评估事件先入 Redis，LogWorker 批量写入 PostgreSQL，不影响线上延迟。
- **RESTful API**：统一 `{code, message, data}` 响应格式，SDK 示例覆盖 Go、Java、JavaScript。
- **Swagger 文档**：启动后直接访问 `/swagger/index.html` 即可查看完整 API。

## 技术细节（给工程师看的部分）

项目分两个主要模块：

### facet-backend（Go）

采用经典的分层架构：Handler → Service → Repository。每一层都是接口注入，方便测试和替换。

```
cmd/server/main.go       # 入口，启动 HTTP + LogWorker
internal/
├── api/                 # HTTP 路由、中间件（JWT 认证）、处理器
├── cache/               # Redis 缓存封装，按租户级别缓存配置
├── domain/              # 领域模型和仓库接口
├── repository/postgres/ # PostgreSQL 数据访问、迁移脚本
├── server/              # 依赖注入（DI）组装
├── service/             # 业务逻辑层
└── worker/              # LogWorker：Redis → PostgreSQL 批量落盘
```

特别提一下认证：登录需要验证码，由 `base64Captcha` 生成图片，验证码校验结果缓存到 Redis。JWT token 有过期时间，可以在配置里调。

### facet-front（React）

前端用 React 19 + Material UI 7。状态管理用的是 Zustand，而不是 Redux——少写很多样板代码。

```
src/
├── api/         # 每个资源一个模块（auth、environment、featureFlags...）
├── stores/      # Zustand store（auth、env、theme、notification）
├── views/       # 页面组件
├── components/  # 通用组件（ErrorBoundary、ConfirmDialog）
├── router/      # 路由 + 守卫
└── utils/       # 工具函数
```

路由守卫的逻辑是：访问 `/manage/*` 路径时会检查 JWT token，没有就重定向到 `/login`。根据角色不同，某些按钮和操作会自动隐藏。

## 不开玩笑，上线前可以看看这几个文件

如果你打算在生产环境使用 Facet，建议先通读这几个文件：

- **`facet-backend/config.yaml`** —— 所有可配项都在这里，包括 JWT 密钥、Redis 地址、PostgreSQL DSN 等
- **`facet-backend/internal/repository/postgres/migrations/`** —— 数据库迁移脚本，建表语句一目了然
- **`facet-backend/docs/example/`** —— Go、Java、JavaScript 三套 SDK 示例，复制粘贴就能跑
- **`docker-compose.yml`** —— 一键启动所有服务

## 后续规划

目前 Facet 已经可以在生产环境跑起来了，但还有几个方向在规划中：

- **WebSocket 实时推送** —— 开关变更后主动通知 SDK，减少轮询
- **规则引擎增强** —— 更灵活的多条件组合、自定义属性匹配
- **OpenTelemetry 集成** —— 链路追踪，方便排查评估链路
- **Kubernetes 部署支持（Helm Chart）** —— 一键部署到 K8s 集群

## 一分钟快速体验

```bash
git clone https://github.com/anomalyco/facet.git
cd facet
docker compose up -d
```

浏览器打开 `http://localhost`，即可开始管理功能开关。首次使用需要通过 API 创建一个初始管理员账号（详见 README），后续管理操作都在 Web 界面完成。

## 写在最后

做这个项目的初衷很简单：需要一个功能开关平台，不想被供应商锁定，不想花不必要的钱，想要自己能掌控全部数据。

Facet 现在还是早期阶段，代码量不算大，架构清晰，适合二次开发和定制。如果你也有类似的需求，欢迎来看看，提 issue，甚至直接提 PR。

GitHub 仓库：[github.com/anomalyco/facet](https://github.com/anomalyco/facet)

---

*如果你觉得这个项目有用，点个 Star 就是最大的支持。*
