package domain

import (
	"context"
	"errors"
	"time"
)

var (
	ErrTenantLevelNotFound = errors.New("tenant level not found")
)

// TenantLevel defines the core domain model for a tenant level (plan)
type TenantLevel struct {
	ID               string  `db:"id"`
	Code             string  `db:"code"`
	Name             string  `db:"name"`
	MaxEnvironments  int     `db:"max_environments"`
	MaxFeatureFlags  int     `db:"max_feature_flags"`
	MaxTenantAdmins  int     `db:"max_tenant_admins"`
	MaxTenantViewers int     `db:"max_tenant_viewers"`
	AllowSSO         bool    `db:"allow_sso"`
	AllowAPIAccess   bool    `db:"allow_api_access"`
	HasAuditLog      bool    `db:"has_audit_log"`
	PrioritySupport  bool    `db:"priority_support"`
	PriceMonthly     float64 `db:"price_monthly"`
	PriceYearly      float64 `db:"price_yearly"`
	IsPublic         bool      `db:"is_public"`
	IsActive         bool      `db:"is_active"`
	CreatedAt        time.Time `db:"created_at"`
	UpdatedAt        time.Time `db:"updated_at"`
}

// TenantLevelRepository defines the interface for tenant level data operations
type TenantLevelRepository interface {
	// GetByID gets plan details by level ID
	GetByID(ctx context.Context, id string) (*TenantLevel, error)

	// GetByCode gets plan details by the unique level code (e.g., 'free', 'pro')
	GetByCode(ctx context.Context, code string) (*TenantLevel, error)

	// ListPublic lists all public, active plans for the pricing page
	ListPublic(ctx context.Context) ([]*TenantLevel, error)

	// ListAll lists all plans for internal management
	ListAll(ctx context.Context) ([]*TenantLevel, error)
}
