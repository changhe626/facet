package domain

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"
)

// ChangeType defines the change types in history records
type ChangeType string

const (
	ChangeTypeCreated       ChangeType = "created"
	ChangeTypeUpdated       ChangeType = "updated"
	ChangeTypeRulesChanged  ChangeType = "rules_changed"
	ChangeTypeStatusToggled ChangeType = "status_toggled"
	ChangeTypeDeleted       ChangeType = "deleted"
)

// FeatureFlagHistory defines the complete snapshot of a feature flag change
type FeatureFlagHistory struct {
	HistoryID         string          `db:"history_id"`
	FeatureFlagID     string          `db:"feature_flag_id"`
	ChangeTime        time.Time       `db:"change_time"`
	ChangeType        ChangeType      `db:"change_type"`
	ChangedBy         string          `db:"changed_by"`
	Name              string          `db:"name"`
	Description       string          `db:"description"`
	Enabled           bool            `db:"enabled"`
	FlagType          string          `db:"flag_type"`
	RolloutPercentage int             `db:"rollout_percentage"`
	ExpiresAt         sql.NullString  `db:"expires_at"`
	RulesSnapshot     json.RawMessage `db:"rules_snapshot"`
}

// FeatureFlagHistoryRepository defines the data operation interface for feature flag history
type FeatureFlagHistoryRepository interface {
	// Create creates a new history record
	Create(ctx context.Context, history *FeatureFlagHistory) error

	// ListByFeatureFlagID lists all history versions of a feature flag in reverse chronological order
	ListByFeatureFlagID(ctx context.Context, flagID string) ([]*FeatureFlagHistory, error)
}
