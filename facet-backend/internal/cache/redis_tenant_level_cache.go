package cache

import (
	"context"
	"encoding/json"
	"errors"
	"facet-backend/internal/domain"
	"fmt"

	"github.com/redis/go-redis/v9"
)

func (s *redisCacheService) generateTenantLevelIDKey(levelID string) string {
	return fmt.Sprintf("tenant_level:id:%s", levelID)
}

func (s *redisCacheService) generateTenantLevelCodeKey(levelCode string) string {
	return fmt.Sprintf("tenant_level:code:%s", levelCode)
}

func (s *redisCacheService) getTenantLevel(ctx context.Context, key string) (*domain.TenantLevel, error) {
	val, err := s.redisClient.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get tenant level from Redis: %w", err)
	}

	var level domain.TenantLevel
	if err := json.Unmarshal([]byte(val), &level); err != nil {
		return nil, fmt.Errorf("failed to parse cached tenant level: %w", err)
	}

	return &level, nil
}

func (s *redisCacheService) GetTenantLevelByID(ctx context.Context, levelID string) (*domain.TenantLevel, error) {
	return s.getTenantLevel(ctx, s.generateTenantLevelIDKey(levelID))
}

func (s *redisCacheService) GetTenantLevelByCode(ctx context.Context, levelCode string) (*domain.TenantLevel, error) {
	return s.getTenantLevel(ctx, s.generateTenantLevelCodeKey(levelCode))
}

func (s *redisCacheService) SetTenantLevel(ctx context.Context, level *domain.TenantLevel) error {
	payloadBytes, err := json.Marshal(level)
	if err != nil {
		return fmt.Errorf("failed to serialize tenant level: %w", err)
	}

	// Cache both ID and Code keys simultaneously
	idKey := s.generateTenantLevelIDKey(level.ID)
	if err := s.redisClient.Set(ctx, idKey, payloadBytes, cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set tenant level cache by ID in Redis: %w", err)
	}

	codeKey := s.generateTenantLevelCodeKey(level.Code)
	if err := s.redisClient.Set(ctx, codeKey, payloadBytes, cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set tenant level cache by Code in Redis: %w", err)
	}

	return nil
}

func (s *redisCacheService) DeleteTenantLevel(ctx context.Context, level *domain.TenantLevel) error {
	// Delete both the ID and Code keys simultaneously
	idKey := s.generateTenantLevelIDKey(level.ID)
	codeKey := s.generateTenantLevelCodeKey(level.Code)

	if err := s.redisClient.Del(ctx, idKey, codeKey).Err(); err != nil {
		return fmt.Errorf("failed to invalidate tenant level cache: %w", err)
	}

	return nil
}
