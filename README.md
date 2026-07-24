# Facet — Feature Flag Management Platform

[![License](https://img.shields.io/badge/license-AGPLv3-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Node](https://img.shields.io/badge/Node-20+-339933?logo=node.js)](https://nodejs.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

Facet is an open-source feature flag (feature toggle) management platform — a self-hosted alternative to LaunchDarkly. Built with Go and React, it provides a centralized dashboard to manage feature flags across environments, tenants, and users, enabling safe, gradual rollouts and targeted feature releases without vendor lock-in.

## Why Facet?

- **Self-hosted & Open Source** — Full control over your data. No third-party API calls, no per-seat pricing.
- **Multi-Tenant by Design** — Built for platforms that need to isolate flags per team or client out of the box.
- **Async Evaluation Logging** — Redis-backed pipeline that never blocks the flag evaluation hot path.
- **SDK-Agnostic** — Simple REST API; use any HTTP client. Official examples in Go, Java, and JavaScript.

## Screenshots

### Tenant Management

![Tenant Management](pictures/%E7%A7%9F%E6%88%B7%E7%AE%A1%E7%90%86.png "Tenant Management")
![Create Tenant](pictures/%E5%88%9B%E5%BB%BA%E7%A7%9F%E6%88%B7.png "Create Tenant")

### User Management

![User List](pictures/%E7%94%A8%E6%88%B7%E5%88%97%E8%A1%A8.png "User List")
![Create User](pictures/%E5%88%9B%E5%BB%BA%E7%94%A8%E6%88%B7.png "Create User")

### Environment Management

![Environment List](pictures/%E7%8E%AF%E5%A2%83%E5%88%97%E8%A1%A8.png "Environment List")
![Create Environment](pictures/%E5%88%9B%E5%BB%BA%E4%B8%80%E4%B8%AA%E7%8E%AF%E5%A2%83.png "Create Environment")

### Feature Flags

![Flag List](pictures/%E5%BC%80%E5%85%B3%E5%88%97%E8%A1%A8.png "Flag List")
![Create Flag](pictures/%E5%88%9B%E5%BB%BA%E4%B8%80%E4%B8%AA%E5%BC%80%E5%85%B3.png "Create Flag")

### System Notifications

![Notification List](pictures/%E7%B3%BB%E7%BB%9F%E9%80%9A%E7%9F%A5%E5%88%97%E8%A1%A8.png "Notification List")
![Create Notification](pictures/%E5%88%9B%E5%BB%BA%E4%B8%80%E4%B8%AA%E7%B3%BB%E7%BB%9F%E9%80%9A%E7%9F%A5.png "Create Notification")

## Features

- **Multi-Tenant Architecture** — Isolated environments for different teams or clients
- **Environment Management** — Dev, staging, production, or any custom environment
- **Feature Flag CRUD** — Create, update, delete, and toggle flags with targeting rules
- **User Targeting** — Whitelist-based rules by account and user ID
- **Role-Based Access Control** — System admin, tenant admin, tenant viewer roles
- **Captcha & JWT Auth** — Secure login with captcha verification
- **System Notifications** — Broadcast announcements to all tenants
- **Evaluation Logs** — Async log processing via Redis → PostgreSQL pipeline
- **i18n (中文/English)** — Built-in language switch between Chinese and English
- **RESTful API** — JSON `{code, message, data}` response convention
- **SDK Examples** — [Go, Java, JavaScript](facet-backend/docs/) (Node.js & Browser)

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go, Gin, pgx/v5, Redis (go-redis/v9), Snowflake IDs |
| Frontend | React 19, MUI v7, React Router v7, Zustand, Axios |
| Database | PostgreSQL 17 |
| Cache | Redis 7 |
| Deployment | Docker Compose |

## Quick Start

```bash
# Clone the repository
git clone https://github.com/anomalyco/facet.git
cd facet

# Start all services (PostgreSQL, Redis, backend, frontend)
docker compose up -d

# Access the dashboard
# Frontend: http://localhost
# Backend API: http://localhost:8080
```

**Default login:** After startup, the system has no pre-seeded user accounts. You need to connect to the database and insert an initial admin user, or use the API to create one. See the backend development docs for instructions.

## Project Structure

```
facet/
├── facet-backend/        # Go backend (cmd/server/internal)
│   ├── cmd/server/       # Entrypoint
│   ├── internal/
│   │   ├── api/          # HTTP handlers, middleware, routing
│   │   ├── cache/        # Redis cache layer
│   │   ├── domain/       # Domain types & interfaces
│   │   ├── repository/   # Data access layer (PostgreSQL)
│   │   ├── server/       # DI wiring & lifecycle
│   │   ├── service/      # Business logic
│   │   └── worker/       # Background workers
│   ├── internal/repository/postgres/migrations/  # Database migrations
│   └── docs/             # API docs & SDK examples
├── facet-front/          # React frontend
│   ├── website/          # Marketing landing page
│   └── src/
│       ├── api/          # API client modules
│       ├── layouts/      # Layout components
│       ├── router/       # Route definitions
│       ├── stores/       # Zustand state stores
│       ├── utils/        # Utilities & formatters
│       └── views/        # Page components
├── docker-compose.yml    # Docker Compose configuration
└── LICENSE
```

## Configuration

Key environment variables (see `.env.example` for full list):

| Variable | Description |
|----------|-------------|
| `POSTGRES_DSN` | PostgreSQL connection string |
| `JWT_SECRET_KEY` | JWT signing secret |
| `JWT_EXPIRES_IN_HOURS` | Token expiration (default: 24) |
| `REDIS_ADDRESS` | Redis server address |
| `REDIS_PASSWORD` | Redis password |
| `LOGGER_LEVEL` | Log level (debug, info, warn, error) |

## API Documentation

See [API_DOC.md](facet-backend/API_DOC.md) for the complete API reference. A Swagger UI is also available when the server is running — visit `/swagger/index.html` on the backend port.

Base URL: `/api/v1`

- `GET /evaluate` — SDK flag evaluation (no auth)
- `/manage/auth/*` — Login & captcha
- `/manage/*` — Tenant management (JWT auth)
- `/system/*` — System admin (system_admin role required)

## SDK Examples

Ready-to-run SDK client examples are available in [`facet-backend/docs/example/`](facet-backend/docs/example/):

| Language | File | Minimum Version |
|----------|------|-----------------|
| Go | [`after_1_18.go`](facet-backend/docs/example/go/after_1_18.go) | 1.18+ |
| Java | [`Java17Demo.java`](facet-backend/docs/example/java/Java17Demo.java) | 17+ |
| JavaScript (Node.js) | [`after_node_18.js`](facet-backend/docs/example/JavaScript/after_node_18.js) | 18+ |
| JavaScript (Browser) | [`web.js`](facet-backend/docs/example/JavaScript/web.js) | Modern browsers |

## Development

### Prerequisites

- Go 1.25+
- Node.js 20+
- PostgreSQL 17
- Redis 7

### Local Setup

```bash
# Start dependencies
docker compose up -d postgres redis

# Backend
cd facet-backend
cp .env.example .env
go run ./cmd/server/main.go

# Frontend (separate terminal)
cd facet-front
npm install
npm run dev
```

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](CONTRIBUTING.md) first.

## License

[AGPL v3](LICENSE)
