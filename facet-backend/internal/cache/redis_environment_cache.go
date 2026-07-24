package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/redis/go-redis/v9"
)

// generateEnvironmentCacheKey creates a standard Redis key for environment configuration.
func (s *redisCacheService) generateEnvironmentCacheKey(sdkKey string) string {
	return fmt.Sprintf("env_config:%s", sdkKey)
}

// GetEnvironmentConfig implements the logic to retrieve and deserialize the optimized environment configuration from Redis.
func (s *redisCacheService) GetEnvironmentConfig(ctx context.Context, sdkKey string) (*LeanEnvironmentConfig, error) {
	cacheKey := s.generateEnvironmentCacheKey(sdkKey)
	val, err := s.redisClient.Get(ctx, cacheKey).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil // Cache miss, this is not an error.
		}
		return nil, fmt.Errorf("failed to get environment config from Redis: %w", err)
	}

	var payload LeanEnvironmentConfig
	if err := json.Unmarshal([]byte(val), &payload); err != nil {
		return nil, fmt.Errorf("failed to parse cached environment config: %w", err)
	}

	return &payload, nil
}

// SetEnvironmentConfig implements the logic to serialize the optimized environment configuration and store it in Redis.
func (s *redisCacheService) SetEnvironmentConfig(ctx context.Context, sdkKey string, payload *LeanEnvironmentConfig) error {
	cacheKey := s.generateEnvironmentCacheKey(sdkKey)
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to serialize optimized environment config: %w", err)
	}

	if err := s.redisClient.Set(ctx, cacheKey, payloadBytes, cacheTTL).Err(); err != nil {
		return fmt.Errorf("failed to set environment config cache in Redis: %w", err)
	}

	return nil
}

// DeleteEnvironmentConfig implements the logic to delete the environment configuration from Redis.
func (s *redisCacheService) DeleteEnvironmentConfig(ctx context.Context, sdkKey string) error {
	cacheKey := s.generateEnvironmentCacheKey(sdkKey)
	if err := s.redisClient.Del(ctx, cacheKey).Err(); err != nil {
		return fmt.Errorf("failed to invalidate environment config cache for sdkKey %s: %w", sdkKey, err)
	}
	return nil
}
