package service

import (
	"context"
	"database/sql"
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/cache"
	"facet-backend/internal/domain"
	"facet-backend/pkg/password"
	"fmt"
	"log/slog"
	"math"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TenantService defines the business logic interface for tenants
type TenantService interface {
	CreateTenantWithAdmin(ctx context.Context, creatorID string, req request.CreateTenantRequest) error
	Update(ctx context.Context, tenantID string, updaterID string, req request.UpdateTenantRequest) error
	List(ctx context.Context, req request.ListTenantsRequest) (*response.ListTenantsResponse, error)
	ResetTenantAdminPassword(ctx context.Context, tenantID, newPassword string) error
	GetByID(ctx context.Context, tenantID string) (*response.Tenant, error)
}

// tenantService implements the TenantService interface
type tenantService struct {
	pool          *pgxpool.Pool
	tenantRepo    domain.TenantRepository
	userRepo      domain.UserRepository
	levelRepo     domain.TenantLevelRepository
	cacheService  cache.Service
	snowflakeNode *snowflake.Node
}

func NewTenantService(
	pool *pgxpool.Pool,
	tenantRepo domain.TenantRepository,
	userRepo domain.UserRepository,
	levelRepo domain.TenantLevelRepository,
	cacheService cache.Service,
	node *snowflake.Node,
) TenantService {
	return &tenantService{
		pool:          pool,
		tenantRepo:    tenantRepo,
		userRepo:      userRepo,
		levelRepo:     levelRepo,
		cacheService:  cacheService,
		snowflakeNode: node,
	}
}

func (s *tenantService) CreateTenantWithAdmin(ctx context.Context, creatorID string, req request.CreateTenantRequest) error {
	exists, err := s.tenantRepo.CheckNameExistsExcludingID(ctx, req.TenantInfo.Name, "")
	if err != nil {
		return fmt.Errorf("failed to check tenant name existence: %w", err)
	}
	if exists {
		return domain.ErrTenantNameExists
	}

	var levelID string
	if req.TenantInfo.LevelID != "" {
		level, err := s.cacheService.GetTenantLevelByID(ctx, req.TenantInfo.LevelID)
		if err != nil {
			slog.Warn("failed to get tenant level from cache", "error", err)
		}
		if level == nil {
			level, err = s.levelRepo.GetByID(ctx, req.TenantInfo.LevelID)
			if err != nil {
				return fmt.Errorf("provided tenant level ID is invalid: %w", err)
			}
			if err := s.cacheService.SetTenantLevel(ctx, level); err != nil {
			slog.Warn("failed to backfill tenant level cache", "error", err)
		}
	}
	levelID = level.ID
} else {
	freeLevel, err := s.cacheService.GetTenantLevelByCode(ctx, "free")
	if err != nil {
		slog.Warn("failed to get 'free' level from cache", "error", err)
		}
		if freeLevel == nil {
			freeLevel, err = s.levelRepo.GetByCode(ctx, "free")
			if err != nil {
				return fmt.Errorf("failed to get default 'free' tenant level from DB, please ensure it exists: %w", err)
			}
			if err := s.cacheService.SetTenantLevel(ctx, freeLevel); err != nil {
				slog.Warn("failed to backfill tenant level cache", "error", err)
			}
		}
		levelID = freeLevel.ID
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	newTenantID := "tenant-" + s.snowflakeNode.Generate().String()
	{
		query := `
			INSERT INTO tenants (id, level_id, name, description, website, status, created_by, updated_by)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		`
		status := req.TenantInfo.Status
		if status == "" {
			status = string(domain.TenantStatusActive)
		}
		_, err := tx.Exec(ctx, query,
			newTenantID,
			levelID,
			req.TenantInfo.Name,
			sql.NullString{String: req.TenantInfo.Description, Valid: req.TenantInfo.Description != ""},
			sql.NullString{String: req.TenantInfo.Website, Valid: req.TenantInfo.Website != ""},
			domain.TenantStatus(status),
			sql.NullString{String: creatorID, Valid: true},
			sql.NullString{String: creatorID, Valid: true},
		)
		if err != nil {
			return fmt.Errorf("failed to create tenant in transaction: %w", err)
		}
	}

	hashedPassword, err := password.HashPassword(req.UserInfo.AdminInitialPassword)
	if err != nil {
		return fmt.Errorf("failed to hash admin password: %w", err)
	}

	{
		query := `
			INSERT INTO users (id, tenant_id, role, email, password_hash, display_name, created_by, updated_by, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		`
		_, err := tx.Exec(ctx, query,
			"user-"+s.snowflakeNode.Generate().String(),
			sql.NullString{String: newTenantID, Valid: true},
			domain.UserRole(req.UserInfo.Role),
			req.UserInfo.AdminEmail,
			hashedPassword,
			sql.NullString{String: req.UserInfo.AdminDisplayName, Valid: true},
			sql.NullString{String: creatorID, Valid: true},
			sql.NullString{String: creatorID, Valid: true},
		)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				return domain.ErrEmailExistsInTenant
			}
			return fmt.Errorf("failed to create admin user in transaction: %w", err)
		}
	}

	return tx.Commit(ctx)
}

func (s *tenantService) Update(ctx context.Context, tenantID string, updaterID string, req request.UpdateTenantRequest) error {
	exists, err := s.tenantRepo.CheckNameExistsExcludingID(ctx, req.Name, tenantID)
	if err != nil {
		return fmt.Errorf("failed to check tenant name existence: %w", err)
	}
	if exists {
		return domain.ErrTenantNameExists
	}

	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return err
	}

	tenant.Name = req.Name
	if req.LevelID != "" {
		// Validate that the provided level ID actually exists
		level, err := s.cacheService.GetTenantLevelByID(ctx, req.LevelID)
		if err != nil {
			slog.Warn("failed to get tenant level from cache", "error", err)
		}
		if level == nil {
			level, err = s.levelRepo.GetByID(ctx, req.LevelID)
			if err != nil {
				return fmt.Errorf("invalid tenant level ID: %w", domain.ErrTenantLevelNotFound)
			}
			if err := s.cacheService.SetTenantLevel(ctx, level); err != nil {
				slog.Warn("failed to backfill tenant level cache", "error", err)
			}
		}
		tenant.LevelID = level.ID
	}
	tenant.Description = sql.NullString{String: req.Description, Valid: req.Description != ""}
	tenant.Website = sql.NullString{String: req.Website, Valid: req.Website != ""}
	tenant.Status = domain.TenantStatus(req.Status)
	tenant.UpdatedBy = sql.NullString{String: updaterID, Valid: true}

	if err := s.tenantRepo.Update(ctx, tenant); err != nil {
		return fmt.Errorf("failed to update tenant in repository: %w", err)
	}

	return nil
}

func (s *tenantService) List(ctx context.Context, req request.ListTenantsRequest) (*response.ListTenantsResponse, error) {
	offset := (req.Page - 1) * req.PageSize
	params := domain.ListTenantsParams{
		Name:   req.Name,
		Limit:  req.PageSize,
		Offset: offset,
	}

	tenants, total, err := s.tenantRepo.ListAndCount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list tenants from repository: %w", err)
	}

	items := make([]response.TenantListItem, len(tenants))
	for i, tenant := range tenants {
		items[i] = response.TenantListItem{
			ID:        tenant.ID,
			Name:      tenant.Name,
			LevelID:   tenant.LevelID,
			Status:    string(tenant.Status),
			CreatedAt: tenant.CreatedAt,
		}
	}

	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(req.PageSize)))
	}

	return &response.ListTenantsResponse{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

func (s *tenantService) ResetTenantAdminPassword(ctx context.Context, tenantID, newPassword string) error {
	admin, err := s.userRepo.FindByTenantIDAndRole(ctx, tenantID, domain.RoleTenantSuperAdmin)
	if err != nil {
		return err
	}

	newPasswordHash, err := password.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if !admin.TenantID.Valid {
		return fmt.Errorf("cannot reset password for a user without a tenant")
	}

	if err := s.userRepo.UpdatePassword(ctx, admin.ID, newPasswordHash); err != nil {
		return fmt.Errorf("failed to update admin password in repository: %w", err)
	}

	return nil
}

func (s *tenantService) GetByID(ctx context.Context, tenantID string) (*response.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		// Could be domain.ErrTenantNotFound, which will be handled by the error middleware
		return nil, fmt.Errorf("failed to get tenant from repository: %w", err)
	}

	// Look up the tenant level name
	var levelName string
	level, err := s.levelRepo.GetByID(ctx, tenant.LevelID)
	if err == nil && level != nil {
		levelName = level.Name
	}

	// Convert domain.Tenant to response.Tenant
	resp := &response.Tenant{
		ID:          tenant.ID,
		Name:        tenant.Name,
		Description: tenant.Description.String,
		Website:     tenant.Website.String,
		LevelID:     tenant.LevelID,
		LevelName:   levelName,
		Status:      string(tenant.Status),
		CreatedAt:   tenant.CreatedAt,
	}

	return resp, nil
}
