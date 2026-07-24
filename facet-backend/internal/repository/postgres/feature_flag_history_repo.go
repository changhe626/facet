package postgres

import (
	"context"
	"facet-backend/internal/domain"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// featureFlagHistoryRepository implements the domain.FeatureFlagHistoryRepository interface
type featureFlagHistoryRepository struct {
	pool *pgxpool.Pool
}

// NewFeatureFlagHistoryRepository creates a new FeatureFlagHistoryRepository instance
func NewFeatureFlagHistoryRepository(pool *pgxpool.Pool) domain.FeatureFlagHistoryRepository {
	return &featureFlagHistoryRepository{pool: pool}
}

// Create inserts a new feature flag history record into the database (updated)
func (r *featureFlagHistoryRepository) Create(ctx context.Context, history *domain.FeatureFlagHistory) error {
	query := `
        INSERT INTO feature_flag_history (history_id, feature_flag_id, change_time, change_type, changed_by, name, description, enabled, flag_type, rollout_percentage, expires_at, rules_snapshot)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
    `
	_, err := r.pool.Exec(ctx, query,
		history.HistoryID,
		history.FeatureFlagID,
		history.ChangeTime,
		history.ChangeType,
		history.ChangedBy,
		history.Name,
		history.Description,
		history.Enabled,
		history.FlagType,
		history.RolloutPercentage,
		history.ExpiresAt,
		history.RulesSnapshot,
	)

	if err != nil {
		return fmt.Errorf("failed to create feature flag history: %w", err)
	}

	return nil
}

// ListByFeatureFlagID queries all history versions of a feature flag in reverse chronological order (updated)
func (r *featureFlagHistoryRepository) ListByFeatureFlagID(ctx context.Context, flagID string) ([]*domain.FeatureFlagHistory, error) {
	query := `
        SELECT history_id, feature_flag_id, change_time, change_type, changed_by, name, description, enabled, flag_type, rollout_percentage, expires_at, rules_snapshot
        FROM feature_flag_history
        WHERE feature_flag_id = $1
        ORDER BY change_time DESC;
    `
	rows, err := r.pool.Query(ctx, query, flagID)
	if err != nil {
		return nil, fmt.Errorf("failed to query feature flag history: %w", err)
	}
	defer rows.Close()

	var histories []*domain.FeatureFlagHistory
	for rows.Next() {
		var history domain.FeatureFlagHistory
		err := rows.Scan(
			&history.HistoryID,
			&history.FeatureFlagID,
			&history.ChangeTime,
			&history.ChangeType,
			&history.ChangedBy,
			&history.Name,
			&history.Description,
			&history.Enabled,
			&history.FlagType,
			&history.RolloutPercentage,
			&history.ExpiresAt,
			&history.RulesSnapshot,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feature flag history row: %w", err)
		}
		histories = append(histories, &history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feature flag history rows: %w", err)
	}

	return histories, nil
}
