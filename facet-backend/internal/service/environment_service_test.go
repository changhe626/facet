package service

import (
	"context"
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/cache"
	"facet-backend/internal/domain"
	"testing"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// mockCacheService implements cache.Service for testing
type mockCacheService struct {
	mock.Mock
}

func (m *mockCacheService) GetEnvironmentConfig(ctx context.Context, sdkKey string) (*cache.LeanEnvironmentConfig, error) {
	args := m.Called(ctx, sdkKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cache.LeanEnvironmentConfig), args.Error(1)
}

func (m *mockCacheService) SetEnvironmentConfig(ctx context.Context, sdkKey string, payload *cache.LeanEnvironmentConfig) error {
	args := m.Called(ctx, sdkKey, payload)
	return args.Error(0)
}

func (m *mockCacheService) DeleteEnvironmentConfig(ctx context.Context, sdkKey string) error {
	args := m.Called(ctx, sdkKey)
	return args.Error(0)
}

func (m *mockCacheService) GetTenantLevelByID(ctx context.Context, levelID string) (*domain.TenantLevel, error) {
	args := m.Called(ctx, levelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantLevel), args.Error(1)
}

func (m *mockCacheService) GetTenantLevelByCode(ctx context.Context, levelCode string) (*domain.TenantLevel, error) {
	args := m.Called(ctx, levelCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantLevel), args.Error(1)
}

func (m *mockCacheService) SetTenantLevel(ctx context.Context, level *domain.TenantLevel) error {
	args := m.Called(ctx, level)
	return args.Error(0)
}

func (m *mockCacheService) DeleteTenantLevel(ctx context.Context, level *domain.TenantLevel) error {
	args := m.Called(ctx, level)
	return args.Error(0)
}

// mockTenantLevelService implements TenantLevelService for testing
type mockTenantLevelService struct {
	mock.Mock
}

func (m *mockTenantLevelService) CanCreateEnvironment(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *mockTenantLevelService) CanCreateFeatureFlag(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *mockTenantLevelService) CanCreateUser(ctx context.Context, tenantID string, role domain.UserRole) error {
	args := m.Called(ctx, tenantID, role)
	return args.Error(0)
}

// MockEnvironmentRepository is a mock implementation of the domain.EnvironmentRepository interface.
type MockEnvironmentRepository struct {
	mock.Mock
}

func (m *MockEnvironmentRepository) Create(ctx context.Context, env *domain.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) Update(ctx context.Context, env *domain.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) Delete(ctx context.Context, id string, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *MockEnvironmentRepository) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Environment, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) GetByID(ctx context.Context, id string, tenantID string) (*domain.Environment, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) CheckNameExistsExcludingID(ctx context.Context, name, tenantID, idToExclude string) (bool, error) {
	args := m.Called(ctx, name, tenantID, idToExclude)
	return args.Bool(0), args.Error(1)
}

func (m *MockEnvironmentRepository) FindBySDKKey(ctx context.Context, sdkKey string) (*domain.Environment, error) {
	args := m.Called(ctx, sdkKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Environment), args.Error(1)
}

func (m *MockEnvironmentRepository) ListAndCount(ctx context.Context, params domain.ListEnvironmentsParams) ([]*domain.Environment, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.Environment), args.Get(1).(int64), args.Error(2)
}

func (m *MockEnvironmentRepository) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func TestEnvironmentService_ListActive(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New().String()
	mockEnvs := []*domain.Environment{
		{ID: "env-1", Name: "Production", IsActive: true},
		{ID: "env-2", Name: "Staging", IsActive: false},
		{ID: "env-3", Name: "Development", IsActive: true},
	}

	t.Run("Success - list only active environments", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockRepo.On("ListByTenant", ctx, tenantID).Return(mockEnvs, nil).Once()

		node, _ := snowflake.NewNode(1)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.ListActive(ctx, tenantID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Len(t, result, 2)
		assert.Equal(t, "env-1", result[0].ID)
		assert.Equal(t, "Production", result[0].Name)
		assert.Equal(t, "env-3", result[1].ID)
		assert.Equal(t, "Development", result[1].Name)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - repository error", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		dbError := errors.New("db error")
		mockRepo.On("ListByTenant", ctx, tenantID).Return(nil, dbError).Once()

		node, _ := snowflake.NewNode(1)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.ListActive(ctx, tenantID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to list active environments from repository")
		mockRepo.AssertExpectations(t)
	})
}

func TestEnvironmentService_List(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New().String()
	req := request.ListEnvironmentsRequest{Page: 1, PageSize: 10}
	params := domain.ListEnvironmentsParams{
		TenantID: tenantID,
		Keyword:  "",
		IsActive: nil,
		Limit:    10,
		Offset:   0,
	}
	mockEnvs := []*domain.Environment{
		{ID: "env-1", Name: "Production", SDKKey: "sdk-1", IsActive: true, CreatedAt: time.Now(), UpdatedAt: time.Now()},
	}

	t.Run("Success - list environments", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockRepo.On("ListAndCount", ctx, params).Return(mockEnvs, int64(1), nil).Once()

		node, _ := snowflake.NewNode(1)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.List(ctx, tenantID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "env-1", result.Items[0].ID)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - repository error", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		dbError := errors.New("db error")
		mockRepo.On("ListAndCount", ctx, params).Return(nil, int64(0), dbError).Once()

		node, _ := snowflake.NewNode(1)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.List(ctx, tenantID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to list environments from repository")
		mockRepo.AssertExpectations(t)
	})
}

func TestEnvironmentService_Create(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New().String()
	creatorID := uuid.New().String()
	req := request.CreateEnvironmentRequest{Name: "New Env", Description: "A new test env"}

	t.Run("Success - create environment", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		mockTenantLevel.On("CanCreateEnvironment", ctx, tenantID).Return(nil).Once()
		mockRepo.On("CheckNameExistsExcludingID", ctx, req.Name, tenantID, "").Return(false, nil).Once()
		mockRepo.On("Create", ctx, mock.AnythingOfType("*domain.Environment")).Return(nil).Once()

		node, _ := snowflake.NewNode(1)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.Create(ctx, tenantID, creatorID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Name, result.Name)
		assert.True(t, result.IsActive)
		mockRepo.AssertExpectations(t)
		mockTenantLevel.AssertExpectations(t)
	})

	t.Run("Failure - name already exists", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		mockTenantLevel.On("CanCreateEnvironment", ctx, tenantID).Return(nil).Once()
		mockRepo.On("CheckNameExistsExcludingID", ctx, req.Name, tenantID, "").Return(true, nil).Once()

		node, _ := snowflake.NewNode(1)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.Create(ctx, tenantID, creatorID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrEnvironmentNameExists, err)
		mockRepo.AssertExpectations(t)
		mockTenantLevel.AssertExpectations(t)
	})

	t.Run("Failure - repository error on check", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		mockTenantLevel.On("CanCreateEnvironment", ctx, tenantID).Return(nil).Once()
		dbError := errors.New("db error")
		mockRepo.On("CheckNameExistsExcludingID", ctx, req.Name, tenantID, "").Return(false, dbError).Once()

		node, _ := snowflake.NewNode(1)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.Create(ctx, tenantID, creatorID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to check environment name existence")
		mockRepo.AssertExpectations(t)
		mockTenantLevel.AssertExpectations(t)
	})
}

func TestEnvironmentService_Update(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New().String()
	envID := "env-" + uuid.New().String()
	updaterID := uuid.New().String()
	isActive := false
	req := request.UpdateEnvironmentRequest{Name: "Updated Env", Description: "Updated desc", IsActive: &isActive}

	mockEnv := &domain.Environment{
		ID:       envID,
		TenantID: tenantID,
		Name:     "Old Name",
	}

	t.Run("Success - update environment", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		mockRepo.On("CheckNameExistsExcludingID", ctx, req.Name, tenantID, envID).Return(false, nil).Once()
		mockRepo.On("GetByID", ctx, envID, tenantID).Return(mockEnv, nil).Once()
		mockRepo.On("Update", ctx, mock.AnythingOfType("*domain.Environment")).Run(func(args mock.Arguments) {
			envArg := args.Get(1).(*domain.Environment)
			assert.Equal(t, req.Name, envArg.Name)
			assert.Equal(t, *req.IsActive, envArg.IsActive)
			assert.Equal(t, updaterID, envArg.UpdatedBy.String)
		}).Return(nil).Once()
		mockCache.On("DeleteEnvironmentConfig", ctx, mock.AnythingOfType("string")).Return(nil).Once()

		node, _ := snowflake.NewNode(1)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.Update(ctx, tenantID, envID, updaterID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Name, result.Name)
		assert.False(t, result.IsActive)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - name already exists", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockRepo.On("CheckNameExistsExcludingID", ctx, req.Name, tenantID, envID).Return(true, nil).Once()

		node, _ := snowflake.NewNode(1)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.Update(ctx, tenantID, envID, updaterID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrEnvironmentNameExists, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - environment not found", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockRepo.On("CheckNameExistsExcludingID", ctx, req.Name, tenantID, envID).Return(false, nil).Once()
		mockRepo.On("GetByID", ctx, envID, tenantID).Return(nil, domain.ErrEnvironmentNotFound).Once()

		node, _ := snowflake.NewNode(1)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		result, err := svc.Update(ctx, tenantID, envID, updaterID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrEnvironmentNotFound, err)
		mockRepo.AssertExpectations(t)
	})
}

func TestEnvironmentService_Delete(t *testing.T) {
	ctx := context.Background()
	tenantID := uuid.New().String()
	envID := "env-" + uuid.New().String()

	t.Run("Success - delete environment", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		mockRepo.On("GetByID", ctx, envID, tenantID).Return(&domain.Environment{SDKKey: "sdk-test"}, nil).Once()
		mockRepo.On("Delete", ctx, envID, tenantID).Return(nil).Once()
		mockCache.On("DeleteEnvironmentConfig", ctx, "sdk-test").Return(nil).Once()

		node, _ := snowflake.NewNode(1)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		err := svc.Delete(ctx, tenantID, envID)

		assert.NoError(t, err)
		mockRepo.AssertExpectations(t)
		mockCache.AssertExpectations(t)
	})

	t.Run("Failure - environment not found", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		mockRepo.On("GetByID", ctx, envID, tenantID).Return(nil, domain.ErrEnvironmentNotFound).Once()

		node, _ := snowflake.NewNode(1)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		err := svc.Delete(ctx, tenantID, envID)

		assert.Error(t, err)
		assert.Equal(t, domain.ErrEnvironmentNotFound, err)
		mockRepo.AssertExpectations(t)
	})

	t.Run("Failure - other repository error", func(t *testing.T) {
		mockRepo := new(MockEnvironmentRepository)
		mockCache := new(mockCacheService)
		mockTenantLevel := new(mockTenantLevelService)
		mockRepo.On("GetByID", ctx, envID, tenantID).Return(&domain.Environment{SDKKey: "sdk-test"}, nil).Once()
		dbError := errors.New("db connection error")
		mockRepo.On("Delete", ctx, envID, tenantID).Return(dbError).Once()

		node, _ := snowflake.NewNode(1)
		svc := NewEnvironmentService(mockRepo, node, mockCache, mockTenantLevel)

		err := svc.Delete(ctx, tenantID, envID)

		assert.Error(t, err)
		assert.NotEqual(t, domain.ErrEnvironmentNotFound, err)
		assert.Contains(t, err.Error(), "failed to delete environment in repository")
		mockRepo.AssertExpectations(t)
	})
}
