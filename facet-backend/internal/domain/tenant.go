package domain

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrTenantNameExists = errors.New("tenant name already exists")
	ErrTenantNotFound   = errors.New("tenant not found")
	ErrTenantInactive   = errors.New("tenant is not active")
)

type TenantStatus string

const (
	TenantStatusActive    TenantStatus = "active"
	TenantStatusSuspended TenantStatus = "suspended"
	TenantStatusArchived  TenantStatus = "archived"
)

// Tenant defines the core domain model for a tenant
type Tenant struct {
	ID          string         `db:"id"`
	LevelID     string         `db:"level_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	Website     sql.NullString `db:"website"`
	Status      TenantStatus   `db:"status"`
	CreatedBy   sql.NullString `db:"created_by"` // <-- Fixed
	UpdatedBy   sql.NullString `db:"updated_by"` // <-- Fixed
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

// ListTenantsParams defines the parameters for querying tenants
type ListTenantsParams struct {
	Name   string
	Limit  int
	Offset int
}

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *Tenant) error
	GetByID(ctx context.Context, id string) (*Tenant, error)
	Update(ctx context.Context, tenant *Tenant) error
	CheckNameExistsExcludingID(ctx context.Context, name, idToExclude string) (bool, error)
	ListAndCount(ctx context.Context, params ListTenantsParams) ([]*Tenant, int64, error)
}
