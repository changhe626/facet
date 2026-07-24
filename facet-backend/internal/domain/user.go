package domain

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

var (
	ErrUserNotFound             = errors.New("user not found")
	ErrEmailExistsInTenant      = errors.New("email already exists in this tenant")
	ErrRoleInTenant             = errors.New("role not exists in this tenant")
	ErrTenantSuperAdminNotFound = errors.New("tenant super admin not found") // New error
)

// UserRole defines the enum type for user roles
type UserRole string

const (
	RoleSystemAdmin      UserRole = "system_admin"
	RoleTenantSuperAdmin UserRole = "tenant_super_admin"
	RoleTenantAdmin      UserRole = "tenant_admin"
	RoleTenantViewer     UserRole = "tenant_viewer"
)

// User defines the core domain model for a user
type User struct {
	ID           string         `db:"id"`
	TenantID     sql.NullString `db:"tenant_id"` // This field is NULL for system admins
	Role         UserRole       `db:"role"`
	Email        string         `db:"email"`
	PasswordHash string         `db:"password_hash"`
	DisplayName  sql.NullString `db:"display_name"`
	LastLoginAt  sql.NullTime   `db:"last_login_at"`
	CreatedBy    sql.NullString `db:"created_by"`
	UpdatedBy    sql.NullString `db:"updated_by"`
	CreatedAt    time.Time      `db:"created_at"`
	UpdatedAt    time.Time      `db:"updated_at"`
	DeletedAt    sql.NullTime   `db:"deleted_at"`
}

// ListUsersParams defines the parameters for querying users
type ListUsersParams struct {
	TenantID string
	UserID   string
	Keyword  string
	Limit    int
	Offset   int
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	ListAndCount(ctx context.Context, params ListUsersParams) ([]*User, int64, error)
	Create(ctx context.Context, user *User) error
	Delete(ctx context.Context, id string) error
	UpdatePassword(ctx context.Context, id, newPasswordHash string) error
	GetByID(ctx context.Context, id, tenantID string) (*User, error)
	CheckEmailExistsInTenant(ctx context.Context, email, tenantID string) (bool, error)
	CheckEmailExistsInTenantExcludingUser(ctx context.Context, email, tenantID, userIDToExclude string) (bool, error) // <-- New method
	FindByTenantIDAndRole(ctx context.Context, tenantID string, role UserRole) (*User, error)
	ListByIDs(ctx context.Context, ids []string) ([]*User, error)
	Update(ctx context.Context, user *User) error
	CountByTenantAndRole(ctx context.Context, tenantID string, role UserRole) (int64, error)
}
