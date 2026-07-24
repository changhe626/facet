package service

import (
	"context"
	"database/sql"
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/domain"
	"facet-backend/pkg/password"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserRepository is a mock implementation of the domain.UserRepository interface.
// It allows us to simulate database interactions without needing a real database connection.
type MockUserRepository struct {
	mock.Mock
}

// GetUserByEmail is the mock implementation for fetching a user.
func (m *MockUserRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	// 'Called' records the call and returns the results that were configured for this call.
	args := m.Called(ctx, email)
	// Return the user object if it's not nil, otherwise return the error.
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// --- Stub implementations for other interface methods (not used in this test) ---

func (m *MockUserRepository) Create(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) GetByID(ctx context.Context, id, tenantID string) (*domain.User, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

// CheckEmailExistsInTenant implements the missing method in the UserRepository interface.
// Even though this test does not use it directly, it exists to make MockUserRepository satisfy the interface.
func (m *MockUserRepository) CheckEmailExistsInTenant(ctx context.Context, email, tenantID string) (bool, error) {
	args := m.Called(ctx, email, tenantID)
	return args.Bool(0), args.Error(1)
}

// --- Remaining methods added to satisfy interface completeness ---

// ListAndCount implements the missing method in the UserRepository interface.
func (m *MockUserRepository) ListAndCount(ctx context.Context, params domain.ListUsersParams) ([]*domain.User, int64, error) {
	args := m.Called(ctx, params)
	var users []*domain.User
	if arg0 := args.Get(0); arg0 != nil {
		users = arg0.([]*domain.User)
	}
	return users, args.Get(1).(int64), args.Error(2)
}

// Delete implements the missing method in the UserRepository interface.
func (m *MockUserRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// UpdatePassword implements the missing method in the UserRepository interface.
func (m *MockUserRepository) UpdatePassword(ctx context.Context, id string, newPasswordHash string) error {
	args := m.Called(ctx, id, newPasswordHash)
	return args.Error(0)
}

// --- Remaining methods added to satisfy interface completeness (continued) ---

func (m *MockUserRepository) CheckEmailExistsInTenantExcludingUser(ctx context.Context, email, tenantID, userIDToExclude string) (bool, error) {
	args := m.Called(ctx, email, tenantID, userIDToExclude)
	return args.Bool(0), args.Error(1)
}

func (m *MockUserRepository) FindByTenantIDAndRole(ctx context.Context, tenantID string, role domain.UserRole) (*domain.User, error) {
	args := m.Called(ctx, tenantID, role)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.User), args.Error(1)
}

func (m *MockUserRepository) ListByIDs(ctx context.Context, ids []string) ([]*domain.User, error) {
	args := m.Called(ctx, ids)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.User), args.Error(1)
}

func (m *MockUserRepository) Update(ctx context.Context, user *domain.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

func (m *MockUserRepository) CountByTenantAndRole(ctx context.Context, tenantID string, role domain.UserRole) (int64, error) {
	args := m.Called(ctx, tenantID, role)
	return args.Get(0).(int64), args.Error(1)
}

// TestAuthService_Login uses a table-driven approach to test the Login method.
func TestAuthService_Login(t *testing.T) {
	// --- Test Setup ---
	ctx := context.Background()
	secretKey := "test-secret-for-jwt"
	expiresIn := 1 // 1 hour

	// Create a sample user and hash their password for realistic testing.
	userID := uuid.New().String()
	tenantID := uuid.New().String()
	userEmail := "test@example.com"
	userPassword := "password123"
	hashedPassword, _ := password.HashPassword(userPassword)

	mockUser := &domain.User{
		ID:           userID,
		TenantID:     sql.NullString{String: tenantID, Valid: true},
		Email:        userEmail,
		PasswordHash: hashedPassword,
		Role:         domain.RoleTenantAdmin, // Use the role constant defined in the domain
		DisplayName:  sql.NullString{String: "Test User", Valid: true},
	}

	// --- Test Cases ---
	testCases := []struct {
		name          string
		request       request.LoginRequest
		setupMock     func(mockRepo *MockUserRepository)
		expectError   bool
		expectedError error
	}{
		{
			name: "Success - Correct credentials",
			request: request.LoginRequest{
				Email:    userEmail,
				Password: userPassword,
			},
			setupMock: func(mockRepo *MockUserRepository) {
				// Expect a call to GetUserByEmail and return the mock user.
				mockRepo.On("GetUserByEmail", ctx, userEmail).Return(mockUser, nil).Once()
			},
			expectError: false,
		},
		{
			name: "Failure - User not found",
			request: request.LoginRequest{
				Email:    "notfound@example.com",
				Password: "password123",
			},
			setupMock: func(mockRepo *MockUserRepository) {
				// Expect a call and return the specific 'ErrUserNotFound' error.
				mockRepo.On("GetUserByEmail", ctx, "notfound@example.com").Return(nil, domain.ErrUserNotFound).Once()
			},
			expectError:   true,
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "Failure - Incorrect password",
			request: request.LoginRequest{
				Email:    userEmail,
				Password: "wrong-password",
			},
			setupMock: func(mockRepo *MockUserRepository) {
				// The repository call is successful, but the password check in the service will fail.
				mockRepo.On("GetUserByEmail", ctx, userEmail).Return(mockUser, nil).Once()
			},
			expectError:   true,
			expectedError: ErrInvalidCredentials,
		},
		{
			name: "Failure - Database error on user lookup",
			request: request.LoginRequest{
				Email:    userEmail,
				Password: userPassword,
			},
			setupMock: func(mockRepo *MockUserRepository) {
				// Simulate a generic database error.
				dbError := errors.New("database connection lost")
				mockRepo.On("GetUserByEmail", ctx, userEmail).Return(nil, dbError).Once()
			},
			expectError:   true,
			expectedError: fmt.Errorf("failed to get user: %w", errors.New("database connection lost")),
		},
	}

	// --- Run Tests ---
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockRepo := new(MockUserRepository)
			tc.setupMock(mockRepo)

			authSvc := NewAuthService(mockRepo, secretKey, expiresIn)
			loginResponse, err := authSvc.Login(ctx, tc.request)

			if tc.expectError {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedError.Error(), "Error message should match expected")
				assert.Nil(t, loginResponse, "Response should be nil on error")
			} else {
				assert.NoError(t, err, "Should not return an error on success")
				assert.NotNil(t, loginResponse, "Response should not be nil on success")
				assert.NotEmpty(t, loginResponse.Token, "Token should be generated")
				assert.Equal(t, userID, loginResponse.User.ID, "User ID in response should match")
				assert.Equal(t, string(domain.RoleTenantAdmin), loginResponse.User.Role, "User role in response should match")
			}

			// Verify that all expected mock calls were made.
			mockRepo.AssertExpectations(t)
		})
	}
}
