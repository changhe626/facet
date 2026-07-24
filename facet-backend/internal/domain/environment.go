package domain

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrEnvironmentNotFound      = errors.New("environment not found")
	ErrEnvironmentNameExists    = errors.New("environment name already exists in this tenant")
	ErrEnvironmentInactive      = errors.New("environment is not active")
	ErrEnvironmentInvalidSDKKey = errors.New("invalid sdk key")
)

// Environment defines the core domain model for an environment
type Environment struct {
	ID          string         `db:"id"`
	TenantID    string         `db:"tenant_id"`
	Name        string         `db:"name"`
	Description sql.NullString `db:"description"`
	SDKKey      string         `db:"sdk_key"`
	IsActive    bool           `db:"is_active"`
	CreatedBy   sql.NullString `db:"created_by"`
	UpdatedBy   sql.NullString `db:"updated_by"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedAt   time.Time      `db:"updated_at"`
	DeletedAt   sql.NullTime   `db:"deleted_at"`
}

// ListEnvironmentsParams defines the parameters for querying environments
type ListEnvironmentsParams struct {
	TenantID string
	Keyword  string // Fuzzy search by name
	IsActive *bool  // Filter by active status
	Limit    int
	Offset   int
}

// EnvironmentRepository defines the interface for environment data operations
type EnvironmentRepository interface {
	Create(ctx context.Context, env *Environment) error
	Update(ctx context.Context, env *Environment) error
	Delete(ctx context.Context, id string, tenantID string) error
	ListByTenant(ctx context.Context, tenantID string) ([]*Environment, error)
	GetByID(ctx context.Context, id string, tenantID string) (*Environment, error)
	CheckNameExistsExcludingID(ctx context.Context, name, tenantID, idToExclude string) (bool, error)
	FindBySDKKey(ctx context.Context, sdkKey string) (*Environment, error)
	ListAndCount(ctx context.Context, params ListEnvironmentsParams) ([]*Environment, int64, error)
	CountByTenant(ctx context.Context, tenantID string) (int64, error)
}
