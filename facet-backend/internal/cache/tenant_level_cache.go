package cache

import (
	"context"
	"facet-backend/internal/domain"
)

// TenantLevelCacheService defines the cache operation interface for tenant levels.
type TenantLevelCacheService interface {
	GetTenantLevelByID(ctx context.Context, levelID string) (*domain.TenantLevel, error)
	GetTenantLevelByCode(ctx context.Context, levelCode string) (*domain.TenantLevel, error)
	SetTenantLevel(ctx context.Context, level *domain.TenantLevel) error
	DeleteTenantLevel(ctx context.Context, level *domain.TenantLevel) error
}
