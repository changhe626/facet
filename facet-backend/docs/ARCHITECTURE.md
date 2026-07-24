# Facet Architecture

## System Overview

```
┌──────────┐   ┌──────────┐   ┌──────────┐
│  Client  │   │  SDK App │   │  Browser │
│  SDKs    │   │  (Go/JS) │   │  (React) │
└────┬─────┘   └────┬─────┘   └────┬─────┘
     │              │              │
     ▼              ▼              ▼
┌──────────────────────────────────────────┐
│              HTTP / Nginx                 │
└────────────────────┬─────────────────────┘
                     │
┌────────────────────▼─────────────────────┐
│            Gin Router                     │
│  /health  /evaluate  /manage/*  /system/* │
└────────────────────┬─────────────────────┘
                     │
┌────────────────────▼─────────────────────┐
│           Handler Layer                   │
│  Auth │ Eval │ Feature │ Env │ User │ ... │
│  (request binding, validation, response)  │
└────────────────────┬─────────────────────┘
                     │
┌────────────────────▼─────────────────────┐
│           Service Layer                   │
│  Business logic, tenants, quotas,         │
│  flag evaluation, segments, archiving     │
│  ┌────────────────────────────────────┐   │
│  │  TenantLevelService (quota arbiter)│   │
│  └────────────────────────────────────┘   │
└────┬────────────┬────────────┬────────────┘
     │            │            │
┌────▼────┐ ┌────▼────┐ ┌────▼────┐
│ Postgres│ │  Redis  │ │  Cache  │
│  Repos  │ │  Events │ │ Service │
└─────────┘ └────┬────┘ └─────────┘
                  │
           ┌──────▼──────┐
           │  LogWorker  │
           │  BRPop → DB │
           └─────────────┘
```

## Layers

### Handler (`internal/api/handler/`)
HTTP handlers that bind requests, validate input, call services, and return responses.

### Service (`internal/service/`)
Business logic layer. Orchestrates repositories, enforces tenant quotas, evaluates feature flags (whitelist → percentage → segments), and manages cache invalidation.

### Repository (`internal/repository/postgres/`)
Data access layer using `pgx/v5`. Direct SQL queries with `search_path=facet` schema.

### Cache (`internal/cache/`)
Redis-based caching for environment configs used in SDK evaluation.

### Worker (`internal/worker/`)
Background goroutine that polls Redis for evaluation log events and batch-inserts them to Postgres via `CopyFrom`.

## Key Flows

### Flag Evaluation
```
SDK → GET /evaluate?sk_key&feature_name&user_id&account_id&context
  → EvaluationHandler
    → EvaluationService
      → EvaluationDataSource (fetches env/flags/rules/segments from cache or DB)
      → 1. Check enabled
      → 2. Check archived
      → 3. Check whitelist (accountId + userIds)
      → 4. Check percentage rollout (FNV32a hash % 100)
      → 5. Check user segments (condition operators)
      → 6. Return matched: true/false
  → EventService (async push to Redis)
  → LogWorker (async batch insert to Postgres)
```

### Flag CRUD
```
React UI → PUT /manage/environments/:envId/features/:id
  → FeatureFlagHandler
    → FeatureFlagService
      1. Validate quota (TenantLevelService)
      2. Validate name uniqueness
      3. Update feature_flag + rules + segment associations (transaction)
      4. Create history snapshot
      5. Invalidate environment cache in Redis
  → Response to UI
```
