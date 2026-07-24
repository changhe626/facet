package service

import (
	"context"
	"fmt"
	"log/slog"
)

func (s *featureFlagService) getSDKKeyByEnvID(ctx context.Context, envID string, tenantID string) (string, error) {
	env, err := s.envRepo.GetByID(ctx, envID, tenantID)
	if err != nil {
		return "", fmt.Errorf("failed to find environment ID %s for cache invalidation: %w", envID, err)
	}
	return env.SDKKey, nil
}

func (s *featureFlagService) invalidateEnvCache(ctx context.Context, envID, tenantID string) {
	sdkKey, err := s.getSDKKeyByEnvID(ctx, envID, tenantID)
	if err != nil {
		slog.Warn("failed to get sdkKey for cache invalidation", "envID", envID, "error", err)
		return
	}
	if err := s.cacheService.DeleteEnvironmentConfig(ctx, sdkKey); err != nil {
		slog.Warn("environment cache invalidation failed", "sdkKey", sdkKey, "error", err)
	}
}
