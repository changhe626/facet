package service

import (
	"context"
	"database/sql"
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/cache"
	"facet-backend/internal/domain"
	"fmt"
	"log/slog"
	"math"

	"github.com/bwmarrin/snowflake"
	"github.com/google/uuid"
)

// EnvironmentService defines the business logic interface for environments
type EnvironmentService interface {
	List(ctx context.Context, tenantID string, req request.ListEnvironmentsRequest) (*response.ListEnvironmentsResponse, error)
	ListActive(ctx context.Context, tenantID string) ([]*response.ActiveEnvironmentListItem, error)
	Create(ctx context.Context, tenantID, creatorID string, req request.CreateEnvironmentRequest) (*response.CreateEnvironmentResponse, error)
	Update(ctx context.Context, tenantID, envID, updaterID string, req request.UpdateEnvironmentRequest) (*response.UpdateEnvironmentResponse, error)
	Delete(ctx context.Context, tenantID string, envID string) error
}

// environmentService implements the EnvironmentService interface
type environmentService struct {
	envRepo            domain.EnvironmentRepository
	snowflakeNode      *snowflake.Node
	cacheService       cache.Service
	tenantLevelService TenantLevelService
}

func NewEnvironmentService(
	envRepo domain.EnvironmentRepository,
	node *snowflake.Node,
	cacheService cache.Service,
	tenantLevelService TenantLevelService,
) EnvironmentService {
	return &environmentService{
		envRepo:            envRepo,
		snowflakeNode:      node,
		cacheService:       cacheService,
		tenantLevelService: tenantLevelService,
	}
}

// ListActive lists all active environments under a tenant
func (s *environmentService) ListActive(ctx context.Context, tenantID string) ([]*response.ActiveEnvironmentListItem, error) {
	envs, err := s.envRepo.ListByTenant(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to list active environments from repository: %w", err)
	}

	activeEnvs := make([]*response.ActiveEnvironmentListItem, 0)
	for _, env := range envs {
		if env.IsActive {
			activeEnvs = append(activeEnvs, &response.ActiveEnvironmentListItem{
				ID:   env.ID,
				Name: env.Name,
			})
		}
	}

	return activeEnvs, nil
}

// List retrieves a paginated list of environments (with search and filter support)
func (s *environmentService) List(ctx context.Context, tenantID string, req request.ListEnvironmentsRequest) (*response.ListEnvironmentsResponse, error) {
	offset := (req.Page - 1) * req.PageSize
	params := domain.ListEnvironmentsParams{
		TenantID: tenantID,
		Keyword:  req.Keyword,
		IsActive: req.IsActive,
		Limit:    req.PageSize,
		Offset:   offset,
	}

	envs, total, err := s.envRepo.ListAndCount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list environments from repository: %w", err)
	}

	items := make([]response.EnvironmentListItem, len(envs))
	for i, env := range envs {
		items[i] = response.EnvironmentListItem{
			ID:        env.ID,
			Name:      env.Name,
			SDKKey:    env.SDKKey,
			IsActive:  env.IsActive,
			CreatedAt: env.CreatedAt,
			UpdatedAt: env.UpdatedAt,
		}
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(req.PageSize)))
	}

	return &response.ListEnvironmentsResponse{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

func (s *environmentService) Create(ctx context.Context, tenantID, creatorID string, req request.CreateEnvironmentRequest) (*response.CreateEnvironmentResponse, error) {
	if err := s.tenantLevelService.CanCreateEnvironment(ctx, tenantID); err != nil {
		return nil, err
	}

	exists, err := s.envRepo.CheckNameExistsExcludingID(ctx, req.Name, tenantID, "")
	if err != nil {
		return nil, fmt.Errorf("failed to check environment name existence: %w", err)
	}
	if exists {
		return nil, domain.ErrEnvironmentNameExists
	}

	newEnv := &domain.Environment{
		ID:          "env-" + s.snowflakeNode.Generate().String(),
		TenantID:    tenantID,
		Name:        req.Name,
		Description: sql.NullString{String: req.Description, Valid: req.Description != ""},
		SDKKey:      "sdk-" + uuid.New().String(),
		IsActive:    true, // active by default
		CreatedBy:   sql.NullString{String: creatorID, Valid: true},
		UpdatedBy:   sql.NullString{String: creatorID, Valid: true},
	}

	if err := s.envRepo.Create(ctx, newEnv); err != nil {
		return nil, fmt.Errorf("failed to create environment in repository: %w", err)
	}

	return &response.CreateEnvironmentResponse{
		ID:       newEnv.ID,
		Name:     newEnv.Name,
		SDKKey:   newEnv.SDKKey,
		IsActive: newEnv.IsActive,
	}, nil
}

// Update updates an existing environment and invalidates cache
func (s *environmentService) Update(ctx context.Context, tenantID, envID, updaterID string, req request.UpdateEnvironmentRequest) (*response.UpdateEnvironmentResponse, error) {
	exists, err := s.envRepo.CheckNameExistsExcludingID(ctx, req.Name, tenantID, envID)
	if err != nil {
		return nil, fmt.Errorf("failed to check environment name existence: %w", err)
	}
	if exists {
		return nil, domain.ErrEnvironmentNameExists
	}

	env, err := s.envRepo.GetByID(ctx, envID, tenantID)
	if err != nil {
		return nil, err
	}

	// Cache invalidation requires the SDK Key, so save it before updating
	sdkKey := env.SDKKey

	env.Name = req.Name
	env.Description = sql.NullString{String: req.Description, Valid: req.Description != ""}
	if req.IsActive != nil {
		env.IsActive = *req.IsActive
	}
	env.UpdatedBy = sql.NullString{String: updaterID, Valid: true}

	if err := s.envRepo.Update(ctx, env); err != nil {
		return nil, fmt.Errorf("failed to update environment in repository: %w", err)
	}

	// Invalidate cache
	if err := s.cacheService.DeleteEnvironmentConfig(ctx, sdkKey); err != nil {
		// Log the error but don't let the request fail, since the main operation succeeded
		slog.Warn("environment cache invalidation failed", "sdkKey", sdkKey, "error", err)
	}

	return &response.UpdateEnvironmentResponse{
		ID:        env.ID,
		Name:      env.Name,
		SDKKey:    env.SDKKey,
		IsActive:  env.IsActive,
		UpdatedAt: env.UpdatedAt,
	}, nil
}

func (s *environmentService) Delete(ctx context.Context, tenantID string, envID string) error {
	env, err := s.envRepo.GetByID(ctx, envID, tenantID)
	if err != nil {
		return err
	}

	if err := s.envRepo.Delete(ctx, envID, tenantID); err != nil {
		if errors.Is(err, domain.ErrEnvironmentNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete environment in repository: %w", err)
	}

	if err := s.cacheService.DeleteEnvironmentConfig(ctx, env.SDKKey); err != nil {
		slog.Warn("environment cache invalidation failed", "sdkKey", env.SDKKey, "error", err)
	}

	return nil
}
