package service

import (
	"context"
	"errors"
	"facet-backend/internal/api/request"
	"facet-backend/internal/api/response"
	"facet-backend/internal/domain"
	"facet-backend/pkg/password"
	"facet-backend/pkg/token"
	"fmt"
)

var ErrInvalidCredentials = errors.New("user not found or password incorrect")

type AuthService interface {
	Login(ctx context.Context, req request.LoginRequest) (*response.LoginResponse, error)
}

type authService struct {
	userRepo  domain.UserRepository
	secretKey string
	expiresIn int
}

func NewAuthService(userRepo domain.UserRepository, secretKey string, expiresIn int) AuthService {
	return &authService{
		userRepo:  userRepo,
		secretKey: secretKey,
		expiresIn: expiresIn,
	}
}

func (s *authService) Login(ctx context.Context, req request.LoginRequest) (*response.LoginResponse, error) {
	user, err := s.userRepo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	if err := password.CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, ErrInvalidCredentials
	}

	jwtToken, err := token.GenerateToken(user.ID, user.TenantID.String, string(user.Role), s.secretKey, s.expiresIn)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &response.LoginResponse{
		Token: jwtToken,
		User: response.UserInfo{
			ID:          user.ID,
			DisplayName: user.DisplayName.String,
			Role:        string(user.Role),
		},
	}, nil
}
