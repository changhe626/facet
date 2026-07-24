package service

import (
	"context"
	"errors"
	"facet-backend/internal/cache"
	"facet-backend/internal/domain"
	"fmt"
	"log/slog"
)

var (
	ErrQuotaExceededEnvironments  = errors.New("maximum number of environments reached for the current plan")
	ErrQuotaExceededFeatureFlags  = errors.New("maximum number of feature flags reached for the current plan")
	ErrQuotaExceededTenantAdmins  = errors.New("maximum number of admin users reached for the current plan")
	ErrQuotaExceededTenantViewers = errors.New("maximum number of viewer users reached for the current plan")
)

// TenantLevelService defines the business logic interface for tenant levels, quotas, and permissions
type TenantLevelService interface {
	// CanCreateEnvironment checks whether the specified tenant can create a new environment
	CanCreateEnvironment(ctx context.Context, tenantID string) error

	// CanCreateFeatureFlag checks whether the specified tenant can create a new feature flag
	CanCreateFeatureFlag(ctx context.Context, tenantID string) error

	// CanCreateUser checks whether the specified tenant can create a new user with a given role
	CanCreateUser(ctx context.Context, tenantID string, role domain.UserRole) error
}

// tenantLevelService implements the TenantLevelService interface
type tenantLevelService struct {
	levelRepo    domain.TenantLevelRepository
	tenantRepo   domain.TenantRepository
	envRepo      domain.EnvironmentRepository
	ffRepo       domain.FeatureFlagRepository
	userRepo     domain.UserRepository
	cacheService cache.Service // <-- inject cache service
}

// NewTenantLevelService creates a new TenantLevelService instance
func NewTenantLevelService(
	levelRepo domain.TenantLevelRepository,
	tenantRepo domain.TenantRepository,
	envRepo domain.EnvironmentRepository,
	ffRepo domain.FeatureFlagRepository,
	userRepo domain.UserRepository,
	cacheService cache.Service, // <-- inject cache service
) TenantLevelService {
	return &tenantLevelService{
		levelRepo:    levelRepo,
		tenantRepo:   tenantRepo,
		envRepo:      envRepo,
		ffRepo:       ffRepo,
		userRepo:     userRepo,
		cacheService: cacheService,
	}
}

// getTenantAndLevel is a helper function to get the tenant and its level information, using cache
func (s *tenantLevelService) getTenantAndLevel(ctx context.Context, tenantID string) (*domain.Tenant, *domain.TenantLevel, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	// 1. Try to get level info from cache first
	level, err := s.cacheService.GetTenantLevelByID(ctx, tenant.LevelID)
	if err != nil {
		// If cache fails, log the error but continue loading from DB
		slog.Warn("failed to get tenant level from cache", "error", err)
	}
	if level != nil {
		return tenant, level, nil // cache hit
	}

	// 2. Cache miss, load from DB
	level, err = s.levelRepo.GetByID(ctx, tenant.LevelID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get tenant level from DB for tenant %s: %w", tenantID, err)
	}

	// 3. Backfill the result from DB into cache
	if err := s.cacheService.SetTenantLevel(ctx, level); err != nil {
		slog.Warn("failed to backfill tenant level cache", "error", err)
	}

	return tenant, level, nil
}

// CanCreateEnvironment checks environment quota
func (s *tenantLevelService) CanCreateEnvironment(ctx context.Context, tenantID string) error {
	_, level, err := s.getTenantAndLevel(ctx, tenantID)
	if err != nil {
		return err
	}

	currentCount, err := s.envRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to count environments for tenant %s: %w", tenantID, err)
	}

	if currentCount >= int64(level.MaxEnvironments) {
		return ErrQuotaExceededEnvironments
	}

	return nil
}

// CanCreateFeatureFlag checks feature flag quota
func (s *tenantLevelService) CanCreateFeatureFlag(ctx context.Context, tenantID string) error {
	_, level, err := s.getTenantAndLevel(ctx, tenantID)
	if err != nil {
		return err
	}

	currentCount, err := s.ffRepo.CountByTenant(ctx, tenantID)
	if err != nil {
		return fmt.Errorf("failed to count feature flags for tenant %s: %w", tenantID, err)
	}

	if currentCount >= int64(level.MaxFeatureFlags) {
		return ErrQuotaExceededFeatureFlags
	}

	return nil
}

// CanCreateUser checks user quota
func (s *tenantLevelService) CanCreateUser(ctx context.Context, tenantID string, role domain.UserRole) error {
	_, level, err := s.getTenantAndLevel(ctx, tenantID)
	if err != nil {
		return err
	}

	currentCount, err := s.userRepo.CountByTenantAndRole(ctx, tenantID, role)
	if err != nil {
		return fmt.Errorf("failed to count users for tenant %s with role %s: %w", tenantID, role, err)
	}

	switch role {
	case domain.RoleTenantAdmin:
		if currentCount >= int64(level.MaxTenantAdmins) {
			return ErrQuotaExceededTenantAdmins
		}
	case domain.RoleTenantViewer:
		if currentCount >= int64(level.MaxTenantViewers) {
			return ErrQuotaExceededTenantViewers
		}
	}

	return nil
}
