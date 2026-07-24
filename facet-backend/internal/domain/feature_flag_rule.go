package domain

import (
	"context"
	"database/sql"
	"time"
)

// FeatureFlagRule defines the core domain model for a feature flag allowlist rule
type FeatureFlagRule struct {
	ID            string         `db:"id"`
	FeatureFlagID string         `db:"feature_flag_id"`
	AccountID     sql.NullString `db:"account_id"`
	UserIDs       []string       `db:"user_ids"`
	CreatedBy     sql.NullString `db:"created_by"`
	UpdatedBy     sql.NullString `db:"updated_by"`
	CreatedAt     time.Time      `db:"created_at"`
	UpdatedAt     time.Time      `db:"updated_at"`
}

// FeatureFlagRuleRepository defines the interface for feature flag rule data operations
type FeatureFlagRuleRepository interface {
	Create(ctx context.Context, rule *FeatureFlagRule) error
	Update(ctx context.Context, rule *FeatureFlagRule) error
	ListByFeatureFlagID(ctx context.Context, featureFlagID string) ([]*FeatureFlagRule, error)
	DeleteByFeatureFlagID(ctx context.Context, featureFlagID string) error
	GetRulesByFlagID(ctx context.Context, featureFlagID string) ([]*FeatureFlagRule, error) // <-- New method
}
