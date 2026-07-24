package postgres

import (
	"context"
	"errors"
	"facet-backend/internal/domain"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// environmentRepository implements the domain.EnvironmentRepository interface
type environmentRepository struct {
	pool *pgxpool.Pool
}

// NewEnvironmentRepository creates a new EnvironmentRepository instance
func NewEnvironmentRepository(pool *pgxpool.Pool) domain.EnvironmentRepository {
	return &environmentRepository{pool: pool}
}

// Create inserts a new environment record into the database
func (r *environmentRepository) Create(ctx context.Context, env *domain.Environment) error {
	query := `
        INSERT INTO environments (id, tenant_id, name, description, sdk_key, is_active, created_by, updated_by, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
        RETURNING created_at, updated_at;
    `
	err := r.pool.QueryRow(ctx, query,
		env.ID,
		env.TenantID,
		env.Name,
		env.Description,
		env.SDKKey,
		env.IsActive,
		env.CreatedBy,
		env.UpdatedBy,
	).Scan(&env.CreatedAt, &env.UpdatedAt)

	if err != nil {
		return fmt.Errorf("failed to create environment: %w", err)
	}

	return nil
}

// Update updates an existing environment record
func (r *environmentRepository) Update(ctx context.Context, env *domain.Environment) error {
	query := `
        UPDATE environments
        SET name = $1, is_active = $2, description = $3, updated_by = $4, updated_at = NOW()
        WHERE id = $5 AND tenant_id = $6 AND deleted_at IS NULL
        RETURNING updated_at;
    `
	err := r.pool.QueryRow(ctx, query,
		env.Name,
		env.IsActive,
		env.Description,
		env.UpdatedBy,
		env.ID,
		env.TenantID,
	).Scan(&env.UpdatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrEnvironmentNotFound
		}
		return fmt.Errorf("failed to update environment: %w", err)
	}

	return nil
}

// Delete soft-deletes an environment
func (r *environmentRepository) Delete(ctx context.Context, id string, tenantID string) error {
	query := `
        UPDATE environments
        SET deleted_at = $1
        WHERE id = $2 AND tenant_id = $3;
    `
	cmdTag, err := r.pool.Exec(ctx, query, time.Now(), id, tenantID)
	if err != nil {
		return fmt.Errorf("failed to soft delete environment: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrEnvironmentNotFound
	}

	return nil
}

// ListByTenant lists all environments by tenant ID (excluding soft-deleted)
func (r *environmentRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Environment, error) {
	query := `
        SELECT id, name, sdk_key, is_active
        FROM environments
        WHERE tenant_id = $1 AND deleted_at IS NULL
        ORDER BY created_at DESC;
    `
	rows, err := r.pool.Query(ctx, query, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to query environments: %w", err)
	}
	defer rows.Close()

	var environments []*domain.Environment
	for rows.Next() {
		var env domain.Environment
		if err := rows.Scan(
			&env.ID,
			&env.Name,
			&env.SDKKey,
			&env.IsActive,
		); err != nil {
			return nil, fmt.Errorf("failed to scan environment row: %w", err)
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environment rows: %w", err)
	}

	return environments, nil
}

// GetByID gets a single environment by ID and tenant ID (excluding soft-deleted)
func (r *environmentRepository) GetByID(ctx context.Context, id string, tenantID string) (*domain.Environment, error) {
	query := `
        SELECT id, tenant_id, name, description, sdk_key, is_active, created_at, updated_at, created_by, updated_by
        FROM environments
        WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;
    `
	row := r.pool.QueryRow(ctx, query, id, tenantID)

	var env domain.Environment
	err := row.Scan(
		&env.ID,
		&env.TenantID,
		&env.Name,
		&env.Description,
		&env.SDKKey,
		&env.IsActive,
		&env.CreatedAt,
		&env.UpdatedAt,
		&env.CreatedBy,
		&env.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEnvironmentNotFound
		}
		return nil, fmt.Errorf("failed to scan environment row for get: %w", err)
	}

	return &env, nil
}

// CheckNameExistsExcludingID checks if an environment name already exists in the specified tenant (optionally excluding an ID)
func (r *environmentRepository) CheckNameExistsExcludingID(ctx context.Context, name, tenantID, idToExclude string) (bool, error) {
	var whereClauses []string
	var args []interface{}
	argCount := 1

	whereClauses = append(whereClauses, fmt.Sprintf("name = $%d", argCount))
	args = append(args, name)
	argCount++

	whereClauses = append(whereClauses, fmt.Sprintf("tenant_id = $%d", argCount))
	args = append(args, tenantID)
	argCount++

	if idToExclude != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("id != $%d", argCount))
		args = append(args, idToExclude)
		argCount++
	}

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	query := fmt.Sprintf(`
        SELECT EXISTS (
            SELECT 1
            FROM environments
            WHERE %s
        );
    `, strings.Join(whereClauses, " AND "))

	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if environment name exists: %w", err)
	}

	return exists, nil
}

// FindBySDKKey gets a single environment by SDK Key (excluding soft-deleted)
func (r *environmentRepository) FindBySDKKey(ctx context.Context, sdkKey string) (*domain.Environment, error) {
	query := `
        SELECT id, tenant_id, name, description, sdk_key, is_active, created_at, updated_at, created_by, updated_by
        FROM environments
        WHERE sdk_key = $1 AND deleted_at IS NULL;
    `
	row := r.pool.QueryRow(ctx, query, sdkKey)

	var env domain.Environment
	err := row.Scan(
		&env.ID,
		&env.TenantID,
		&env.Name,
		&env.Description,
		&env.SDKKey,
		&env.IsActive,
		&env.CreatedAt,
		&env.UpdatedAt,
		&env.CreatedBy,
		&env.UpdatedBy,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrEnvironmentInvalidSDKKey
		}
		return nil, fmt.Errorf("failed to scan environment row by sdk key: %w", err)
	}

	return &env, nil
}

// ListAndCount lists all environments by tenant ID with pagination (excluding soft-deleted)
func (r *environmentRepository) ListAndCount(ctx context.Context, params domain.ListEnvironmentsParams) ([]*domain.Environment, int64, error) {
	var total int64
	var whereClauses []string
	var args []interface{}
	argCount := 1

	// Always filter by tenant ID
	whereClauses = append(whereClauses, fmt.Sprintf("tenant_id = $%d", argCount))
	args = append(args, params.TenantID)
	argCount++

	// If keyword is provided, add fuzzy search condition on name
	if params.Keyword != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argCount))
		args = append(args, "%"+params.Keyword+"%")
		argCount++
	}

	// If IsActive is provided, add filter condition on is_active
	if params.IsActive != nil {
		whereClauses = append(whereClauses, fmt.Sprintf("is_active = $%d", argCount))
		args = append(args, *params.IsActive)
		argCount++
	}

	// Always exclude soft-deleted records
	whereClauses = append(whereClauses, "deleted_at IS NULL")

	whereStr := "WHERE " + strings.Join(whereClauses, " AND ")

	// --- Execute COUNT query ---
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM environments %s", whereStr)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count environments: %w", err)
	}

	if total == 0 {
		return []*domain.Environment{}, 0, nil
	}

	// --- Execute SELECT query ---
	listQuery := fmt.Sprintf(`
        SELECT id, name, sdk_key, is_active, created_at, updated_at
        FROM environments
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, whereStr, argCount, argCount+1)

	listArgs := append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list environments: %w", err)
	}
	defer rows.Close()

	var environments []*domain.Environment
	for rows.Next() {
		var env domain.Environment
		if err := rows.Scan(
			&env.ID,
			&env.Name,
			&env.SDKKey,
			&env.IsActive,
			&env.CreatedAt,
			&env.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan environment row for list: %w", err)
		}
		environments = append(environments, &env)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating environment rows: %w", err)
	}

	return environments, total, nil
}

// CountByTenant counts the number of environments in the specified tenant (excluding soft-deleted)
func (r *environmentRepository) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	query := `SELECT COUNT(*) FROM environments WHERE tenant_id = $1 AND deleted_at IS NULL`
	var count int64
	err := r.pool.QueryRow(ctx, query, tenantID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count environments by tenant: %w", err)
	}
	return count, nil
}
