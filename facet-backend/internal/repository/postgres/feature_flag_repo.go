package postgres

import (
	"context"
	"errors"
	"facet-backend/internal/domain"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// featureFlagRepository implements the domain.FeatureFlagRepository interface
type featureFlagRepository struct {
	pool *pgxpool.Pool
}

// NewFeatureFlagRepository creates a new FeatureFlagRepository instance
func NewFeatureFlagRepository(pool *pgxpool.Pool) domain.FeatureFlagRepository {
	return &featureFlagRepository{pool: pool}
}

// ListAndCountByEnvironment lists and searches feature flags by environment with pagination
func (r *featureFlagRepository) ListAndCountByEnvironment(ctx context.Context, params domain.ListFeatureFlagsParams) ([]*domain.FeatureFlag, int64, error) {
	var total int64
	var whereClauses []string
	var args []interface{}
	argCount := 1

	// Always filter by environment_id
	whereClauses = append(whereClauses, fmt.Sprintf("environment_id = $%d", argCount))
	args = append(args, params.EnvironmentID)
	argCount++

	// If keyword is provided, add fuzzy search condition on name
	if params.Keyword != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argCount))
		args = append(args, "%"+params.Keyword+"%")
		argCount++
	}

	// Always exclude soft-deleted records
	whereClauses = append(whereClauses, "deleted_at IS NULL")

	// Exclude archived records unless explicitly requested
	if !params.IncludeArchived {
		whereClauses = append(whereClauses, "archived_at IS NULL")
	}

	whereStr := "WHERE " + strings.Join(whereClauses, " AND ")

	// --- Execute COUNT query ---
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM feature_flags %s", whereStr)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count feature flags: %w", err)
	}

	if total == 0 {
		return []*domain.FeatureFlag{}, 0, nil
	}

	// --- Execute SELECT query ---
	listQuery := fmt.Sprintf(`
        SELECT id, name, enabled, flag_type, rollout_percentage, expires_at, archived_at, updated_at, updated_by
        FROM feature_flags
        %s
        ORDER BY updated_at DESC
        LIMIT $%d OFFSET $%d
    `, whereStr, argCount, argCount+1)

	listArgs := append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list feature flags: %w", err)
	}
	defer rows.Close()

	var flags []*domain.FeatureFlag
	for rows.Next() {
		var flag domain.FeatureFlag
		if err := rows.Scan(
			&flag.ID,
			&flag.Name,
			&flag.Enabled,
			&flag.FlagType,
			&flag.RolloutPercentage,
			&flag.ExpiresAt,
			&flag.ArchivedAt,
			&flag.UpdatedAt,
			&flag.UpdatedBy,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan feature flag row for list: %w", err)
		}
		flags = append(flags, &flag)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating feature flag rows: %w", err)
	}

	return flags, total, nil
}

// Create creates a new feature flag
func (r *featureFlagRepository) Create(ctx context.Context, flag *domain.FeatureFlag) error {
	query := `
        INSERT INTO feature_flags (id, tenant_id, environment_id, name, description, enabled, flag_type, rollout_percentage, expires_at, created_by, updated_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
        RETURNING created_at, updated_at;
    `
	err := r.pool.QueryRow(ctx, query,
		flag.ID,
		flag.TenantID,
		flag.EnvironmentID,
		flag.Name,
		flag.Description,
		flag.Enabled,
		flag.FlagType,
		flag.RolloutPercentage,
		flag.ExpiresAt,
		flag.CreatedBy,
		flag.UpdatedBy,
	).Scan(&flag.CreatedAt, &flag.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create feature flag: %w", err)
	}
	return nil
}

// GetByID gets the details of a single feature flag
func (r *featureFlagRepository) GetByID(ctx context.Context, id string, envID string) (*domain.FeatureFlag, error) {
	query := `
        SELECT id, tenant_id, environment_id, name, description, enabled, flag_type, rollout_percentage, expires_at, archived_at, created_by, updated_by, created_at, updated_at
        FROM feature_flags
        WHERE id = $1 AND environment_id = $2 AND deleted_at IS NULL;
    `
	row := r.pool.QueryRow(ctx, query, id, envID)
	var flag domain.FeatureFlag
	err := row.Scan(
		&flag.ID, &flag.TenantID, &flag.EnvironmentID, &flag.Name, &flag.Description,
		&flag.Enabled, &flag.FlagType, &flag.RolloutPercentage, &flag.ExpiresAt, &flag.ArchivedAt, &flag.CreatedBy, &flag.UpdatedBy,
		&flag.CreatedAt, &flag.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFeatureFlagNotFound
		}
		return nil, fmt.Errorf("failed to scan feature flag detail row: %w", err)
	}
	return &flag, nil
}

// Update updates a feature flag
func (r *featureFlagRepository) Update(ctx context.Context, flag *domain.FeatureFlag) error {
	query := `
        UPDATE feature_flags
        SET description = $1, enabled = $2, flag_type = $3, rollout_percentage = $4, expires_at = $5, updated_by = $6, updated_at = NOW()
        WHERE id = $7;
    `
	cmdTag, err := r.pool.Exec(ctx, query,
		flag.Description, flag.Enabled, flag.FlagType, flag.RolloutPercentage, flag.ExpiresAt, flag.UpdatedBy, flag.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update feature flag: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrFeatureFlagNotFound
	}
	return nil
}

// Delete soft-deletes a feature flag
func (r *featureFlagRepository) Delete(ctx context.Context, id string, envID string) error {
	query := `UPDATE feature_flags SET deleted_at = NOW() WHERE id = $1 AND environment_id = $2`
	cmdTag, err := r.pool.Exec(ctx, query, id, envID)
	if err != nil {
		return fmt.Errorf("failed to soft delete feature flag: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrFeatureFlagNotFound
	}
	return nil
}

// FindByNameAndEnvironmentID ... (this method has been fixed)
func (r *featureFlagRepository) FindByNameAndEnvironmentID(ctx context.Context, name string, envID string, tenantID string) (*domain.FeatureFlag, error) {
	query := `
        SELECT id, tenant_id, environment_id, name, description, enabled, flag_type, rollout_percentage, expires_at, archived_at, created_by, updated_by, created_at, updated_at
        FROM feature_flags
        WHERE tenant_id = $3 AND name = $1 AND environment_id = $2 AND deleted_at IS NULL;
    `
	rows := r.pool.QueryRow(ctx, query, name, envID, tenantID)

	var flag domain.FeatureFlag
	err := rows.Scan(
		&flag.ID, &flag.TenantID, &flag.EnvironmentID, &flag.Name, &flag.Description,
		&flag.Enabled, &flag.FlagType, &flag.RolloutPercentage, &flag.ExpiresAt, &flag.ArchivedAt, &flag.CreatedBy, &flag.UpdatedBy,
		&flag.CreatedAt, &flag.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrFeatureFlagNotFound
		}
		return nil, fmt.Errorf("failed to scan feature flag row by name and env: %w", err)
	}

	return &flag, nil
}

// CheckNameExistsInTenant checks if a feature flag with the same name exists in the tenant
func (r *featureFlagRepository) CheckNameExistsInTenant(ctx context.Context, name, tenantID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM feature_flags WHERE name = $1 AND tenant_id = $2 AND deleted_at IS NULL)`
	var exists bool
	err := r.pool.QueryRow(ctx, query, name, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if feature flag name exists in tenant: %w", err)
	}
	return exists, nil
}

// FindByEnvironmentID gets all feature flags in an environment
func (r *featureFlagRepository) FindByEnvironmentID(ctx context.Context, envID string) ([]*domain.FeatureFlag, error) {
	query := `
        SELECT id, tenant_id, environment_id, name, description, enabled, flag_type, rollout_percentage, expires_at, archived_at, created_by, updated_by, created_at, updated_at
        FROM feature_flags
        WHERE environment_id = $1 AND deleted_at IS NULL ORDER BY created_at DESC;
    `
	rows, err := r.pool.Query(ctx, query, envID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var flags []*domain.FeatureFlag
	for rows.Next() {
		var flag domain.FeatureFlag
		err := rows.Scan(
			&flag.ID, &flag.TenantID, &flag.EnvironmentID, &flag.Name, &flag.Description,
			&flag.Enabled, &flag.FlagType, &flag.RolloutPercentage, &flag.ExpiresAt, &flag.ArchivedAt, &flag.CreatedBy, &flag.UpdatedBy,
			&flag.CreatedAt, &flag.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan feature flag row for FindByEnvironmentID: %w", err)
		}
		flags = append(flags, &flag)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating feature flag rows for FindByEnvironmentID: %w", err)
	}

	return flags, nil
}

// CountByTenant counts the total number of feature flags in the specified tenant (excluding soft-deleted)
func (r *featureFlagRepository) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	query := `SELECT COUNT(*) FROM feature_flags WHERE tenant_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count feature flags by tenant: %w", err)
	}
	return count, nil
}

// Archive sets archived_at for a feature flag
func (r *featureFlagRepository) Archive(ctx context.Context, id string) error {
	query := `UPDATE feature_flags SET archived_at = NOW(), updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to archive feature flag: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrFeatureFlagNotFound
	}
	return nil
}

// Unarchive clears archived_at for a feature flag
func (r *featureFlagRepository) Unarchive(ctx context.Context, id string) error {
	query := `UPDATE feature_flags SET archived_at = NULL, updated_at = NOW() WHERE id = $1 AND deleted_at IS NULL`
	cmdTag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to unarchive feature flag: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrFeatureFlagNotFound
	}
	return nil
}
