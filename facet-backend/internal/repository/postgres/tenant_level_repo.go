package postgres

import (
	"context"
	"errors"
	"facet-backend/internal/domain"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// tenantLevelRepository implements the domain.TenantLevelRepository interface
type tenantLevelRepository struct {
	pool *pgxpool.Pool
}

// NewTenantLevelRepository creates a new TenantLevelRepository instance
func NewTenantLevelRepository(pool *pgxpool.Pool) domain.TenantLevelRepository {
	return &tenantLevelRepository{pool: pool}
}

// scanTenantLevel is a helper function for scanning data from pgx.Row into a TenantLevel struct
func scanTenantLevel(row pgx.Row) (*domain.TenantLevel, error) {
	var level domain.TenantLevel
	err := row.Scan(
		&level.ID,
		&level.Code,
		&level.Name,
		&level.MaxEnvironments,
		&level.MaxFeatureFlags,
		&level.MaxTenantAdmins,
		&level.MaxTenantViewers,
		&level.AllowSSO,
		&level.AllowAPIAccess,
		&level.HasAuditLog,
		&level.PrioritySupport,
		&level.PriceMonthly,
		&level.PriceYearly,
		&level.IsPublic,
		&level.IsActive,
		&level.CreatedAt,
		&level.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTenantLevelNotFound
		}
		return nil, fmt.Errorf("failed to scan tenant level row: %w", err)
	}
	return &level, nil
}

// GetByID gets plan details by level ID
func (r *tenantLevelRepository) GetByID(ctx context.Context, id string) (*domain.TenantLevel, error) {
	query := `SELECT * FROM tenant_levels WHERE id = $1`
	row := r.pool.QueryRow(ctx, query, id)
	return scanTenantLevel(row)
}

// GetByCode gets plan details by the unique level code
func (r *tenantLevelRepository) GetByCode(ctx context.Context, code string) (*domain.TenantLevel, error) {
	query := `SELECT * FROM tenant_levels WHERE code = $1`
	row := r.pool.QueryRow(ctx, query, code)
	return scanTenantLevel(row)
}

// ListPublic lists all public, active plans
func (r *tenantLevelRepository) ListPublic(ctx context.Context) ([]*domain.TenantLevel, error) {
	query := `SELECT * FROM tenant_levels WHERE is_public = true AND is_active = true ORDER BY price_monthly ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query public tenant levels: %w", err)
	}
	defer rows.Close()

	var levels []*domain.TenantLevel
	for rows.Next() {
		level, err := scanTenantLevel(rows)
		if err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}

	return levels, rows.Err()
}

// ListAll lists all plans
func (r *tenantLevelRepository) ListAll(ctx context.Context) ([]*domain.TenantLevel, error) {
	query := `SELECT * FROM tenant_levels ORDER BY price_monthly ASC`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query all tenant levels: %w", err)
	}
	defer rows.Close()

	var levels []*domain.TenantLevel
	for rows.Next() {
		level, err := scanTenantLevel(rows)
		if err != nil {
			return nil, err
		}
		levels = append(levels, level)
	}

	return levels, rows.Err()
}
