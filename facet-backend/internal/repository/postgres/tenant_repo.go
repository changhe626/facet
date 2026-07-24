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

// tenantRepository implements the domain.TenantRepository interface
type tenantRepository struct {
	pool *pgxpool.Pool
}

// NewTenantRepository creates a new TenantRepository instance
func NewTenantRepository(pool *pgxpool.Pool) domain.TenantRepository {
	return &tenantRepository{pool: pool}
}

// Create ... (this method remains unchanged)
func (r *tenantRepository) Create(ctx context.Context, tenant *domain.Tenant) error {
	query := `
        INSERT INTO tenants (id, level_id, name, description, website, status, created_by, updated_by)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        RETURNING created_at, updated_at;
    `
	err := r.pool.QueryRow(ctx, query,
		tenant.ID,
		tenant.LevelID,
		tenant.Name,
		tenant.Description,
		tenant.Website,
		tenant.Status,
		tenant.CreatedBy,
		tenant.UpdatedBy,
	).Scan(&tenant.CreatedAt, &tenant.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create tenant: %w", err)
	}
	return nil
}

// GetByID ... (this method remains unchanged)
func (r *tenantRepository) GetByID(ctx context.Context, id string) (*domain.Tenant, error) {
	query := `
        SELECT id, level_id, name, description, website, status, created_by, updated_by, created_at, updated_at
        FROM tenants
        WHERE id = $1 AND deleted_at IS NULL;
    `
	row := r.pool.QueryRow(ctx, query, id)
	var tenant domain.Tenant
	err := row.Scan(
		&tenant.ID,
		&tenant.LevelID,
		&tenant.Name,
		&tenant.Description,
		&tenant.Website,
		&tenant.Status,
		&tenant.CreatedBy,
		&tenant.UpdatedBy,
		&tenant.CreatedAt,
		&tenant.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTenantNotFound
		}
		return nil, fmt.Errorf("failed to scan tenant row: %w", err)
	}
	return &tenant, nil
}

// Update ... (this method remains unchanged)
func (r *tenantRepository) Update(ctx context.Context, tenant *domain.Tenant) error {
	query := `
        UPDATE tenants
        SET
            level_id = $1,
            name = $2,
            description = $3,
            website = $4,
            status = $5,
            updated_by = $6,
            updated_at = $7
        WHERE
            id = $8;
    `
	cmdTag, err := r.pool.Exec(ctx, query,
		tenant.LevelID,
		tenant.Name,
		tenant.Description,
		tenant.Website,
		tenant.Status,
		tenant.UpdatedBy,
		time.Now(),
		tenant.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update tenant: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrTenantNotFound
	}
	return nil
}

// CheckNameExistsExcludingID ... (this method remains unchanged)
func (r *tenantRepository) CheckNameExistsExcludingID(ctx context.Context, name, idToExclude string) (bool, error) {
	var whereClauses []string
	var args []interface{}
	argCount := 1
	whereClauses = append(whereClauses, fmt.Sprintf("name = $%d", argCount))
	args = append(args, name)
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
            FROM tenants
            WHERE %s
        );
    `, strings.Join(whereClauses, " AND "))
	var exists bool
	err := r.pool.QueryRow(ctx, query, args...).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if tenant name exists: %w", err)
	}
	return exists, nil
}

// ListAndCount lists and searches tenants by conditions with pagination (excluding soft-deleted)
func (r *tenantRepository) ListAndCount(ctx context.Context, params domain.ListTenantsParams) ([]*domain.Tenant, int64, error) {
	var total int64

	// --- 1. Build dynamic WHERE clause and parameters ---
	var whereClauses []string
	var args []interface{}
	argCount := 1

	if params.Name != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("name ILIKE $%d", argCount))
		args = append(args, "%"+params.Name+"%")
		argCount++
	}

	whereClauses = append(whereClauses, "deleted_at IS NULL")

	whereStr := "WHERE " + strings.Join(whereClauses, " AND ")

	// --- 2. Execute COUNT query ---
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tenants %s", whereStr)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tenants: %w", err)
	}

	if total == 0 {
		return []*domain.Tenant{}, 0, nil
	}

	// --- 3. Execute SELECT query ---
	listQuery := fmt.Sprintf(`
        SELECT id, name, level_id, status, created_at
        FROM tenants
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, whereStr, argCount, argCount+1)

	listArgs := append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tenants: %w", err)
	}
	defer rows.Close()

	// --- 4. Scan results ---
	var tenants []*domain.Tenant
	for rows.Next() {
		var tenant domain.Tenant
		if err := rows.Scan(
			&tenant.ID,
			&tenant.Name,
			&tenant.LevelID,
			&tenant.Status,
			&tenant.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan tenant row for list: %w", err)
		}
		tenants = append(tenants, &tenant)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating tenant rows: %w", err)
	}

	return tenants, total, nil
}
