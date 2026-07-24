# Facet Backend API Documentation

Base URL: `/api/v1`

## Response Format

Success:
```json
{"code": 200, "message": "OK", "data": {...}}
```
Error:
```json
{"code": 404, "message": "resource not found"}
```

## SDK API (no auth)

### Evaluate Feature Flag

`GET /evaluate`

| Param | Type | Required | Description |
|-------|------|----------|-------------|
| `sdk_key` | string | yes | Environment SDK key |
| `feature_name` | string | yes | Feature flag name |
| `user_id` | string | no | User identifier |
| `account_id` | string | no | Account identifier |
| `app_name` | string | no | App name |
| `context` | string | no | JSON object for segment matching |

Response: `{"matched": true}`

## Auth API

### Captcha

`GET /manage/auth/captcha`

Response:
```json
{"captchaId": "...", "base64Image": "data:image/png;base64,..."}
```

### Login

`POST /manage/auth/login`

```json
{"email": "a@b.com", "password": "...", "captchaId": "...", "captchaAnswer": "..."}
```

Response:
```json
{"token": "jwt...", "user": {"id": "...", "displayName": "...", "role": "..."}}
```

## System Admin API (`/system`) — requires `system_admin` role

### Tenant Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/tenants` | Create tenant + super admin |
| PUT | `/tenants/:tenantId` | Update tenant |
| PUT | `/tenants/:tenantId/password` | Reset super admin password |
| GET | `/tenants` | List tenants (paginated) |
| GET | `/tenants/:tenantId` | Get tenant details |

### System Notifications

| Method | Path | Description |
|--------|------|-------------|
| POST | `/notifications` | Create notification |
| GET | `/notifications` | List notifications (paginated) |
| GET | `/notifications/:notificationId` | Get notification detail |
| PUT | `/notifications/:notificationId` | Update notification |
| DELETE | `/notifications/:notificationId` | Delete notification |

## Tenant Admin API (`/manage`) — requires auth

### User Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/users` | Create user |
| PUT | `/users/:userId` | Update user |
| DELETE | `/users/:userId` | Soft delete user |
| PUT | `/users/:userId/password` | Reset user password |
| GET | `/users/:userId` | Get user detail |
| GET | `/users` | List users (paginated) |

### Environment Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/environments` | Create environment |
| PUT | `/environments/:envId` | Update environment |
| DELETE | `/environments/:envId` | Soft delete environment |
| GET | `/environments` | List environments (paginated) |
| GET | `/environments/active-list` | List active environments (brief) |

### Feature Flag Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/environments/:envId/features` | Create flag |
| GET | `/environments/:envId/features` | List flags (paginated) |
| GET | `/environments/:envId/features/:featureId` | Get flag detail |
| PUT | `/environments/:envId/features/:featureId` | Update flag |
| DELETE | `/environments/:envId/features/:featureId` | Soft delete flag |
| GET | `/environments/:envId/features/:featureId/history` | Get flag history |
| POST | `/environments/:envId/features/:featureId/archive` | Archive flag |
| POST | `/environments/:envId/features/:featureId/unarchive` | Unarchive flag |

Query params for list: `page`, `pageSize`, `keyword`, `includeArchived`

### Segment Management

| Method | Path | Description |
|--------|------|-------------|
| POST | `/segments` | Create segment |
| GET | `/segments` | List segments |
| GET | `/segments/:segmentId` | Get segment detail |
| PUT | `/segments/:segmentId` | Update segment |
| DELETE | `/segments/:segmentId` | Delete segment |

Segment request body:
```json
{
  "name": "Premium Users",
  "description": "Users on premium plan",
  "matchType": "all",
  "conditions": [
    {"field": "plan", "operator": "equals", "value": "premium"}
  ]
}
```
Operators: `equals`, `not_equals`, `contains`, `not_contains`, `starts_with`, `ends_with`, `gt`, `gte`, `lt`, `lte`, `in`

### User Notifications

| Method | Path | Description |
|--------|------|-------------|
| GET | `/notifications` | List user notifications |
| GET | `/notifications/status` | Get unread count |
| POST | `/notifications/mark-as-read` | Mark notifications as read |

## Error Codes

| Code | Description |
|------|-------------|
| 400 | Bad request (missing params, invalid input) |
| 401 | Unauthorized (invalid credentials) |
| 403 | Forbidden (inactive resource, quota exceeded) |
| 404 | Not found |
| 409 | Conflict (duplicate name) |
| 500 | Internal server error |
