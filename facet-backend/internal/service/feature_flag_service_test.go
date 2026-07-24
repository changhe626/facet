package service

import (
	"context"
	"database/sql"
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

// MockFeatureFlagRepository is a mock for the domain.FeatureFlagRepository interface.
type MockFeatureFlagRepository struct {
	mock.Mock
}

func (m *MockFeatureFlagRepository) ListAndCountByEnvironment(ctx context.Context, params domain.ListFeatureFlagsParams) ([]*domain.FeatureFlag, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.FeatureFlag), args.Get(1).(int64), args.Error(2)
}

func (m *MockFeatureFlagRepository) Create(ctx context.Context, flag *domain.FeatureFlag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockFeatureFlagRepository) GetByID(ctx context.Context, id, environmentID string) (*domain.FeatureFlag, error) {
	args := m.Called(ctx, id, environmentID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FeatureFlag), args.Error(1)
}

func (m *MockFeatureFlagRepository) Update(ctx context.Context, flag *domain.FeatureFlag) error {
	args := m.Called(ctx, flag)
	return args.Error(0)
}

func (m *MockFeatureFlagRepository) Delete(ctx context.Context, id, environmentID string) error {
	args := m.Called(ctx, id, environmentID)
	return args.Error(0)
}

func (m *MockFeatureFlagRepository) CheckNameExistsInTenant(ctx context.Context, name, tenantID string) (bool, error) {
	args := m.Called(ctx, name, tenantID)
	return args.Bool(0), args.Error(1)
}

func (m *MockFeatureFlagRepository) FindByNameAndEnvironmentID(ctx context.Context, name, environmentID string, tenantID string) (*domain.FeatureFlag, error) {
	args := m.Called(ctx, name, environmentID, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.FeatureFlag), args.Error(1)
}

func (m *MockFeatureFlagRepository) FindByEnvironmentID(ctx context.Context, envID string) ([]*domain.FeatureFlag, error) {
	args := m.Called(ctx, envID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FeatureFlag), args.Error(1)
}

func (m *MockFeatureFlagRepository) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockFeatureFlagRepository) Archive(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFeatureFlagRepository) Unarchive(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockFeatureFlagRuleRepository is a mock for the domain.FeatureFlagRuleRepository interface.
type MockFeatureFlagRuleRepository struct {
	mock.Mock
}

func (m *MockFeatureFlagRuleRepository) Create(ctx context.Context, rule *domain.FeatureFlagRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockFeatureFlagRuleRepository) Update(ctx context.Context, rule *domain.FeatureFlagRule) error {
	args := m.Called(ctx, rule)
	return args.Error(0)
}

func (m *MockFeatureFlagRuleRepository) ListByFeatureFlagID(ctx context.Context, featureFlagID string) ([]*domain.FeatureFlagRule, error) {
	args := m.Called(ctx, featureFlagID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FeatureFlagRule), args.Error(1)
}

func (m *MockFeatureFlagRuleRepository) DeleteByFeatureFlagID(ctx context.Context, featureFlagID string) error {
	args := m.Called(ctx, featureFlagID)
	return args.Error(0)
}

func (m *MockFeatureFlagRuleRepository) GetRulesByFlagID(ctx context.Context, featureFlagID string) ([]*domain.FeatureFlagRule, error) {
	args := m.Called(ctx, featureFlagID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FeatureFlagRule), args.Error(1)
}

// mockFFCacheService implements cache.Service with mock expectations for feature flag tests
type mockFFCacheService struct {
	mock.Mock
}

func (m *mockFFCacheService) GetEnvironmentConfig(ctx context.Context, sdkKey string) (*cache.LeanEnvironmentConfig, error) {
	args := m.Called(ctx, sdkKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*cache.LeanEnvironmentConfig), args.Error(1)
}

func (m *mockFFCacheService) SetEnvironmentConfig(ctx context.Context, sdkKey string, payload *cache.LeanEnvironmentConfig) error {
	args := m.Called(ctx, sdkKey, payload)
	return args.Error(0)
}

func (m *mockFFCacheService) DeleteEnvironmentConfig(ctx context.Context, sdkKey string) error {
	args := m.Called(ctx, sdkKey)
	return args.Error(0)
}

func (m *mockFFCacheService) GetTenantLevelByID(ctx context.Context, levelID string) (*domain.TenantLevel, error) {
	args := m.Called(ctx, levelID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantLevel), args.Error(1)
}

func (m *mockFFCacheService) GetTenantLevelByCode(ctx context.Context, levelCode string) (*domain.TenantLevel, error) {
	args := m.Called(ctx, levelCode)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.TenantLevel), args.Error(1)
}

func (m *mockFFCacheService) SetTenantLevel(ctx context.Context, level *domain.TenantLevel) error {
	args := m.Called(ctx, level)
	return args.Error(0)
}

func (m *mockFFCacheService) DeleteTenantLevel(ctx context.Context, level *domain.TenantLevel) error {
	args := m.Called(ctx, level)
	return args.Error(0)
}

// mockFFTenantLevelService implements TenantLevelService for testing
type mockFFTenantLevelService struct {
	mock.Mock
}

func (m *mockFFTenantLevelService) CanCreateEnvironment(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *mockFFTenantLevelService) CanCreateFeatureFlag(ctx context.Context, tenantID string) error {
	args := m.Called(ctx, tenantID)
	return args.Error(0)
}

func (m *mockFFTenantLevelService) CanCreateUser(ctx context.Context, tenantID string, role domain.UserRole) error {
	args := m.Called(ctx, tenantID, role)
	return args.Error(0)
}

// mockFFEnvRepo implements EnvironmentRepository for feature flag tests
type mockFFEnvRepo struct {
	mock.Mock
}

func (m *mockFFEnvRepo) Create(ctx context.Context, env *domain.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *mockFFEnvRepo) Update(ctx context.Context, env *domain.Environment) error {
	args := m.Called(ctx, env)
	return args.Error(0)
}

func (m *mockFFEnvRepo) Delete(ctx context.Context, id string, tenantID string) error {
	args := m.Called(ctx, id, tenantID)
	return args.Error(0)
}

func (m *mockFFEnvRepo) ListByTenant(ctx context.Context, tenantID string) ([]*domain.Environment, error) {
	args := m.Called(ctx, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.Environment), args.Error(1)
}

func (m *mockFFEnvRepo) GetByID(ctx context.Context, id string, tenantID string) (*domain.Environment, error) {
	args := m.Called(ctx, id, tenantID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Environment), args.Error(1)
}

func (m *mockFFEnvRepo) CheckNameExistsExcludingID(ctx context.Context, name, tenantID, idToExclude string) (bool, error) {
	args := m.Called(ctx, name, tenantID, idToExclude)
	return args.Bool(0), args.Error(1)
}

func (m *mockFFEnvRepo) FindBySDKKey(ctx context.Context, sdkKey string) (*domain.Environment, error) {
	args := m.Called(ctx, sdkKey)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domain.Environment), args.Error(1)
}

func (m *mockFFEnvRepo) ListAndCount(ctx context.Context, params domain.ListEnvironmentsParams) ([]*domain.Environment, int64, error) {
	args := m.Called(ctx, params)
	if args.Get(0) == nil {
		return nil, 0, args.Error(2)
	}
	return args.Get(0).([]*domain.Environment), args.Get(1).(int64), args.Error(2)
}

func (m *mockFFEnvRepo) CountByTenant(ctx context.Context, tenantID string) (int64, error) {
	args := m.Called(ctx, tenantID)
	return args.Get(0).(int64), args.Error(1)
}

// mockFFHistoryRepo implements FeatureFlagHistoryRepository
type mockFFHistoryRepo struct {
	mock.Mock
}

func (m *mockFFHistoryRepo) Create(ctx context.Context, history *domain.FeatureFlagHistory) error {
	args := m.Called(ctx, history)
	return args.Error(0)
}

func (m *mockFFHistoryRepo) ListByFeatureFlagID(ctx context.Context, flagID string) ([]*domain.FeatureFlagHistory, error) {
	args := m.Called(ctx, flagID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domain.FeatureFlagHistory), args.Error(1)
}

type mockFFSegmentRepo struct {
	mock.Mock
}

func (m *mockFFSegmentRepo) SetSegments(ctx context.Context, flagID string, segmentIDs []string) error {
	args := m.Called(ctx, flagID, segmentIDs)
	return args.Error(0)
}

func (m *mockFFSegmentRepo) GetSegmentIDs(ctx context.Context, flagID string) ([]string, error) {
	args := m.Called(ctx, flagID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]string), args.Error(1)
}

var mockHistoryRepo = new(mockFFHistoryRepo)

func setupServiceWithMocks(t *testing.T) (*featureFlagService, *MockFeatureFlagRepository, *MockFeatureFlagRuleRepository, *MockUserRepository, *mockFFEnvRepo, *mockFFCacheService, *mockFFTenantLevelService, *mockFFSegmentRepo) {
	mockFFRepo := new(MockFeatureFlagRepository)
	mockRuleRepo := new(MockFeatureFlagRuleRepository)
	mockUserRepo := new(MockUserRepository)
	mockEnvRepo := new(mockFFEnvRepo)
	mockCache := new(mockFFCacheService)
	mockTenantLevel := new(mockFFTenantLevelService)
	mockFFSegment := new(mockFFSegmentRepo)

	node, err := snowflake.NewNode(1)
	assert.NoError(t, err)

	svc := NewFeatureFlagService(mockFFRepo, mockRuleRepo, mockHistoryRepo, mockUserRepo, mockEnvRepo, node, mockCache, mockTenantLevel, mockFFSegment).(*featureFlagService)
	return svc, mockFFRepo, mockRuleRepo, mockUserRepo, mockEnvRepo, mockCache, mockTenantLevel, mockFFSegment
}

func TestFeatureFlagService_List(t *testing.T) {
	ctx := context.Background()
	envID := "env-" + uuid.New().String()
	userID := "user-" + uuid.New().String()
	req := request.ListFeatureFlagsRequest{Page: 1, PageSize: 10}
	params := domain.ListFeatureFlagsParams{
		EnvironmentID: envID,
		Keyword:       "",
		Limit:         10,
		Offset:        0,
	}

	mockFlags := []*domain.FeatureFlag{
		{ID: "ff-1", Name: "Flag 1", Enabled: true, UpdatedAt: time.Now(), UpdatedBy: sql.NullString{String: userID, Valid: true}},
	}
	mockUsers := []*domain.User{
		{ID: userID, DisplayName: sql.NullString{String: "Test User", Valid: true}},
	}

	t.Run("Success - list feature flags", func(t *testing.T) {
		svc, mockFFRepo, _, mockUserRepo, _, _, _, _ := setupServiceWithMocks(t)

		mockFFRepo.On("ListAndCountByEnvironment", ctx, params).Return(mockFlags, int64(1), nil).Once()
		mockUserRepo.On("ListByIDs", ctx, []string{userID}).Return(mockUsers, nil).Once()

		result, err := svc.List(ctx, envID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(1), result.Total)
		assert.Len(t, result.Items, 1)
		assert.Equal(t, "Flag 1", result.Items[0].Name)
		assert.Equal(t, "Test User", result.Items[0].UpdatedBy)

		mockFFRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Success - empty list", func(t *testing.T) {
		svc, mockFFRepo, _, _, _, _, _, _ := setupServiceWithMocks(t)

		mockFFRepo.On("ListAndCountByEnvironment", ctx, params).Return([]*domain.FeatureFlag{}, int64(0), nil).Once()

		result, err := svc.List(ctx, envID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, int64(0), result.Total)
		assert.Len(t, result.Items, 0)

		mockFFRepo.AssertExpectations(t)
	})

	t.Run("Failure - repository error on ListAndCount", func(t *testing.T) {
		svc, mockFFRepo, _, _, _, _, _, _ := setupServiceWithMocks(t)
		dbError := errors.New("db error")

		mockFFRepo.On("ListAndCountByEnvironment", ctx, params).Return(nil, int64(0), dbError).Once()

		result, err := svc.List(ctx, envID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to list feature flags from repository")

		mockFFRepo.AssertExpectations(t)
	})
}

func TestFeatureFlagService_Create(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-" + uuid.New().String()
	envID := "env-" + uuid.New().String()
	userID := "user-" + uuid.New().String()
	req := request.CreateFeatureFlagRequest{
		Name:   "New Flag",
		Enabled: true,
		WhitelistRules: []request.WhitelistRule{
			{AccountID: "acc-1", UserIDs: []string{"user-1", "user-2"}},
		},
	}

	t.Run("Success - create feature flag", func(t *testing.T) {
		svc, mockFFRepo, mockRuleRepo, mockUserRepo, mockEnvRepo, mockCache, mockTenantLevel, _ := setupServiceWithMocks(t)

		mockTenantLevel.On("CanCreateFeatureFlag", ctx, tenantID).Return(nil).Once()
		mockFFRepo.On("CheckNameExistsInTenant", ctx, req.Name, tenantID).Return(false, nil).Once()
		mockFFRepo.On("Create", ctx, mock.AnythingOfType("*domain.FeatureFlag")).Return(nil).Once()
		mockRuleRepo.On("Create", ctx, mock.AnythingOfType("*domain.FeatureFlagRule")).Return(nil).Once()
		mockFFRepo.On("GetByID", ctx, mock.AnythingOfType("string"), envID).Return(&domain.FeatureFlag{ID: "ff-123", Name: "New Flag", Enabled: true}, nil).Once()
		mockRuleRepo.On("ListByFeatureFlagID", ctx, mock.AnythingOfType("string")).Return([]*domain.FeatureFlagRule{}, nil).Once()
		mockUserRepo.On("ListByIDs", ctx, mock.Anything).Return([]*domain.User{{ID: userID, DisplayName: sql.NullString{String: "Creator", Valid: true}}}, nil).Once()
		mockEnvRepo.On("GetByID", ctx, envID, tenantID).Return(&domain.Environment{SDKKey: "sdk-test"}, nil).Once()
		mockCache.On("DeleteEnvironmentConfig", ctx, "sdk-test").Return(nil).Once()
		mockHistoryRepo.On("Create", ctx, mock.AnythingOfType("*domain.FeatureFlagHistory")).Return(nil).Once()

		result, err := svc.Create(ctx, tenantID, envID, userID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, req.Name, result.Name)
		assert.Equal(t, "Creator", result.CreatedBy)

		mockFFRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
	})

	t.Run("Failure - name already exists", func(t *testing.T) {
		svc, mockFFRepo, _, _, _, _, mockTenantLevel, _ := setupServiceWithMocks(t)

		mockTenantLevel.On("CanCreateFeatureFlag", ctx, tenantID).Return(nil).Once()
		mockFFRepo.On("CheckNameExistsInTenant", ctx, req.Name, tenantID).Return(true, nil).Once()

		result, err := svc.Create(ctx, tenantID, envID, userID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrFeatureFlagNameExistsInTenant, err)

		mockFFRepo.AssertExpectations(t)
	})

	t.Run("Failure - repository error on Create", func(t *testing.T) {
		svc, mockFFRepo, _, _, _, _, mockTenantLevel, _ := setupServiceWithMocks(t)
		dbError := errors.New("db error")

		mockTenantLevel.On("CanCreateFeatureFlag", ctx, tenantID).Return(nil).Once()
		mockFFRepo.On("CheckNameExistsInTenant", ctx, req.Name, tenantID).Return(false, nil).Once()
		mockFFRepo.On("Create", ctx, mock.AnythingOfType("*domain.FeatureFlag")).Return(dbError).Once()

		result, err := svc.Create(ctx, tenantID, envID, userID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Contains(t, err.Error(), "failed to create feature flag in transaction")

		mockFFRepo.AssertExpectations(t)
	})
}

func TestFeatureFlagService_GetByID(t *testing.T) {
	ctx := context.Background()
	envID := "env-" + uuid.New().String()
	flagID := "ff-" + uuid.New().String()
	creatorID := "user-creator"
	updaterID := "user-updater"

	mockFlag := &domain.FeatureFlag{
		ID:        flagID,
		Name:      "Detailed Flag",
		Enabled:   true,
		CreatedBy: sql.NullString{String: creatorID, Valid: true},
		UpdatedBy: sql.NullString{String: updaterID, Valid: true},
	}
	mockRules := []*domain.FeatureFlagRule{
		{ID: "ffr-1", AccountID: sql.NullString{String: "acc-1", Valid: true}, UserIDs: []string{"user-1"}},
	}
	mockUsers := []*domain.User{
		{ID: creatorID, DisplayName: sql.NullString{String: "Creator User", Valid: true}},
		{ID: updaterID, DisplayName: sql.NullString{String: "Updater User", Valid: true}},
	}

	t.Run("Success - get feature flag detail", func(t *testing.T) {
		svc, mockFFRepo, mockRuleRepo, mockUserRepo, _, _, _, mockFFSegment := setupServiceWithMocks(t)

		mockFFRepo.On("GetByID", ctx, flagID, envID).Return(mockFlag, nil).Once()
		mockRuleRepo.On("ListByFeatureFlagID", ctx, flagID).Return(mockRules, nil).Once()
		mockUserRepo.On("ListByIDs", ctx, mock.Anything).Return(mockUsers, nil).Once()
		mockFFSegment.On("GetSegmentIDs", ctx, flagID).Return([]string{}, nil).Once()

		result, err := svc.GetByID(ctx, envID, flagID)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, "Detailed Flag", result.Name)
		assert.Len(t, result.WhitelistRules, 1)
		assert.Equal(t, "acc-1", result.WhitelistRules[0].AccountID)
		assert.Equal(t, "Creator User", result.CreatedBy)
		assert.Equal(t, "Updater User", result.UpdatedBy)

		mockFFRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
		mockUserRepo.AssertExpectations(t)
		mockFFSegment.AssertExpectations(t)
	})

	t.Run("Failure - flag not found", func(t *testing.T) {
		svc, mockFFRepo, _, _, _, _, _, _ := setupServiceWithMocks(t)

		mockFFRepo.On("GetByID", ctx, flagID, envID).Return(nil, domain.ErrFeatureFlagNotFound).Once()

		result, err := svc.GetByID(ctx, envID, flagID)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrFeatureFlagNotFound, err)

		mockFFRepo.AssertExpectations(t)
	})
}

func TestFeatureFlagService_Update(t *testing.T) {
	ctx := context.Background()
	envID := "env-" + uuid.New().String()
	flagID := "ff-" + uuid.New().String()
	userID := "user-" + uuid.New().String()
	tenantID := "tenant-" + uuid.New().String()
	newDesc := "Updated Description"
	req := request.UpdateFeatureFlagRequest{
		Description: &newDesc,
		Enabled:     true,
		WhitelistRules: []request.WhitelistRule{
			{AccountID: "new-acc-1"},
		},
	}

	mockFlag := &domain.FeatureFlag{ID: flagID, Name: "Original Flag", TenantID: tenantID}

	t.Run("Success - update feature flag", func(t *testing.T) {
		svc, mockFFRepo, mockRuleRepo, _, mockEnvRepo, mockCache, _, mockFFSegment := setupServiceWithMocks(t)

		mockFFRepo.On("GetByID", ctx, flagID, envID).Return(mockFlag, nil).Once()
		mockFFRepo.On("Update", ctx, mock.AnythingOfType("*domain.FeatureFlag")).Return(nil).Once()
		mockRuleRepo.On("DeleteByFeatureFlagID", ctx, flagID).Return(nil).Once()
		mockRuleRepo.On("Create", ctx, mock.AnythingOfType("*domain.FeatureFlagRule")).Return(nil).Once()
		mockFFRepo.On("GetByID", ctx, flagID, envID).Return(&domain.FeatureFlag{
			ID:          flagID,
			Name:        "Original Flag",
			Enabled:     true,
			Description: sql.NullString{String: newDesc, Valid: true},
		}, nil).Twice()
		mockRuleRepo.On("ListByFeatureFlagID", ctx, flagID).Return([]*domain.FeatureFlagRule{}, nil).Once()
		mockEnvRepo.On("GetByID", ctx, envID, tenantID).Return(&domain.Environment{SDKKey: "sdk-test"}, nil).Once()
		mockCache.On("DeleteEnvironmentConfig", ctx, "sdk-test").Return(nil).Once()
		mockHistoryRepo.On("Create", ctx, mock.AnythingOfType("*domain.FeatureFlagHistory")).Return(nil).Once()
		mockFFSegment.On("SetSegments", ctx, flagID, []string(nil)).Return(nil).Once()

		result, err := svc.Update(ctx, envID, userID, flagID, req)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, newDesc, result.Description)
		assert.Len(t, result.WhitelistRules, 1)

		mockFFRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
		mockFFSegment.AssertExpectations(t)
	})

	t.Run("Failure - flag not found on initial get", func(t *testing.T) {
		svc, mockFFRepo, _, _, _, _, _, _ := setupServiceWithMocks(t)

		mockFFRepo.On("GetByID", ctx, flagID, envID).Return(nil, domain.ErrFeatureFlagNotFound).Once()

		result, err := svc.Update(ctx, envID, userID, flagID, req)

		assert.Error(t, err)
		assert.Nil(t, result)
		assert.Equal(t, domain.ErrFeatureFlagNotFound, err)

		mockFFRepo.AssertExpectations(t)
	})
}

func TestFeatureFlagService_Delete(t *testing.T) {
	ctx := context.Background()
	envID := "env-" + uuid.New().String()
	flagID := "ff-" + uuid.New().String()
	userID := "user-" + uuid.New().String()
	tenantID := "tenant-" + uuid.New().String()

	mockFlag := &domain.FeatureFlag{ID: flagID, TenantID: tenantID}

	t.Run("Success - delete feature flag and its rules", func(t *testing.T) {
		svc, mockFFRepo, mockRuleRepo, _, mockEnvRepo, mockCache, _, _ := setupServiceWithMocks(t)

		mockFFRepo.On("GetByID", ctx, flagID, envID).Return(mockFlag, nil).Twice()
		mockEnvRepo.On("GetByID", ctx, envID, tenantID).Return(&domain.Environment{SDKKey: "sdk-test"}, nil).Once()
		mockRuleRepo.On("ListByFeatureFlagID", ctx, flagID).Return([]*domain.FeatureFlagRule{}, nil).Once()
		mockHistoryRepo.On("Create", ctx, mock.AnythingOfType("*domain.FeatureFlagHistory")).Return(nil).Once()
		mockFFRepo.On("Delete", ctx, flagID, envID).Return(nil).Once()
		mockRuleRepo.On("DeleteByFeatureFlagID", ctx, flagID).Return(nil).Once()
		mockCache.On("DeleteEnvironmentConfig", ctx, "sdk-test").Return(nil).Once()

		err := svc.Delete(ctx, envID, flagID, userID)

		assert.NoError(t, err)

		mockFFRepo.AssertExpectations(t)
		mockRuleRepo.AssertExpectations(t)
	})
}