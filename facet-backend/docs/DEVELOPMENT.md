# Local Development Guide

## Prerequisites

- Go 1.24+
- Node.js 22+
- PostgreSQL 17+
- Redis 7+

## Quick Start

### 1. Clone and configure

```bash
git clone <repo-url>
cd facet
cp .env.example .env
```

Edit `.env` with your local PostgreSQL and Redis credentials.

### 2. Start databases

```bash
# Using Docker (recommended)
docker compose up -d postgres redis

# Or start manually
pg_ctl start   # PostgreSQL
redis-server   # Redis
```

### 3. Backend

```bash
cd facet-backend
# edit config.yaml if needed
go run ./cmd/server/main.go
```

Server starts at `http://localhost:8080`.

Environment variables override `config.yaml`:
- `POSTGRES_DSN` — Postgres connection string
- `JWT_SECRET_KEY` — JWT signing key
- `REDIS_ADDRESS`, `REDIS_PASSWORD`, `REDIS_DB` — Redis config
- `LOGGER_LEVEL` — Log level (debug/info/warn/error)

### 4. Frontend

```bash
cd facet-front
npm install
npm run dev
```

Dev server at `http://localhost:5173`, proxies `/api` to `localhost:8080`.

### 5. Full stack with Docker

```bash
docker compose up --build
```

- Frontend: `http://localhost`
- Backend API: `http://localhost:8080`

## Testing

```bash
# Backend
cd facet-backend
go test ./...

# Frontend
cd facet-front
npm run test
```

## Database Migrations

Migrations are in `facet-backend/migrations/`. Run them manually against your Postgres:

```bash
psql -U postgres -d postgres -f migrations/001_init.sql
```

## Project Structure

```
facet/
├── facet-backend/          # Go backend
│   ├── cmd/server/         # Entrypoint
│   ├── internal/
│   │   ├── api/            # HTTP layer (handlers, middleware, router)
│   │   ├── cache/          # Redis cache service
│   │   ├── config/         # Configuration loading
│   │   ├── domain/         # Domain models and errors
│   │   ├── repository/     # Postgres data access
│   │   ├── server/         # DI wiring and lifecycle
│   │   ├── service/        # Business logic
│   │   └── worker/         # Background log worker
│   ├── migrations/         # SQL migrations
│   ├── pkg/                # Shared utilities (logger, password, token)
│   └── config.yaml         # Default configuration
├── facet-front/            # React frontend
│   ├── src/
│   │   ├── api/            # Axios API client
│   │   ├── components/     # Reusable components
│   │   ├── router/         # React Router config
│   │   ├── stores/         # Zustand state
│   │   ├── views/          # Page components
│   │   └── utils/          # Helpers
│   └── nginx.conf          # Production Nginx config
├── docker-compose.yml      # Full stack orchestration
└── .env.example            # Environment variable template
```
