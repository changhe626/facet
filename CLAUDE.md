# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Facet is a feature flag management platform (like LaunchDarkly). It evaluates boolean feature flags for SDK clients and provides a web admin panel for managing flags, environments, users, tenants, and notifications.

## Commands

### Backend (Go)

```sh
# facet-backend/
go run ./cmd/server/main.go        # Dev server on :8080
go test ./...                       # All tests
go test ./internal/api/middleware/  # Single package
go test -run TestName ./...         # Single test
```

### Frontend (React)

```sh
# facet-front/
npm run dev       # Vite dev server on :5173 (proxies /api/v1 to :8080)
npm run build     # Production build to dist/
npm run test      # Vitest
npm run preview   # Preview production build
```

### Full Stack

```sh
docker-compose up   # Starts postgres, redis, backend, frontend
```

## Architecture

### Monorepo Structure

```
facet/
├── facet-backend/     # Go 1.25, Gin, PostgreSQL (pgx/v5), Redis
├── facet-front/       # React 19, Vite, MUI v7, React Router v7, Zustand, Axios
├── docker-compose.yml # postgres + redis + backend(:8080) + frontend(:80)
└── logs/              # Shared application logs
```

### Backend: Layered Architecture (Handler → Service → Repository)

DI wiring in `internal/server/server.go`:

- **Handler** (`internal/api/handler/`): HTTP handlers, parse request/response. Auth, users, environments, feature flags, tenants, notifications, evaluations.
- **Middleware** (`internal/api/middleware/`): JWT auth, global error handler, context keys.
- **Service** (`internal/service/`): Business logic. Auth, captcha, environments, feature flags (with rules/history), evaluation, events, notifications, tenants, tenant levels, users.
- **Repository** (`internal/repository/postgres/`): PostgreSQL data access. One file per entity.
- **Cache** (`internal/cache/`): Redis-based cache for environment data and tenant-level configs.
- **Domain** (`internal/domain/`): Go structs for all entities (environment, feature_flag, user, tenant, etc.).
- **Worker** (`internal/worker/`): Background `LogWorker` goroutine that processes async evaluation events from Redis.
- **Config** (`internal/config/`): Viper-based config from `config.yaml` + env var overrides.
- **Observability** (`internal/observability/`): Metrics logging middleware.

### Backend Entrypoint

`facet-backend/cmd/server/main.go` → `server.NewServer()` in `internal/server/server.go` which:
1. Loads config
2. Connects PostgreSQL + Redis
3. Wires all dependencies (repos → services → handlers)
4. Registers routes via `api.NewRouter()`
5. Starts LogWorker goroutine + HTTP server

### Backend Routes

Defined in `internal/api/router.go`. Base: `/api/v1`.

| Route | Auth | Description |
|---|---|---|
| `GET /evaluate` | None (sdk_key param) | SDK flag evaluation |
| `POST /manage/auth/login` | None | Login |
| `GET /manage/auth/captcha` | None | Captcha image |
| `/manage/*` | JWT | Tenant-level management |
| `/system/tenants` | JWT + system_admin | Tenant CRUD |
| `/system/notifications` | JWT + system_admin | System notifications |

### Frontend: React 19 (Migrated from Vue 3)

The frontend was migrated from Vue 3 (Element Plus) to React 19 (MUI v7). Vue files have been deleted; the React codebase is being built out per `migration-plan-vue-to-react.md`.

Key frontend files:

- `src/utils/request.js` — Axios instance with baseURL `/api/v1`, Bearer token injection, and 401 redirect to `/login`.
- `src/utils/formatters.js` — `formatRole()`, `formatDateTime()`, `formatFlagType()`.
- `src/stores/` — Zustand stores:
  - `authStore.js` — `token`, `user`, `isAuthenticated`, `login()`, `logout()`
  - `envStore.js` — `environments`, `currentEnvId`, `fetchEnvironments()`
  - `themeStore.js` — `mode` (light/dark), `toggleTheme()`
  - `notificationStore.js` — `unreadCount`, `fetchUnreadCount()`

### API Response Convention

All API responses follow: `{ code: number, message: string, data: any }`.
Success codes: `200`, `201`. On `401`, the frontend clears token and redirects to `/login`.

### Infrastructure

- **Docker Compose**: PostgreSQL 17.6, Redis 7.2.5, backend (port 8080), frontend (port 80).
- **Nginx**: Two server blocks — official website (`qianmiandev.cn`) and admin panel (`admin.qianmiandev.cn`). API proxy pass to `backend:8080`.
- **Backend Dockerfile**: Multi-stage build with `golang:1.25.1-alpine` → `alpine:3.20.1`. Static binary. Non-root user.
- **Frontend Dockerfile**: Multi-stage build with `node:22.20.0-alpine` → `nginx:1.27.0-alpine`. Build output goes to `/usr/share/nginx/html/admin`.

### Backend Config & Env

Backend reads `config.yaml`, overridable by env vars:
`POSTGRES_DSN`, `JWT_SECRET_KEY`, `JWT_EXPIRES_IN_HOURS`, `LOGGER_LEVEL`, `REDIS_ADDRESS`, `REDIS_PASSWORD`, `REDIS_DB`

### Key Conventions

- `@` alias maps to `src/` in frontend imports
- Soft deletes used for users, environments, feature flags
- IDs generated via Snowflake
- Go tests use `testify/mock` with table-driven patterns
- All frontend UI text is in English
- Local storage keys: `token`, `user`, `theme`