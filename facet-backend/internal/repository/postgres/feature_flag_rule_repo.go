package postgres

import (
	"context"
	"facet-backend/internal/domain"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// featureFlagRuleRepository implements the domain.FeatureFlagRuleRepository interface
type featureFlagRuleRepository struct {
	pool *pgxpool.Pool
}

// NewFeatureFlagRuleRepository creates a new FeatureFlagRuleRepository instance
func NewFeatureFlagRuleRepository(pool *pgxpool.Pool) domain.FeatureFlagRuleRepository {
	return &featureFlagRuleRepository{pool: pool}
}

// Create inserts a new feature flag allowlist rule record into the database
func (r *featureFlagRuleRepository) Create(ctx context.Context, rule *domain.FeatureFlagRule) error {
	query := `
        INSERT INTO feature_flags_rules (id, feature_flag_id, account_id, user_ids, created_by, updated_by, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
        RETURNING id;
    `
	err := r.pool.QueryRow(ctx, query,
		rule.ID,
		rule.FeatureFlagID,
		rule.AccountID,
		rule.UserIDs,
		rule.CreatedBy,
		rule.UpdatedBy,
	).Scan(&rule.ID)

	if err != nil {
		return fmt.Errorf("failed to create feature flag rule: %w", err)
	}

	return nil
}

// Update updates an existing feature flag allowlist rule record
func (r *featureFlagRuleRepository) Update(ctx context.Context, rule *domain.FeatureFlagRule) error {
	query := `
        UPDATE feature_flags_rules
        SET account_id = $1, user_ids = $2, updated_by = $3, updated_at = NOW()
        WHERE id = $4;
    `
	cmdTag, err := r.pool.Exec(ctx, query,
		rule.AccountID,
		rule.UserIDs,
		rule.UpdatedBy,
		rule.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update feature flag rule: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		// In this "Upsert" scenario, not finding the record may not be a fatal error,
		// but we still return an error so the upper layer logic knows.
		// Alternatively, we could choose not to return an error here and let the upper layer logic continue.
		// For clarity, we return a generic not found error.
		return domain.ErrFeatureFlagNotFound // Reuse this error type
	}
	return nil
}

// ListByFeatureFlagID gets all rules by FeatureFlagID
func (r *featureFlagRuleRepository) ListByFeatureFlagID(ctx context.Context, featureFlagID string) ([]*domain.FeatureFlagRule, error) {
	query := `
        SELECT id, feature_flag_id, account_id, user_ids, created_at, updated_at, created_by, updated_by
        FROM feature_flags_rules
        WHERE feature_flag_id = $1;
    `
	rows, err := r.pool.Query(ctx, query, featureFlagID)
	if err != nil {
		return nil, fmt.Errorf("failed to query feature flag rules by feature flag id: %w", err)
	}
	defer rows.Close()

	var rules []*domain.FeatureFlagRule
	for rows.Next() {
		var rule domain.FeatureFlagRule
		if err := rows.Scan(
			&rule.ID,
			&rule.FeatureFlagID,
			&rule.AccountID,
			&rule.UserIDs,
			&rule.CreatedAt,
			&rule.UpdatedAt,
			&rule.CreatedBy,
			&rule.UpdatedBy,
		); err != nil {
			return nil, fmt.Errorf("failed to scan feature flag rule row: %w", err)
		}
		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feature flag rule rows: %w", err)
	}

	return rules, nil
}

// DeleteByFeatureFlagID deletes all rules by FeatureFlagID
func (r *featureFlagRuleRepository) DeleteByFeatureFlagID(ctx context.Context, featureFlagID string) error {
	query := `
        DELETE FROM feature_flags_rules
        WHERE feature_flag_id = $1;
    `
	_, err := r.pool.Exec(ctx, query, featureFlagID)
	if err != nil {
		return fmt.Errorf("failed to delete feature flag rules by feature flag id: %w", err)
	}
	return nil
}

// GetRulesByFlagID gets all rules for evaluation by FeatureFlagID (only queries necessary fields)
func (r *featureFlagRuleRepository) GetRulesByFlagID(ctx context.Context, featureFlagID string) ([]*domain.FeatureFlagRule, error) {
	query := `
        SELECT account_id, user_ids
        FROM feature_flags_rules
        WHERE feature_flag_id = $1;
    `
	rows, err := r.pool.Query(ctx, query, featureFlagID)
	if err != nil {
		return nil, fmt.Errorf("failed to query feature flag rules for evaluation: %w", err)
	}
	defer rows.Close()

	var rules []*domain.FeatureFlagRule
	for rows.Next() {
		var rule domain.FeatureFlagRule
		if err := rows.Scan(
			&rule.AccountID,
			&rule.UserIDs,
		); err != nil {
			return nil, fmt.Errorf("failed to scan feature flag rule row for evaluation: %w", err)
		}
		rules = append(rules, &rule)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feature flag rule rows for evaluation: %w", err)
	}

	return rules, nil
}
