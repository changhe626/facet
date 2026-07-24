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

// userRepository implements the domain.UserRepository interface
type userRepository struct {
	pool *pgxpool.Pool
}

// NewUserRepository creates a new UserRepository instance
func NewUserRepository(pool *pgxpool.Pool) domain.UserRepository {
	return &userRepository{pool: pool}
}

// GetUserByEmail ... (this method remains unchanged)
func (r *userRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
        SELECT id, tenant_id, role, email, password_hash, display_name, last_login_at, created_by, updated_by, created_at, updated_at
        FROM users
        WHERE email = $1 AND deleted_at IS NULL;
    `
	row := r.pool.QueryRow(ctx, query, email)
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.TenantID,
		&user.Role,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.LastLoginAt,
		&user.CreatedBy,
		&user.UpdatedBy,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to scan user row: %w", err)
	}
	return &user, nil
}

// ListAndCount lists and searches users by conditions with pagination (added email search)
func (r *userRepository) ListAndCount(ctx context.Context, params domain.ListUsersParams) ([]*domain.User, int64, error) {
	var total int64
	var whereClauses []string
	var args []interface{}
	argCount := 1

	// Always filter by tenant ID
	if params.TenantID != "" {
		whereClauses = append(whereClauses, fmt.Sprintf("tenant_id = $%d", argCount))
		args = append(args, params.TenantID)
		argCount++
	}

	// If keyword is provided, add fuzzy search condition on display_name or email
	if params.Keyword != "" {
		searchCondition := fmt.Sprintf("(display_name ILIKE $%d OR email ILIKE $%d)", argCount, argCount)
		whereClauses = append(whereClauses, searchCondition)
		args = append(args, "%"+params.Keyword+"%")
		argCount++
	}

	// Always exclude soft-deleted records
	whereClauses = append(whereClauses, "deleted_at IS NULL")

	whereStr := "WHERE " + strings.Join(whereClauses, " AND ")

	// --- Execute COUNT query ---
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM users %s", whereStr)
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count users: %w", err)
	}

	if total == 0 {
		return []*domain.User{}, 0, nil
	}

	// --- Execute SELECT query ---
	listQuery := fmt.Sprintf(`
        SELECT id, tenant_id, role, email, display_name, created_at, last_login_at
        FROM users
        %s
        ORDER BY created_at DESC
        LIMIT $%d OFFSET $%d
    `, whereStr, argCount, argCount+1)

	listArgs := append(args, params.Limit, params.Offset)

	rows, err := r.pool.Query(ctx, listQuery, listArgs...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(
			&user.ID,
			&user.TenantID,
			&user.Role,
			&user.Email,
			&user.DisplayName,
			&user.CreatedAt,
			&user.LastLoginAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan user row for list: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("error iterating user rows: %w", err)
	}

	return users, total, nil
}

// Create ... (this method remains unchanged)
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
        INSERT INTO users (id, tenant_id, role, email, password_hash, display_name, created_by, updated_by, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
        RETURNING created_at, updated_at;
    `
	err := r.pool.QueryRow(ctx, query,
		user.ID,
		user.TenantID,
		user.Role,
		user.Email,
		user.PasswordHash,
		user.DisplayName,
		user.CreatedBy,
		user.UpdatedBy,
		user.CreatedAt,
		user.UpdatedAt,
	).Scan(&user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

// Delete ... (this method remains unchanged)
func (r *userRepository) Delete(ctx context.Context, id string) error {
	query := `
        UPDATE users
        SET deleted_at = $1
        WHERE id = $2 AND deleted_at IS NULL;
    `
	cmdTag, err := r.pool.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to soft delete user: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// UpdatePassword ... (this method has been fixed)
func (r *userRepository) UpdatePassword(ctx context.Context, id, newPasswordHash string) error {
	query := `
        UPDATE users
        SET password_hash = $1, updated_at = NOW()
        WHERE id = $2 AND deleted_at IS NULL;
    `
	cmdTag, err := r.pool.Exec(ctx, query, newPasswordHash, id)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}
	return nil
}

// GetByID ... (this method remains unchanged)
func (r *userRepository) GetByID(ctx context.Context, id, tenantID string) (*domain.User, error) {
	query := `
        SELECT id, tenant_id, role, email, display_name, created_at, last_login_at, updated_at, created_by, updated_by
        FROM users
        WHERE id = $1 AND tenant_id = $2 AND deleted_at IS NULL;
    `
	row := r.pool.QueryRow(ctx, query, id, tenantID)
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.TenantID,
		&user.Role,
		&user.Email,
		&user.DisplayName,
		&user.CreatedAt,
		&user.LastLoginAt,
		&user.UpdatedAt,
		&user.CreatedBy,
		&user.UpdatedBy,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to scan user detail row: %w", err)
	}
	return &user, nil
}

// CheckEmailExistsInTenant ... (this method remains unchanged)
func (r *userRepository) CheckEmailExistsInTenant(ctx context.Context, email, tenantID string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM users
            WHERE email = $1 AND tenant_id = $2 AND deleted_at IS NULL
        );
    `
	var exists bool
	err := r.pool.QueryRow(ctx, query, email, tenantID).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if email exists in tenant: %w", err)
	}
	return exists, nil
}

// FindByTenantIDAndRole ... (this method remains unchanged)
func (r *userRepository) FindByTenantIDAndRole(ctx context.Context, tenantID string, role domain.UserRole) (*domain.User, error) {
	query := `
        SELECT id, tenant_id, role, email, password_hash, display_name, last_login_at, created_by, updated_by, created_at, updated_at
        FROM users
        WHERE tenant_id = $1 AND role = $2 AND deleted_at IS NULL;
    `
	row := r.pool.QueryRow(ctx, query, tenantID, role)

	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.TenantID,
		&user.Role,
		&user.Email,
		&user.PasswordHash,
		&user.DisplayName,
		&user.LastLoginAt,
		&user.CreatedBy,
		&user.UpdatedBy,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to scan user row by tenant and role: %w", err)
	}

	return &user, nil
}

// ListByIDs ... (this method remains unchanged)
func (r *userRepository) ListByIDs(ctx context.Context, ids []string) ([]*domain.User, error) {
	if len(ids) == 0 {
		return []*domain.User{}, nil
	}

	query := `
        SELECT id, display_name, email
        FROM users
        WHERE id = ANY($1) AND deleted_at IS NULL;
    `
	rows, err := r.pool.Query(ctx, query, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to list users by ids: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.ID, &user.DisplayName, &user.Email); err != nil {
			return nil, fmt.Errorf("failed to scan user row for list by ids: %w", err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating user rows for list by ids: %w", err)
	}

	return users, nil
}

// Update ... (this method remains unchanged)
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
        UPDATE users
        SET email = $1, display_name = $2, role = $3, updated_by = $4, updated_at = NOW()
        WHERE id = $5 AND tenant_id = $6 AND deleted_at IS NULL;
    `
	cmdTag, err := r.pool.Exec(ctx, query,
		user.Email,
		user.DisplayName,
		user.Role,
		user.UpdatedBy,
		user.ID,
		user.TenantID,
	)

	if err != nil {
		return fmt.Errorf("failed to update user: %w", err)
	}

	if cmdTag.RowsAffected() == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

// CheckEmailExistsInTenantExcludingUser checks if an email is already taken by another user in the specified tenant
func (r *userRepository) CheckEmailExistsInTenantExcludingUser(ctx context.Context, email, tenantID, userIDToExclude string) (bool, error) {
	query := `
        SELECT EXISTS (
            SELECT 1
            FROM users
            WHERE email = $1 AND tenant_id = $2 AND id != $3 AND deleted_at IS NULL
        );
    `
	var exists bool
	err := r.pool.QueryRow(ctx, query, email, tenantID, userIDToExclude).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check if email exists in tenant excluding user: %w", err)
	}

	return exists, nil
}

// CountByTenantAndRole counts the number of users with a specific role in the specified tenant (excluding soft-deleted)
func (r *userRepository) CountByTenantAndRole(ctx context.Context, tenantID string, role domain.UserRole) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE tenant_id = $1 AND role = $2 AND deleted_at IS NULL`
	var count int64
	err := r.pool.QueryRow(ctx, query, tenantID, role).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count users by tenant and role: %w", err)
	}
	return count, nil
}
