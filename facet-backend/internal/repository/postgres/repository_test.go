//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"facet-backend/internal/domain"
	"os"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestPool returns a pgxpool connection to the test database.
// If DATABASE_URL is not set, it skips the test.
func getTestPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("POSTGRES_DSN")
	}
	if dsn == "" {
		t.Skip("Skipping integration test: DATABASE_URL or POSTGRES_DSN not set")
	}

	ctx := context.Background()
	pool, err := pgxpool.New(ctx, dsn)
	require.NoError(t, err, "failed to connect to test database")

	t.Cleanup(func() {
		pool.Close()
	})

	return pool
}

// getTestSnowflakeNode returns a snowflake node for generating IDs in tests.
func getTestSnowflakeNode(t *testing.T) *snowflake.Node {
	t.Helper()
	node, err := snowflake.NewNode(1)
	require.NoError(t, err, "failed to create snowflake node")
	return node
}

// --- Tenant Repository Tests ---

func TestTenantRepository_CreateAndGetByID(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewTenantRepository(pool)
	node := getTestSnowflakeNode(t)

	tenant := &domain.Tenant{
		ID:          node.Generate().String(),
		LevelID:     "test-level-id",
		Name:        "Test Tenant " + node.Generate().String(),
		Description: sql.NullString{String: "A test tenant", Valid: true},
		Website:     sql.NullString{String: "https://example.com", Valid: true},
		Status:      domain.TenantStatusActive,
		CreatedBy:   sql.NullString{String: "system", Valid: true},
		UpdatedBy:   sql.NullString{String: "system", Valid: true},
	}

	t.Run("Create - success", func(t *testing.T) {
		err := repo.Create(ctx, tenant)
		require.NoError(t, err)
		assert.False(t, tenant.CreatedAt.IsZero(), "CreatedAt should be set")
		assert.False(t, tenant.UpdatedAt.IsZero(), "UpdatedAt should be set")
	})

	t.Run("GetByID - success", func(t *testing.T) {
		result, err := repo.GetByID(ctx, tenant.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, tenant.ID, result.ID)
		assert.Equal(t, tenant.Name, result.Name)
		assert.Equal(t, tenant.LevelID, result.LevelID)
		assert.Equal(t, tenant.Status, result.Status)
		assert.Equal(t, tenant.Description.String, result.Description.String)
		assert.Equal(t, tenant.Website.String, result.Website.String)
	})

	t.Run("GetByID - not found", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "nonexistent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrTenantNotFound)
		assert.Nil(t, result)
	})

	t.Run("GetByID - soft deleted tenant", func(t *testing.T) {
		// Soft delete the tenant
		_, err := pool.Exec(ctx, `UPDATE tenants SET deleted_at = NOW() WHERE id = $1`, tenant.ID)
		require.NoError(t, err)

		// Should not be found
		result, err := repo.GetByID(ctx, tenant.ID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrTenantNotFound)
		assert.Nil(t, result)
	})
}

func TestTenantRepository_Update(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewTenantRepository(pool)
	node := getTestSnowflakeNode(t)

	tenant := &domain.Tenant{
		ID:          node.Generate().String(),
		LevelID:     "level-1",
		Name:        "Tenant To Update " + node.Generate().String(),
		Description: sql.NullString{String: "Original description", Valid: true},
		Website:     sql.NullString{String: "https://original.com", Valid: true},
		Status:      domain.TenantStatusActive,
		CreatedBy:   sql.NullString{String: "system", Valid: true},
		UpdatedBy:   sql.NullString{String: "system", Valid: true},
	}
	err := repo.Create(ctx, tenant)
	require.NoError(t, err)

	t.Run("Update - success", func(t *testing.T) {
		tenant.Name = "Updated Tenant Name"
		tenant.Description = sql.NullString{String: "Updated description", Valid: true}
		tenant.Website = sql.NullString{String: "https://updated.com", Valid: true}
		tenant.Status = domain.TenantStatusSuspended
		tenant.UpdatedBy = sql.NullString{String: "admin", Valid: true}

		err := repo.Update(ctx, tenant)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, tenant.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Updated Tenant Name", result.Name)
		assert.Equal(t, "Updated description", result.Description.String)
		assert.Equal(t, "https://updated.com", result.Website.String)
		assert.Equal(t, domain.TenantStatusSuspended, result.Status)
	})

	t.Run("Update - not found", func(t *testing.T) {
		nonexistent := &domain.Tenant{
			ID: "nonexistent-id",
		}
		err := repo.Update(ctx, nonexistent)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrTenantNotFound)
	})
}

func TestTenantRepository_CheckNameExistsExcludingID(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewTenantRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantName := "Unique-Tenant-Name-" + node.Generate().String()
	tenant := &domain.Tenant{
		ID:        node.Generate().String(),
		LevelID:   "level-1",
		Name:      tenantName,
		Status:    domain.TenantStatusActive,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	err := repo.Create(ctx, tenant)
	require.NoError(t, err)

	t.Run("CheckNameExists - name exists", func(t *testing.T) {
		exists, err := repo.CheckNameExistsExcludingID(ctx, tenantName, "")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("CheckNameExists - name does not exist", func(t *testing.T) {
		exists, err := repo.CheckNameExistsExcludingID(ctx, "NonExistentName", "")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CheckNameExists - excluding self returns false", func(t *testing.T) {
		exists, err := repo.CheckNameExistsExcludingID(ctx, tenantName, tenant.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestTenantRepository_ListAndCount(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewTenantRepository(pool)
	node := getTestSnowflakeNode(t)

	// Create some test tenants
	tenant1 := &domain.Tenant{
		ID:        node.Generate().String(),
		LevelID:   "level-1",
		Name:      "Alpha-Tenant-" + node.Generate().String(),
		Status:    domain.TenantStatusActive,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	tenant2 := &domain.Tenant{
		ID:        node.Generate().String(),
		LevelID:   "level-2",
		Name:      "Beta-Tenant-" + node.Generate().String(),
		Status:    domain.TenantStatusActive,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, tenant1))
	require.NoError(t, repo.Create(ctx, tenant2))

	t.Run("ListAndCount - all tenants", func(t *testing.T) {
		tenants, total, err := repo.ListAndCount(ctx, domain.ListTenantsParams{
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.GreaterOrEqual(t, len(tenants), 2)
	})

	t.Run("ListAndCount - pagination", func(t *testing.T) {
		tenants, total, err := repo.ListAndCount(ctx, domain.ListTenantsParams{
			Limit:  1,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.Len(t, tenants, 1)
	})

	t.Run("ListAndCount - search by name", func(t *testing.T) {
		tenants, total, err := repo.ListAndCount(ctx, domain.ListTenantsParams{
			Name:   "Alpha",
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, tenants, 1)
		assert.Contains(t, tenants[0].Name, "Alpha")
	})

	t.Run("ListAndCount - no results", func(t *testing.T) {
		tenants, total, err := repo.ListAndCount(ctx, domain.ListTenantsParams{
			Name:   "NonExistentTenantName",
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, tenants)
	})
}

// --- User Repository Tests ---

func TestUserRepository_CreateAndGetByID(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantAdmin,
		Email:        "test-" + node.Generate().String() + "@example.com",
		PasswordHash: "hashed-password",
		DisplayName:  sql.NullString{String: "Test User", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	t.Run("Create - success", func(t *testing.T) {
		err := repo.Create(ctx, user)
		require.NoError(t, err)
	})

	t.Run("GetByID - success", func(t *testing.T) {
		result, err := repo.GetByID(ctx, user.ID, tenantID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, user.Email, result.Email)
		assert.Equal(t, user.DisplayName.String, result.DisplayName.String)
		assert.Equal(t, user.Role, result.Role)
	})

	t.Run("GetByID - not found with wrong tenant", func(t *testing.T) {
		result, err := repo.GetByID(ctx, user.ID, "wrong-tenant-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, result)
	})
}

func TestUserRepository_GetUserByEmail(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	email := "findme-" + node.Generate().String() + "@example.com"
	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: node.Generate().String(), Valid: true},
		Role:         domain.RoleTenantAdmin,
		Email:        email,
		PasswordHash: "hashed-password",
		DisplayName:  sql.NullString{String: "Find Me", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("GetUserByEmail - success", func(t *testing.T) {
		result, err := repo.GetUserByEmail(ctx, email)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, email, result.Email)
		assert.Equal(t, user.PasswordHash, result.PasswordHash)
	})

	t.Run("GetUserByEmail - not found", func(t *testing.T) {
		result, err := repo.GetUserByEmail(ctx, "nonexistent@example.com")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, result)
	})
}

func TestUserRepository_Update(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantViewer,
		Email:        "update-test-" + node.Generate().String() + "@example.com",
		PasswordHash: "original-hash",
		DisplayName:  sql.NullString{String: "Original Name", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("Update - success", func(t *testing.T) {
		user.Email = "updated-" + node.Generate().String() + "@example.com"
		user.DisplayName = sql.NullString{String: "Updated Name", Valid: true}
		user.Role = domain.RoleTenantAdmin
		user.UpdatedBy = sql.NullString{String: "admin", Valid: true}

		err := repo.Update(ctx, user)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, user.ID, tenantID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Updated Name", result.DisplayName.String)
		assert.Equal(t, domain.RoleTenantAdmin, result.Role)
	})

	t.Run("Update - not found", func(t *testing.T) {
		nonexistent := &domain.User{
			ID:       "nonexistent-id",
			TenantID: sql.NullString{String: tenantID, Valid: true},
		}
		err := repo.Update(ctx, nonexistent)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

func TestUserRepository_Delete(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: node.Generate().String(), Valid: true},
		Role:         domain.RoleTenantViewer,
		Email:        "delete-test-" + node.Generate().String() + "@example.com",
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "To Delete", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("Delete - success", func(t *testing.T) {
		err := repo.Delete(ctx, user.ID)
		require.NoError(t, err)

		// Verify soft delete
		result, err := repo.GetByID(ctx, user.ID, user.TenantID.String)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, result)
	})

	t.Run("Delete - not found", func(t *testing.T) {
		err := repo.Delete(ctx, "nonexistent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

func TestUserRepository_UpdatePassword(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: node.Generate().String(), Valid: true},
		Role:         domain.RoleTenantViewer,
		Email:        "password-test-" + node.Generate().String() + "@example.com",
		PasswordHash: "old-hash",
		DisplayName:  sql.NullString{String: "Password Test", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("UpdatePassword - success", func(t *testing.T) {
		err := repo.UpdatePassword(ctx, user.ID, "new-hashed-password")
		require.NoError(t, err)

		// Verify by fetching the user with email (which includes password_hash)
		result, err := repo.GetUserByEmail(ctx, user.Email)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "new-hashed-password", result.PasswordHash)
	})

	t.Run("UpdatePassword - not found", func(t *testing.T) {
		err := repo.UpdatePassword(ctx, "nonexistent-id", "any-hash")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
	})
}

func TestUserRepository_ListAndCount(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	user1 := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantAdmin,
		Email:        "list-a-" + node.Generate().String() + "@example.com",
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "Alice", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantViewer,
		Email:        "list-b-" + node.Generate().String() + "@example.com",
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "Bob", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user1))
	require.NoError(t, repo.Create(ctx, user2))

	t.Run("ListAndCount - by tenant", func(t *testing.T) {
		users, total, err := repo.ListAndCount(ctx, domain.ListUsersParams{
			TenantID: tenantID,
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, users, 2)
	})

	t.Run("ListAndCount - search by keyword", func(t *testing.T) {
		users, total, err := repo.ListAndCount(ctx, domain.ListUsersParams{
			TenantID: tenantID,
			Keyword:  "Alice",
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, users, 1)
		assert.Equal(t, "Alice", users[0].DisplayName.String)
	})

	t.Run("ListAndCount - search by email", func(t *testing.T) {
		users, total, err := repo.ListAndCount(ctx, domain.ListUsersParams{
			TenantID: tenantID,
			Keyword:  user1.Email,
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, users, 1)
		assert.Equal(t, user1.Email, users[0].Email)
	})

	t.Run("ListAndCount - no results for wrong tenant", func(t *testing.T) {
		users, total, err := repo.ListAndCount(ctx, domain.ListUsersParams{
			TenantID: "nonexistent-tenant",
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, users)
	})

	t.Run("ListAndCount - pagination", func(t *testing.T) {
		users, total, err := repo.ListAndCount(ctx, domain.ListUsersParams{
			TenantID: tenantID,
			Limit:    1,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, users, 1)
	})
}

func TestUserRepository_CheckEmailExistsInTenant(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	email := "exists-test-" + node.Generate().String() + "@example.com"
	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantViewer,
		Email:        email,
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "Exists Test", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("CheckEmailExistsInTenant - exists", func(t *testing.T) {
		exists, err := repo.CheckEmailExistsInTenant(ctx, email, tenantID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("CheckEmailExistsInTenant - not exists", func(t *testing.T) {
		exists, err := repo.CheckEmailExistsInTenant(ctx, "nonexistent@example.com", tenantID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CheckEmailExistsInTenant - wrong tenant", func(t *testing.T) {
		exists, err := repo.CheckEmailExistsInTenant(ctx, email, "other-tenant")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestUserRepository_CheckEmailExistsInTenantExcludingUser(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	email := "exclude-test-" + node.Generate().String() + "@example.com"
	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantViewer,
		Email:        email,
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "Exclude Test", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("CheckEmailExistsInTenantExcludingUser - excludes self", func(t *testing.T) {
		exists, err := repo.CheckEmailExistsInTenantExcludingUser(ctx, email, tenantID, user.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CheckEmailExistsInTenantExcludingUser - finds other user", func(t *testing.T) {
		exists, err := repo.CheckEmailExistsInTenantExcludingUser(ctx, email, tenantID, "other-user-id")
		require.NoError(t, err)
		assert.True(t, exists)
	})
}

func TestUserRepository_FindByTenantIDAndRole(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantSuperAdmin,
		Email:        "superadmin-" + node.Generate().String() + "@example.com",
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "Super Admin", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("FindByTenantIDAndRole - success", func(t *testing.T) {
		result, err := repo.FindByTenantIDAndRole(ctx, tenantID, domain.RoleTenantSuperAdmin)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, user.ID, result.ID)
		assert.Equal(t, domain.RoleTenantSuperAdmin, result.Role)
	})

	t.Run("FindByTenantIDAndRole - not found", func(t *testing.T) {
		result, err := repo.FindByTenantIDAndRole(ctx, tenantID, domain.RoleTenantAdmin)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrUserNotFound)
		assert.Nil(t, result)
	})
}

func TestUserRepository_ListByIDs(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	user1 := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantViewer,
		Email:        "listbyids-a-" + node.Generate().String() + "@example.com",
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "User A", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	user2 := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantViewer,
		Email:        "listbyids-b-" + node.Generate().String() + "@example.com",
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "User B", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user1))
	require.NoError(t, repo.Create(ctx, user2))

	t.Run("ListByIDs - multiple IDs", func(t *testing.T) {
		users, err := repo.ListByIDs(ctx, []string{user1.ID, user2.ID})
		require.NoError(t, err)
		assert.Len(t, users, 2)
	})

	t.Run("ListByIDs - single ID", func(t *testing.T) {
		users, err := repo.ListByIDs(ctx, []string{user1.ID})
		require.NoError(t, err)
		assert.Len(t, users, 1)
		assert.Equal(t, user1.ID, users[0].ID)
	})

	t.Run("ListByIDs - empty IDs", func(t *testing.T) {
		users, err := repo.ListByIDs(ctx, []string{})
		require.NoError(t, err)
		assert.Empty(t, users)
	})

	t.Run("ListByIDs - nonexistent IDs", func(t *testing.T) {
		users, err := repo.ListByIDs(ctx, []string{"nonexistent-id"})
		require.NoError(t, err)
		assert.Empty(t, users)
	})
}

func TestUserRepository_CountByTenantAndRole(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewUserRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	user := &domain.User{
		ID:           node.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Role:         domain.RoleTenantSuperAdmin,
		Email:        "count-test-" + node.Generate().String() + "@example.com",
		PasswordHash: "hash",
		DisplayName:  sql.NullString{String: "Count Test", Valid: true},
		CreatedBy:    sql.NullString{String: "system", Valid: true},
		UpdatedBy:    sql.NullString{String: "system", Valid: true},
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
	require.NoError(t, repo.Create(ctx, user))

	t.Run("CountByTenantAndRole - found", func(t *testing.T) {
		count, err := repo.CountByTenantAndRole(ctx, tenantID, domain.RoleTenantSuperAdmin)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("CountByTenantAndRole - not found", func(t *testing.T) {
		count, err := repo.CountByTenantAndRole(ctx, tenantID, domain.RoleTenantAdmin)
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// --- Environment Repository Tests ---

func TestEnvironmentRepository_CreateAndGetByID(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewEnvironmentRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	env := &domain.Environment{
		ID:          node.Generate().String(),
		TenantID:    tenantID,
		Name:        "Test Env " + node.Generate().String(),
		Description: sql.NullString{String: "A test environment", Valid: true},
		SDKKey:      "sdk-" + node.Generate().String(),
		IsActive:    true,
		CreatedBy:   sql.NullString{String: "system", Valid: true},
		UpdatedBy:   sql.NullString{String: "system", Valid: true},
	}

	t.Run("Create - success", func(t *testing.T) {
		err := repo.Create(ctx, env)
		require.NoError(t, err)
	})

	t.Run("GetByID - success", func(t *testing.T) {
		result, err := repo.GetByID(ctx, env.ID, tenantID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, env.ID, result.ID)
		assert.Equal(t, env.Name, result.Name)
		assert.Equal(t, env.SDKKey, result.SDKKey)
		assert.True(t, result.IsActive)
	})

	t.Run("GetByID - not found", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "nonexistent-id", tenantID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEnvironmentNotFound)
		assert.Nil(t, result)
	})

	t.Run("GetByID - wrong tenant", func(t *testing.T) {
		result, err := repo.GetByID(ctx, env.ID, "wrong-tenant")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEnvironmentNotFound)
		assert.Nil(t, result)
	})
}

func TestEnvironmentRepository_Update(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewEnvironmentRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	env := &domain.Environment{
		ID:          node.Generate().String(),
		TenantID:    tenantID,
		Name:        "Env To Update " + node.Generate().String(),
		Description: sql.NullString{String: "Original desc", Valid: true},
		SDKKey:      "sdk-update-" + node.Generate().String(),
		IsActive:    true,
		CreatedBy:   sql.NullString{String: "system", Valid: true},
		UpdatedBy:   sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, env))

	t.Run("Update - success", func(t *testing.T) {
		env.Name = "Updated Env Name"
		env.IsActive = false
		env.Description = sql.NullString{String: "Updated desc", Valid: true}
		env.UpdatedBy = sql.NullString{String: "admin", Valid: true}

		err := repo.Update(ctx, env)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, env.ID, tenantID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Updated Env Name", result.Name)
		assert.False(t, result.IsActive)
		assert.Equal(t, "Updated desc", result.Description.String)
	})

	t.Run("Update - not found", func(t *testing.T) {
		nonexistent := &domain.Environment{
			ID:       "nonexistent-id",
			TenantID: tenantID,
		}
		err := repo.Update(ctx, nonexistent)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEnvironmentNotFound)
	})
}

func TestEnvironmentRepository_Delete(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewEnvironmentRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	env := &domain.Environment{
		ID:        node.Generate().String(),
		TenantID:  tenantID,
		Name:      "Env To Delete " + node.Generate().String(),
		SDKKey:    "sdk-delete-" + node.Generate().String(),
		IsActive:  true,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, env))

	t.Run("Delete - success", func(t *testing.T) {
		err := repo.Delete(ctx, env.ID, tenantID)
		require.NoError(t, err)

		// Verify soft delete
		result, err := repo.GetByID(ctx, env.ID, tenantID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEnvironmentNotFound)
		assert.Nil(t, result)
	})

	t.Run("Delete - not found", func(t *testing.T) {
		err := repo.Delete(ctx, "nonexistent-id", tenantID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEnvironmentNotFound)
	})
}

func TestEnvironmentRepository_ListByTenant(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewEnvironmentRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	env1 := &domain.Environment{
		ID:        node.Generate().String(),
		TenantID:  tenantID,
		Name:      "Env A " + node.Generate().String(),
		SDKKey:    "sdk-a-" + node.Generate().String(),
		IsActive:  true,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	env2 := &domain.Environment{
		ID:        node.Generate().String(),
		TenantID:  tenantID,
		Name:      "Env B " + node.Generate().String(),
		SDKKey:    "sdk-b-" + node.Generate().String(),
		IsActive:  false,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, env1))
	require.NoError(t, repo.Create(ctx, env2))

	t.Run("ListByTenant - success", func(t *testing.T) {
		envs, err := repo.ListByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Len(t, envs, 2)
	})

	t.Run("ListByTenant - no environments", func(t *testing.T) {
		envs, err := repo.ListByTenant(ctx, "nonexistent-tenant")
		require.NoError(t, err)
		assert.Empty(t, envs)
	})
}

func TestEnvironmentRepository_FindBySDKKey(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewEnvironmentRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	sdkKey := "sdk-find-" + node.Generate().String()
	env := &domain.Environment{
		ID:        node.Generate().String(),
		TenantID:  tenantID,
		Name:      "SDK Find Env " + node.Generate().String(),
		SDKKey:    sdkKey,
		IsActive:  true,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, env))

	t.Run("FindBySDKKey - success", func(t *testing.T) {
		result, err := repo.FindBySDKKey(ctx, sdkKey)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, env.ID, result.ID)
		assert.Equal(t, sdkKey, result.SDKKey)
	})

	t.Run("FindBySDKKey - not found", func(t *testing.T) {
		result, err := repo.FindBySDKKey(ctx, "nonexistent-sdk-key")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrEnvironmentInvalidSDKKey)
		assert.Nil(t, result)
	})
}

func TestEnvironmentRepository_CheckNameExistsExcludingID(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewEnvironmentRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	envName := "Unique-Env-Name-" + node.Generate().String()
	env := &domain.Environment{
		ID:        node.Generate().String(),
		TenantID:  tenantID,
		Name:      envName,
		SDKKey:    "sdk-check-" + node.Generate().String(),
		IsActive:  true,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, env))

	t.Run("CheckNameExists - name exists in tenant", func(t *testing.T) {
		exists, err := repo.CheckNameExistsExcludingID(ctx, envName, tenantID, "")
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("CheckNameExists - name does not exist", func(t *testing.T) {
		exists, err := repo.CheckNameExistsExcludingID(ctx, "NonExistentEnvName", tenantID, "")
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CheckNameExists - excluding self returns false", func(t *testing.T) {
		exists, err := repo.CheckNameExistsExcludingID(ctx, envName, tenantID, env.ID)
		require.NoError(t, err)
		assert.False(t, exists)
	})

	t.Run("CheckNameExists - same name different tenant returns false", func(t *testing.T) {
		exists, err := repo.CheckNameExistsExcludingID(ctx, envName, "other-tenant", "")
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestEnvironmentRepository_ListAndCount(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewEnvironmentRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	env1 := &domain.Environment{
		ID:        node.Generate().String(),
		TenantID:  tenantID,
		Name:      "Production " + node.Generate().String(),
		SDKKey:    "sdk-prod-" + node.Generate().String(),
		IsActive:  true,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	env2 := &domain.Environment{
		ID:        node.Generate().String(),
		TenantID:  tenantID,
		Name:      "Staging " + node.Generate().String(),
		SDKKey:    "sdk-staging-" + node.Generate().String(),
		IsActive:  false,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, env1))
	require.NoError(t, repo.Create(ctx, env2))

	t.Run("ListAndCount - all", func(t *testing.T) {
		envs, total, err := repo.ListAndCount(ctx, domain.ListEnvironmentsParams{
			TenantID: tenantID,
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, envs, 2)
	})

	t.Run("ListAndCount - filter by active", func(t *testing.T) {
		active := true
		envs, total, err := repo.ListAndCount(ctx, domain.ListEnvironmentsParams{
			TenantID: tenantID,
			IsActive: &active,
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, envs, 1)
		assert.True(t, envs[0].IsActive)
	})

	t.Run("ListAndCount - search by keyword", func(t *testing.T) {
		envs, total, err := repo.ListAndCount(ctx, domain.ListEnvironmentsParams{
			TenantID: tenantID,
			Keyword:  "Production",
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, envs, 1)
		assert.Contains(t, envs[0].Name, "Production")
	})

	t.Run("ListAndCount - no results", func(t *testing.T) {
		envs, total, err := repo.ListAndCount(ctx, domain.ListEnvironmentsParams{
			TenantID: "nonexistent-tenant",
			Limit:    10,
			Offset:   0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, envs)
	})
}

func TestEnvironmentRepository_CountByTenant(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewEnvironmentRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	env := &domain.Environment{
		ID:        node.Generate().String(),
		TenantID:  tenantID,
		Name:      "Count Env " + node.Generate().String(),
		SDKKey:    "sdk-count-" + node.Generate().String(),
		IsActive:  true,
		CreatedBy: sql.NullString{String: "system", Valid: true},
		UpdatedBy: sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, env))

	t.Run("CountByTenant - success", func(t *testing.T) {
		count, err := repo.CountByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("CountByTenant - no environments", func(t *testing.T) {
		count, err := repo.CountByTenant(ctx, "nonexistent-tenant")
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

// --- Feature Flag Repository Tests ---

func TestFeatureFlagRepository_CreateAndGetByID(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	envID := node.Generate().String()
	flag := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              "test-flag-" + node.Generate().String(),
		Description:       sql.NullString{String: "A test flag", Valid: true},
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 50,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}

	t.Run("Create - success", func(t *testing.T) {
		err := repo.Create(ctx, flag)
		require.NoError(t, err)
		assert.False(t, flag.CreatedAt.IsZero())
		assert.False(t, flag.UpdatedAt.IsZero())
	})

	t.Run("GetByID - success", func(t *testing.T) {
		result, err := repo.GetByID(ctx, flag.ID, envID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, flag.ID, result.ID)
		assert.Equal(t, flag.Name, result.Name)
		assert.Equal(t, flag.Enabled, result.Enabled)
		assert.Equal(t, flag.FlagType, result.FlagType)
		assert.Equal(t, flag.RolloutPercentage, result.RolloutPercentage)
	})

	t.Run("GetByID - not found", func(t *testing.T) {
		result, err := repo.GetByID(ctx, "nonexistent-id", envID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
		assert.Nil(t, result)
	})

	t.Run("GetByID - wrong environment", func(t *testing.T) {
		result, err := repo.GetByID(ctx, flag.ID, "wrong-env")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
		assert.Nil(t, result)
	})
}

func TestFeatureFlagRepository_Update(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	envID := node.Generate().String()
	flag := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              "update-flag-" + node.Generate().String(),
		Description:       sql.NullString{String: "Original desc", Valid: true},
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 25,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, flag))

	t.Run("Update - success", func(t *testing.T) {
		flag.Description = sql.NullString{String: "Updated desc", Valid: true}
		flag.Enabled = false
		flag.FlagType = domain.FlagTypeExperiment
		flag.RolloutPercentage = 75
		flag.UpdatedBy = sql.NullString{String: "admin", Valid: true}

		err := repo.Update(ctx, flag)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, flag.ID, envID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Updated desc", result.Description.String)
		assert.False(t, result.Enabled)
		assert.Equal(t, domain.FlagTypeExperiment, result.FlagType)
		assert.Equal(t, 75, result.RolloutPercentage)
	})

	t.Run("Update - not found", func(t *testing.T) {
		nonexistent := &domain.FeatureFlag{
			ID: "nonexistent-id",
		}
		err := repo.Update(ctx, nonexistent)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
	})
}

func TestFeatureFlagRepository_Delete(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	envID := node.Generate().String()
	flag := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              "delete-flag-" + node.Generate().String(),
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 0,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, flag))

	t.Run("Delete - success", func(t *testing.T) {
		err := repo.Delete(ctx, flag.ID, envID)
		require.NoError(t, err)

		// Verify soft delete
		result, err := repo.GetByID(ctx, flag.ID, envID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
		assert.Nil(t, result)
	})

	t.Run("Delete - not found", func(t *testing.T) {
		err := repo.Delete(ctx, "nonexistent-id", envID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
	})
}

func TestFeatureFlagRepository_ArchiveAndUnarchive(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	envID := node.Generate().String()
	flag := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              "archive-flag-" + node.Generate().String(),
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 0,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, flag))

	t.Run("Archive - success", func(t *testing.T) {
		err := repo.Archive(ctx, flag.ID)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, flag.ID, envID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.False(t, result.ArchivedAt.Time.IsZero(), "ArchivedAt should be set")
	})

	t.Run("Unarchive - success", func(t *testing.T) {
		err := repo.Unarchive(ctx, flag.ID)
		require.NoError(t, err)

		result, err := repo.GetByID(ctx, flag.ID, envID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.ArchivedAt.Time.IsZero(), "ArchivedAt should be cleared")
	})

	t.Run("Archive - not found", func(t *testing.T) {
		err := repo.Archive(ctx, "nonexistent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
	})

	t.Run("Unarchive - not found", func(t *testing.T) {
		err := repo.Unarchive(ctx, "nonexistent-id")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
	})
}

func TestFeatureFlagRepository_FindByNameAndEnvironmentID(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	envID := node.Generate().String()
	flagName := "find-by-name-" + node.Generate().String()
	flag := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              flagName,
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 0,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, flag))

	t.Run("FindByNameAndEnvironmentID - success", func(t *testing.T) {
		result, err := repo.FindByNameAndEnvironmentID(ctx, flagName, envID, tenantID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, flag.ID, result.ID)
		assert.Equal(t, flagName, result.Name)
	})

	t.Run("FindByNameAndEnvironmentID - not found", func(t *testing.T) {
		result, err := repo.FindByNameAndEnvironmentID(ctx, "nonexistent-name", envID, tenantID)
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
		assert.Nil(t, result)
	})

	t.Run("FindByNameAndEnvironmentID - wrong tenant", func(t *testing.T) {
		result, err := repo.FindByNameAndEnvironmentID(ctx, flagName, envID, "wrong-tenant")
		require.Error(t, err)
		assert.ErrorIs(t, err, domain.ErrFeatureFlagNotFound)
		assert.Nil(t, result)
	})
}

func TestFeatureFlagRepository_CheckNameExistsInTenant(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	flagName := "check-name-" + node.Generate().String()
	flag := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     node.Generate().String(),
		Name:              flagName,
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 0,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, flag))

	t.Run("CheckNameExistsInTenant - exists", func(t *testing.T) {
		exists, err := repo.CheckNameExistsInTenant(ctx, flagName, tenantID)
		require.NoError(t, err)
		assert.True(t, exists)
	})

	t.Run("CheckNameExistsInTenant - not exists", func(t *testing.T) {
		exists, err := repo.CheckNameExistsInTenant(ctx, "nonexistent-name", tenantID)
		require.NoError(t, err)
		assert.False(t, exists)
	})
}

func TestFeatureFlagRepository_FindByEnvironmentID(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	envID := node.Generate().String()
	flag1 := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              "env-flag-a-" + node.Generate().String(),
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 0,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	flag2 := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              "env-flag-b-" + node.Generate().String(),
		Enabled:           false,
		FlagType:          domain.FlagTypeOperational,
		RolloutPercentage: 100,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, flag1))
	require.NoError(t, repo.Create(ctx, flag2))

	t.Run("FindByEnvironmentID - success", func(t *testing.T) {
		flags, err := repo.FindByEnvironmentID(ctx, envID)
		require.NoError(t, err)
		assert.Len(t, flags, 2)
	})

	t.Run("FindByEnvironmentID - no flags", func(t *testing.T) {
		flags, err := repo.FindByEnvironmentID(ctx, "nonexistent-env")
		require.NoError(t, err)
		assert.Empty(t, flags)
	})
}

func TestFeatureFlagRepository_CountByTenant(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	flag := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     node.Generate().String(),
		Name:              "count-flag-" + node.Generate().String(),
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 0,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, flag))

	t.Run("CountByTenant - success", func(t *testing.T) {
		count, err := repo.CountByTenant(ctx, tenantID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})

	t.Run("CountByTenant - no flags", func(t *testing.T) {
		count, err := repo.CountByTenant(ctx, "nonexistent-tenant")
		require.NoError(t, err)
		assert.Equal(t, int64(0), count)
	})
}

func TestFeatureFlagRepository_ListAndCountByEnvironment(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	repo := NewFeatureFlagRepository(pool)
	node := getTestSnowflakeNode(t)

	tenantID := node.Generate().String()
	envID := node.Generate().String()
	flag1 := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              "list-flag-a-" + node.Generate().String(),
		Enabled:           true,
		FlagType:          domain.FlagTypeRelease,
		RolloutPercentage: 50,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	flag2 := &domain.FeatureFlag{
		ID:                node.Generate().String(),
		TenantID:          tenantID,
		EnvironmentID:     envID,
		Name:              "list-flag-b-" + node.Generate().String(),
		Enabled:           false,
		FlagType:          domain.FlagTypeExperiment,
		RolloutPercentage: 0,
		CreatedBy:         sql.NullString{String: "system", Valid: true},
		UpdatedBy:         sql.NullString{String: "system", Valid: true},
	}
	require.NoError(t, repo.Create(ctx, flag1))
	require.NoError(t, repo.Create(ctx, flag2))

	t.Run("ListAndCountByEnvironment - all", func(t *testing.T) {
		flags, total, err := repo.ListAndCountByEnvironment(ctx, domain.ListFeatureFlagsParams{
			EnvironmentID: envID,
			Limit:         10,
			Offset:        0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, flags, 2)
	})

	t.Run("ListAndCountByEnvironment - search by keyword", func(t *testing.T) {
		flags, total, err := repo.ListAndCountByEnvironment(ctx, domain.ListFeatureFlagsParams{
			EnvironmentID: envID,
			Keyword:       "flag-a",
			Limit:         10,
			Offset:        0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(1), total)
		assert.Len(t, flags, 1)
		assert.Contains(t, flags[0].Name, "flag-a")
	})

	t.Run("ListAndCountByEnvironment - no results", func(t *testing.T) {
		flags, total, err := repo.ListAndCountByEnvironment(ctx, domain.ListFeatureFlagsParams{
			EnvironmentID: "nonexistent-env",
			Limit:         10,
			Offset:        0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(0), total)
		assert.Empty(t, flags)
	})

	t.Run("ListAndCountByEnvironment - pagination", func(t *testing.T) {
		flags, total, err := repo.ListAndCountByEnvironment(ctx, domain.ListFeatureFlagsParams{
			EnvironmentID: envID,
			Limit:         1,
			Offset:        0,
		})
		require.NoError(t, err)
		assert.Equal(t, int64(2), total)
		assert.Len(t, flags, 1)
	})
}

// --- Notification Repository Tests ---

func TestSystemNotificationRepository_CreateAndFind(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	node := getTestSnowflakeNode(t)
	repo := NewNotificationRepository(pool, node)

	notif := &domain.SystemNotification{
		ID:      node.Generate().String(),
		Level:   domain.LevelNormal,
		Type:    domain.TypeAnnouncement,
		Title:   "Test Notification",
		Content: "This is a test notification content.",
	}

	t.Run("CreateSystemNotification - success", func(t *testing.T) {
		err := repo.CreateSystemNotification(ctx, notif)
		require.NoError(t, err)
	})

	t.Run("FindSystemNotificationByID - success", func(t *testing.T) {
		result, err := repo.FindSystemNotificationByID(ctx, notif.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, notif.ID, result.ID)
		assert.Equal(t, notif.Title, result.Title)
		assert.Equal(t, notif.Content, result.Content)
		assert.Equal(t, notif.Level, result.Level)
		assert.Equal(t, notif.Type, result.Type)
	})

	t.Run("FindSystemNotificationByID - not found", func(t *testing.T) {
		result, err := repo.FindSystemNotificationByID(ctx, "nonexistent-id")
		require.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestSystemNotificationRepository_Update(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	node := getTestSnowflakeNode(t)
	repo := NewNotificationRepository(pool, node)

	notif := &domain.SystemNotification{
		ID:      node.Generate().String(),
		Level:   domain.LevelNormal,
		Type:    domain.TypeAnnouncement,
		Title:   "Original Title",
		Content: "Original content",
	}
	require.NoError(t, repo.CreateSystemNotification(ctx, notif))

	t.Run("UpdateSystemNotification - success", func(t *testing.T) {
		notif.Level = domain.LevelImportant
		notif.Type = domain.TypeProductUpdate
		notif.Title = "Updated Title"
		notif.Content = "Updated content"

		err := repo.UpdateSystemNotification(ctx, notif)
		require.NoError(t, err)

		result, err := repo.FindSystemNotificationByID(ctx, notif.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Updated Title", result.Title)
		assert.Equal(t, "Updated content", result.Content)
		assert.Equal(t, domain.LevelImportant, result.Level)
		assert.Equal(t, domain.TypeProductUpdate, result.Type)
	})

	t.Run("UpdateSystemNotification - not found", func(t *testing.T) {
		nonexistent := &domain.SystemNotification{
			ID: "nonexistent-id",
		}
		err := repo.UpdateSystemNotification(ctx, nonexistent)
		require.Error(t, err)
	})
}

func TestSystemNotificationRepository_Delete(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	node := getTestSnowflakeNode(t)
	repo := NewNotificationRepository(pool, node)

	notif := &domain.SystemNotification{
		ID:      node.Generate().String(),
		Level:   domain.LevelNormal,
		Type:    domain.TypeAnnouncement,
		Title:   "To Delete",
		Content: "Will be deleted",
	}
	require.NoError(t, repo.CreateSystemNotification(ctx, notif))

	t.Run("DeleteSystemNotification - success", func(t *testing.T) {
		err := repo.DeleteSystemNotification(ctx, notif.ID)
		require.NoError(t, err)

		// Verify soft delete
		result, err := repo.FindSystemNotificationByID(ctx, notif.ID)
		require.Error(t, err)
		assert.Nil(t, result)
	})

	t.Run("DeleteSystemNotification - not found", func(t *testing.T) {
		err := repo.DeleteSystemNotification(ctx, "nonexistent-id")
		require.Error(t, err)
	})
}

func TestSystemNotificationRepository_ListAllSystem(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	node := getTestSnowflakeNode(t)
	repo := NewNotificationRepository(pool, node)

	notif1 := &domain.SystemNotification{
		ID:      node.Generate().String(),
		Level:   domain.LevelNormal,
		Type:    domain.TypeAnnouncement,
		Title:   "List Test 1",
		Content: "Content 1",
	}
	notif2 := &domain.SystemNotification{
		ID:      node.Generate().String(),
		Level:   domain.LevelImportant,
		Type:    domain.TypeSecurityAlert,
		Title:   "List Test 2",
		Content: "Content 2",
	}
	require.NoError(t, repo.CreateSystemNotification(ctx, notif1))
	require.NoError(t, repo.CreateSystemNotification(ctx, notif2))

	t.Run("ListAllSystem - success", func(t *testing.T) {
		notifications, err := repo.ListAllSystem(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, len(notifications), 2)
	})

	t.Run("ListAndCountSystemNotifications - success", func(t *testing.T) {
		notifications, total, err := repo.ListAndCountSystemNotifications(ctx, domain.ListSystemNotificationsParams{
			Limit:  10,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.GreaterOrEqual(t, len(notifications), 2)
	})

	t.Run("ListAndCountSystemNotifications - pagination", func(t *testing.T) {
		notifications, total, err := repo.ListAndCountSystemNotifications(ctx, domain.ListSystemNotificationsParams{
			Limit:  1,
			Offset: 0,
		})
		require.NoError(t, err)
		assert.GreaterOrEqual(t, total, int64(2))
		assert.Len(t, notifications, 1)
	})
}

func TestSystemNotificationRepository_Counts(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	node := getTestSnowflakeNode(t)
	repo := NewNotificationRepository(pool, node)

	notif := &domain.SystemNotification{
		ID:      node.Generate().String(),
		Level:   domain.LevelNormal,
		Type:    domain.TypeAnnouncement,
		Title:   "Count Test",
		Content: "Content",
	}
	require.NoError(t, repo.CreateSystemNotification(ctx, notif))

	t.Run("CountAllSystem - success", func(t *testing.T) {
		count, err := repo.CountAllSystem(ctx)
		require.NoError(t, err)
		assert.GreaterOrEqual(t, count, int64(1))
	})
}

func TestUserNotification_MarkAsRead(t *testing.T) {
	pool := getTestPool(t)
	ctx := context.Background()
	node := getTestSnowflakeNode(t)
	repo := NewNotificationRepository(pool, node)

	notif := &domain.SystemNotification{
		ID:      node.Generate().String(),
		Level:   domain.LevelNormal,
		Type:    domain.TypeAnnouncement,
		Title:   "Mark Read Test",
		Content: "Content",
	}
	require.NoError(t, repo.CreateSystemNotification(ctx, notif))

	tenantID := node.Generate().String()
	userID := node.Generate().String()

	t.Run("MarkAsRead - success", func(t *testing.T) {
		err := repo.MarkAsRead(ctx, tenantID, userID, []string{notif.ID})
		require.NoError(t, err)
	})

	t.Run("MarkAsRead - idempotent", func(t *testing.T) {
		err := repo.MarkAsRead(ctx, tenantID, userID, []string{notif.ID})
		require.NoError(t, err)
	})

	t.Run("FindUserNotifications - shows read status", func(t *testing.T) {
		userNotifs, err := repo.FindUserNotifications(ctx, userID)
		require.NoError(t, err)
		require.Len(t, userNotifs, 1)
		assert.True(t, userNotifs[0].IsRead)
		assert.Equal(t, notif.ID, userNotifs[0].NotificationID)
	})

	t.Run("CountReadForUser - shows correct count", func(t *testing.T) {
		count, err := repo.CountReadForUser(ctx, userID)
		require.NoError(t, err)
		assert.Equal(t, int64(1), count)
	})
}