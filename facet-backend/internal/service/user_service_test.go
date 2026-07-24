package service

import (
	"context"
	"database/sql"
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/domain"
	"facet-backend/pkg/password"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockUserTenantLevelService implements TenantLevelService for user tests
type mockUserTenantLevelService struct {
	mock.Mock
}

func (m *mockUserTenantLevelService) CanCreateEnvironment(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *mockUserTenantLevelService) CanCreateFeatureFlag(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *mockUserTenantLevelService) CanCreateUser(ctx context.Context, tenantID string, role domain.UserRole) error {
	args := m.Called(ctx, tenantID, role)
	return args.Error(0)
}

func setupUserServiceWithMocks(t *testing.T) (*userService, *MockUserRepository, *mockUserTenantLevelService) {
	mockUserRepo := new(MockUserRepository)
	mockTenantLevel := new(mockUserTenantLevelService)
	node, err := snowflake.NewNode(1)
	assert.NoError(t, err)

	svc := NewUserService(mockUserRepo, node, mockTenantLevel).(*userService)
	return svc, mockUserRepo, mockTenantLevel
}

func TestUserService_ListUsers(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-" + uuid.New().String()
	userID := "user-" + uuid.New().String()
	req := request.ListUsersRequest{Page: 1, PageSize: 10}

	mockUsers := []*domain.User{
		{ID: "user-1", DisplayName: sql.NullString{String: "Test User 1", Valid: true}, Email: "test1@example.com", Role: domain.RoleTenantAdmin, CreatedAt: time.Now()},
	}

	testCases := []struct {
		name           string
		role           domain.UserRole
		setupMock      func(mockRepo *MockUserRepository)
		expectedParams domain.ListUsersParams
		expectError    bool
	}{
		{
			name: "Success - System Admin lists all users",
			role: domain.RoleSystemAdmin,
			expectedParams: domain.ListUsersParams{
				TenantID: "", // System admin sees all
				UserID:   "",
				Keyword:  "",
				Limit:    10,
				Offset:   0,
			},
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("ListAndCount", ctx, mock.Anything).Return(mockUsers, int64(1), nil).Once()
			},
			expectError: false,
		},
		{
			name: "Success - Tenant Admin lists users in their tenant",
			role: domain.RoleTenantAdmin,
			expectedParams: domain.ListUsersParams{
				TenantID: tenantID, // Tenant admin sees their tenant
				UserID:   "",
				Keyword:  "",
				Limit:    10,
				Offset:   0,
			},
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("ListAndCount", ctx, mock.Anything).Return(mockUsers, int64(1), nil).Once()
			},
			expectError: false,
		},
		{
			name: "Success - Tenant Viewer lists only themselves",
			role: domain.RoleTenantViewer,
			expectedParams: domain.ListUsersParams{
				TenantID: "",
				UserID:   userID, // Viewer only sees themself
				Keyword:  "",
				Limit:    10,
				Offset:   0,
			},
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("ListAndCount", ctx, mock.Anything).Return(mockUsers, int64(1), nil).Once()
			},
			expectError: false,
		},
		{
			name: "Success - Invalid role returns empty list",
			role: "invalid_role",
			setupMock: func(mockRepo *MockUserRepository) {
				// No call to repository is expected
			},
			expectError: false,
		},
		{
			name: "Failure - Repository error",
			role: domain.RoleSystemAdmin,
			expectedParams: domain.ListUsersParams{
				TenantID: "",
				UserID:   "",
				Keyword:  "",
				Limit:    10,
				Offset:   0,
			},
			setupMock: func(mockRepo *MockUserRepository) {
				mockRepo.On("ListAndCount", ctx, mock.Anything).Return(nil, int64(0), errors.New("db error")).Once()
			},
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			svc, mockUserRepo, _ := setupUserServiceWithMocks(t)
			tc.setupMock(mockUserRepo)

			result, err := svc.ListUsers(ctx, tenantID, userID, tc.role, req)

			if tc.expectError {
				assert.Error(t, err)
				assert.Nil(t, result)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				if tc.role != "invalid_role" {
					// Check if the correct params were passed to the mock
					assert.Equal(t, tc.expectedParams, mockUserRepo.Calls[0].Arguments.Get(1).(domain.ListUsersParams))
				}
			}
			mockUserRepo.AssertExpectations(t)
		})
	}
}

func TestUserService_Create(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-" + uuid.New().String()
	creatorID := "user-" + uuid.New().String()
	req := request.CreateUserRequest{
		Email:           "new.user@example.com",
		InitialPassword: "password123",
		DisplayName:     "New User",
		Role:            string(domain.RoleTenantAdmin),
	}

	t.Run("Success - create user", func(t *testing.T) {
		svc, mockUserRepo, mockTenantLevel := setupUserServiceWithMocks(t)

		mockTenantLevel.On("CanCreateUser", ctx, tenantID, domain.RoleTenantAdmin).Return(nil).Once()
		mockUserRepo.On("CheckEmailExistsInTenant", ctx, req.Email, tenantID).Return(false, nil).Once()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(nil).Once()

		result, err := svc.Create(ctx, tenantID, creatorID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Email, result.Email)
		assert.Equal(t, req.DisplayName, result.DisplayName)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - email already exists", func(t *testing.T) {
		svc, mockUserRepo, mockTenantLevel := setupUserServiceWithMocks(t)

		mockTenantLevel.On("CanCreateUser", ctx, tenantID, domain.RoleTenantAdmin).Return(nil).Once()
		mockUserRepo.On("CheckEmailExistsInTenant", ctx, req.Email, tenantID).Return(true, nil).Once()

		result, err := svc.Create(ctx, tenantID, creatorID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrEmailExistsInTenant, err)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - unique constraint violation on create", func(t *testing.T) {
		svc, mockUserRepo, mockTenantLevel := setupUserServiceWithMocks(t)
		pgErr := &pgconn.PgError{Code: "23505"} // unique_violation

		mockTenantLevel.On("CanCreateUser", ctx, tenantID, domain.RoleTenantAdmin).Return(nil).Once()
		mockUserRepo.On("CheckEmailExistsInTenant", ctx, req.Email, tenantID).Return(false, nil).Once()
		mockUserRepo.On("Create", ctx, mock.AnythingOfType("*domain.User")).Return(pgErr).Once()

		result, err := svc.Create(ctx, tenantID, creatorID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrEmailExistsInTenant, err)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_Delete(t *testing.T) {
	ctx := context.Background()
	userID := "user-to-delete"

	t.Run("Success - delete user", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)
		mockUserRepo.On("Delete", ctx, userID).Return(nil).Once()

		err := svc.Delete(ctx, userID)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - user not found", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)
		mockUserRepo.On("Delete", ctx, userID).Return(domain.ErrUserNotFound).Once()

		err := svc.Delete(ctx, userID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_ResetPassword(t *testing.T) {
	ctx := context.Background()
	userID := "user-to-reset"
	newPassword := "new-strong-password"

	t.Run("Success - reset password", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)

		mockUserRepo.On("UpdatePassword", ctx, userID, mock.AnythingOfType("string")).Run(func(args mock.Arguments) {
			hashedPassword := args.String(2)
			err := password.CheckPassword(newPassword, hashedPassword)
			assert.NoError(t, err, "The new password should be correctly hashed")
		}).Return(nil).Once()

		err := svc.ResetPassword(ctx, userID, newPassword)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - user not found", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)
		mockUserRepo.On("UpdatePassword", ctx, userID, mock.AnythingOfType("string")).Return(domain.ErrUserNotFound).Once()

		err := svc.ResetPassword(ctx, userID, newPassword)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_GetByID(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-" + uuid.New().String()
	userID := "user-" + uuid.New().String()
	mockUser := &domain.User{
		ID:          userID,
		Email:       "get.me@example.com",
		DisplayName: sql.NullString{String: "Get Me", Valid: true},
		Role:        domain.RoleTenantAdmin,
		CreatedAt:   time.Now(),
		LastLoginAt: sql.NullTime{Time: time.Now().Add(-time.Hour), Valid: true},
	}

	t.Run("Success - get user by id", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)
		mockUserRepo.On("GetByID", ctx, userID, tenantID).Return(mockUser, nil).Once()

		result, err := svc.GetByID(ctx, tenantID, userID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, mockUser.ID, result.ID)
		assert.Equal(t, mockUser.Email, result.Email)
		assert.Equal(t, mockUser.DisplayName.String, result.DisplayName)

		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - user not found", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)
		mockUserRepo.On("GetByID", ctx, userID, tenantID).Return(nil, domain.ErrUserNotFound).Once()

		result, err := svc.GetByID(ctx, tenantID, userID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrUserNotFound, err)

		mockUserRepo.AssertExpectations(t)
	})
}

func TestUserService_Update(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-" + uuid.New().String()
	userID := "user-to-update"
	updaterID := "user-updater"
	req := request.UpdateUserRequest{
		Email:       "updated.email@example.com",
		DisplayName: "Updated Name",
		Role:        string(domain.RoleTenantViewer),
	}
	mockUser := &domain.User{ID: userID, Email: "original@example.com"}

	t.Run("Success - update user", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)

		mockUserRepo.On("CheckEmailExistsInTenantExcludingUser", ctx, req.Email, tenantID, userID).Return(false, nil).Once()
		mockUserRepo.On("GetByID", ctx, userID, tenantID).Return(mockUser, nil).Once()
		mockUserRepo.On("Update", ctx, mock.AnythingOfType("*domain.User")).Run(func(args mock.Arguments) {
			userArg := args.Get(1).(*domain.User)
			assert.Equal(t, req.Email, userArg.Email)
			assert.Equal(t, req.DisplayName, userArg.DisplayName.String)
			assert.Equal(t, domain.UserRole(req.Role), userArg.Role)
		}).Return(nil).Once()

		err := svc.Update(ctx, tenantID, userID, updaterID, req)

		assert.NoError(t, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - new email already exists", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)

		mockUserRepo.On("CheckEmailExistsInTenantExcludingUser", ctx, req.Email, tenantID, userID).Return(true, nil).Once()

		err := svc.Update(ctx, tenantID, userID, updaterID, req)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrEmailExistsInTenant, err)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - user to update not found", func(t *testing.T) {
		svc, mockUserRepo, _ := setupUserServiceWithMocks(t)

		mockUserRepo.On("CheckEmailExistsInTenantExcludingUser", ctx, req.Email, tenantID, userID).Return(false, nil).Once()
		mockUserRepo.On("GetByID", ctx, userID, tenantID).Return(nil, domain.ErrUserNotFound).Once()

		err := svc.Update(ctx, tenantID, userID, updaterID, req)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrUserNotFound, err)
		mockUserRepo.AssertExpectations(t)
	})
}
