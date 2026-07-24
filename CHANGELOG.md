# Changelog

## v0.1.0 (2026-07-23)

Initial open-source release of Facet — a Feature Flag management platform.

### Features

- **Feature Flags**: Create, update, delete, archive/unarchive feature flags with whitelist rules
- **Percentage Rollout**: Gradual rollout from 0% to 100% using FNV32a hash
- **User Segments**: Attribute-based targeting with 11 condition operators (equals, contains, in, gt, etc.)
- **Evaluation API**: SDK endpoint with whitelist → percentage → segment matching pipeline
- **Tenant Management**: Multi-tenant with role-based access (system_admin, tenant_super_admin, tenant_admin, tenant_viewer)
- **Environment Management**: Isolated environments with SDK keys per tenant
- **User Management**: CRUD with password reset and role assignment
- **Authentication**: JWT-based login with captcha verification
- **Notifications**: System notifications + per-user notification center
- **Audit History**: Full change history for every feature flag

### Technical

- **Backend**: Go 1.24, Gin, pgx/v5, Redis, layered architecture (Handler → Service → Repository)
- **Frontend**: React 19, MUI v7, Vite 8, Zustand, React Router
- **Infrastructure**: Docker Compose (PostgreSQL 17 + Redis 7 + Nginx)
- **Testing**: testify (31 handler tests, 32 service tests), vitest
- **Docs**: API documentation, architecture guide, local dev guide, deployment guide
