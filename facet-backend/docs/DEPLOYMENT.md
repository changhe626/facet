# Docker Deployment

## Quick Deploy

```bash
# 1. Clone
git clone <repo-url>
cd facet

# 2. Configure
cp .env.example .env
# Edit .env with your secrets (POSTGRES_PASSWORD, JWT_SECRET_KEY, etc.)

# 3. Start all services
docker compose up --build -d
```

## Services

| Service | Port | Description |
|---------|------|-------------|
| Frontend (Nginx) | 80 | React SPA |
| Backend (Gin) | 8080 | REST API |
| PostgreSQL | 5432 | Primary database |
| Redis | 6379 | Cache + event queue |

## Environment Variables

See `.env.example` for all available variables. Key ones:

| Variable | Default | Description |
|----------|---------|-------------|
| `POSTGRES_PASSWORD` | `facet_secret` | Postgres password |
| `REDIS_PASSWORD` | `facet_secret` | Redis password |
| `JWT_SECRET_KEY` | (set in docker-compose) | JWT signing key |
| `JWT_EXPIRES_IN_HOURS` | `24` | Token expiry |
| `LOGGER_LEVEL` | `debug` | Log level |

## Health Checks

Each service has a Docker healthcheck:
- **Postgres**: `pg_isready`
- **Redis**: `redis-cli ping`
- **Backend**: `GET /health` — returns `{"status": "ok"}`

Services start in dependency order (backend waits for healthy postgres + redis).

## Data Persistence

- Postgres data: `postgres_data` volume
- Redis data: `redis-data` volume
- Backend logs: stdout (Docker logs); file logging disabled in container

## Scaling

The backend is stateless (sessions stored in JWT, cache in Redis). Scale with:

```bash
docker compose up -d --scale backend=3
```

Add a reverse proxy (Nginx, Traefik) for load balancing.
