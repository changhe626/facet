package domain

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrFeatureFlagNotFound           = errors.New("feature flag not found")
	ErrFeatureFlagNameExistsInTenant = errors.New("feature flag name already exists in this tenant")
	ErrFeatureFlagDisabled           = errors.New("feature flag is disabled")
	ErrFeatureFlagArchived           = errors.New("feature flag is archived")
)

// FlagType defines the type of feature flag
type FlagType string

const (
	FlagTypeRelease     FlagType = "RELEASE"     // Gradual rollout flag
	FlagTypeExperiment  FlagType = "EXPERIMENT"  // A/B test flag
	FlagTypeOperational FlagType = "OPERATIONAL" // Operational flag
	FlagTypePermission  FlagType = "PERMISSION"  // Permission flag
)

// FeatureFlag defines the core domain model for a feature flag
type FeatureFlag struct {
	ID                string         `db:"id"`
	TenantID          string         `db:"tenant_id"`
	EnvironmentID     string         `db:"environment_id"`
	Name              string         `db:"name"`
	Description       sql.NullString `db:"description"`
	Enabled           bool           `db:"enabled"`
	FlagType          FlagType       `db:"flag_type"`
	RolloutPercentage int            `db:"rollout_percentage"`
	ExpiresAt         sql.NullString `db:"expires_at"`
	ArchivedAt        sql.NullTime   `db:"archived_at"`
	CreatedBy         sql.NullString `db:"created_by"`
	UpdatedBy         sql.NullString `db:"updated_by"`
	CreatedAt         time.Time      `db:"created_at"`
	UpdatedAt         time.Time      `db:"updated_at"`
	DeletedAt         sql.NullTime   `db:"deleted_at"`
}

// ListFeatureFlagsParams defines the parameters for querying feature flags
type ListFeatureFlagsParams struct {
	EnvironmentID string
	Keyword       string
	IncludeArchived bool
	Limit         int
	Offset        int
}

// FeatureFlagRepository defines the interface for feature flag data operations
type FeatureFlagRepository interface {
	ListAndCountByEnvironment(ctx context.Context, params ListFeatureFlagsParams) ([]*FeatureFlag, int64, error)
	Create(ctx context.Context, flag *FeatureFlag) error
	GetByID(ctx context.Context, id string, envID string) (*FeatureFlag, error)
	Update(ctx context.Context, flag *FeatureFlag) error
	Delete(ctx context.Context, id string, envID string) error
	CheckNameExistsInTenant(ctx context.Context, name, tenantID string) (bool, error)
	FindByNameAndEnvironmentID(ctx context.Context, name string, envID string, tenantID string) (*FeatureFlag, error)
	FindByEnvironmentID(ctx context.Context, envID string) ([]*FeatureFlag, error)
	CountByTenant(ctx context.Context, tenantID string) (int64, error)
	Archive(ctx context.Context, id string) error
	Unarchive(ctx context.Context, id string) error
}
