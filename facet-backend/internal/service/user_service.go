package service

import (
	"context"
	"database/sql"
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"facet-backend/pkg/password"
	"fmt"
	"math"
	"time"

	"github.com/bwmarrin/snowflake"
	"github.com/jackc/pgx/v5/pgconn"
)

// UserService defines the business logic interface for users
type UserService interface {
	ListUsers(ctx context.Context, tenantID, userID string, role domain.UserRole, req request.ListUsersRequest) (*response.ListUsersResponse, error)
	Create(ctx context.Context, tenantID, userID string, req request.CreateUserRequest) (*response.CreateUserResponse, error)
	Delete(ctx context.Context, userID string) error
	ResetPassword(ctx context.Context, userID, newPassword string) error
	GetByID(ctx context.Context, tenantID, userID string) (*response.UserDetailResponse, error)
	Update(ctx context.Context, tenantID, userID, updaterID string, req request.UpdateUserRequest) error
}

// userService implements the UserService interface
type userService struct {
	userRepo           domain.UserRepository
	snowflakeNode      *snowflake.Node
	tenantLevelService TenantLevelService
}

func NewUserService(
	userRepo domain.UserRepository,
	node *snowflake.Node,
	tenantLevelService TenantLevelService,
) UserService {
	return &userService{
		userRepo:           userRepo,
		snowflakeNode:      node,
		tenantLevelService: tenantLevelService,
	}
}

func (s *userService) ListUsers(ctx context.Context, tenantID, userID string, role domain.UserRole, req request.ListUsersRequest) (*response.ListUsersResponse, error) {
	offset := (req.Page - 1) * req.PageSize
	params := domain.ListUsersParams{
		Keyword: req.Keyword,
		Limit:   req.PageSize,
		Offset:  offset,
	}
	switch role {
	case domain.RoleSystemAdmin:
	case domain.RoleTenantSuperAdmin, domain.RoleTenantAdmin:
		params.TenantID = tenantID
	case domain.RoleTenantViewer:
		params.UserID = userID
	default:
		return &response.ListUsersResponse{
			Page:       req.Page,
			PageSize:   req.PageSize,
			Total:      0,
			TotalPages: 0,
			Items:      []response.UserListItem{},
		}, nil
	}
	users, total, err := s.userRepo.ListAndCount(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to list users from repository: %w", err)
	}
	items := make([]response.UserListItem, len(users))
	for i, user := range users {
		items[i] = response.UserListItem{
			ID:          user.ID,
			DisplayName: user.DisplayName.String,
			Email:       user.Email,
			Role:        string(user.Role),
			CreatedAt:   user.CreatedAt,
		}
		if user.LastLoginAt.Valid {
			items[i].LastLoginAt = &user.LastLoginAt.Time
		}
	}
	totalPages := 0
	if total > 0 {
		totalPages = int(math.Ceil(float64(total) / float64(req.PageSize)))
	}
	return &response.ListUsersResponse{
		Page:       req.Page,
		PageSize:   req.PageSize,
		Total:      total,
		TotalPages: totalPages,
		Items:      items,
	}, nil
}

// Create creates a new user with quota check
func (s *userService) Create(ctx context.Context, tenantID, userId string, req request.CreateUserRequest) (*response.CreateUserResponse, error) {
	// --- Quota check ---
	if err := s.tenantLevelService.CanCreateUser(ctx, tenantID, domain.UserRole(req.Role)); err != nil {
		return nil, err // return error directly if quota exceeded or check fails
	}

	exists, err := s.userRepo.CheckEmailExistsInTenant(ctx, req.Email, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return nil, domain.ErrEmailExistsInTenant
	}

	hashedPassword, err := password.HashPassword(req.InitialPassword)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now()
	newUser := &domain.User{
		ID:           "user-" + s.snowflakeNode.Generate().String(),
		TenantID:     sql.NullString{String: tenantID, Valid: tenantID != ""},
		Role:         domain.UserRole(req.Role),
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  sql.NullString{String: req.DisplayName, Valid: req.DisplayName != ""},
		CreatedBy:    sql.NullString{String: userId, Valid: userId != ""},
		UpdatedBy:    sql.NullString{String: userId, Valid: userId != ""},
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := s.userRepo.Create(ctx, newUser); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique_violation
			return nil, domain.ErrEmailExistsInTenant
		}
		return nil, fmt.Errorf("failed to create user in repository: %w", err)
	}

	return &response.CreateUserResponse{
		ID:          newUser.ID,
		Email:       newUser.Email,
		DisplayName: newUser.DisplayName.String,
		Role:        string(newUser.Role),
	}, nil
}

func (s *userService) Delete(ctx context.Context, userID string) error {
	if err := s.userRepo.Delete(ctx, userID); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("failed to delete user in repository: %w", err)
	}
	return nil
}

func (s *userService) ResetPassword(ctx context.Context, userID, newPassword string) error {
	newPasswordHash, err := password.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash new password: %w", err)
	}

	if err := s.userRepo.UpdatePassword(ctx, userID, newPasswordHash); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return err
		}
		return fmt.Errorf("failed to update password in repository: %w", err)
	}

	return nil
}

// GetByID retrieves user details
func (s *userService) GetByID(ctx context.Context, tenantID, userID string) (*response.UserDetailResponse, error) {
	user, err := s.userRepo.GetByID(ctx, userID, tenantID)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, err
		}
		return nil, fmt.Errorf("failed to get user by id from repository: %w", err)
	}

	return &response.UserDetailResponse{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName.String,
		Role:        string(user.Role),
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
		CreatedBy:   user.CreatedBy.String,
		UpdatedBy:   user.UpdatedBy.String,
		LastLoginAt: user.LastLoginAt.Time,
	}, nil
}

func (s *userService) Update(ctx context.Context, tenantID, userID, updaterID string, req request.UpdateUserRequest) error {
	exists, err := s.userRepo.CheckEmailExistsInTenantExcludingUser(ctx, req.Email, tenantID, userID)
	if err != nil {
		return fmt.Errorf("failed to check email existence: %w", err)
	}
	if exists {
		return domain.ErrEmailExistsInTenant
	}

	userToUpdate, err := s.userRepo.GetByID(ctx, userID, tenantID)
	if err != nil {
		return err
	}

	userToUpdate.Email = req.Email
	userToUpdate.DisplayName = sql.NullString{String: req.DisplayName, Valid: req.DisplayName != ""}
	userToUpdate.Role = domain.UserRole(req.Role)
	if req.Role == "system_admin" {
		return domain.ErrRoleInTenant
	}
	userToUpdate.UpdatedBy = sql.NullString{String: updaterID, Valid: true}

	if err := s.userRepo.Update(ctx, userToUpdate); err != nil {
		return fmt.Errorf("failed to update user in repository: %w", err)
	}

	return nil
}
