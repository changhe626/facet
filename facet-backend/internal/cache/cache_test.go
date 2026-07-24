package cache

import (
	"context"
	"facet-backend/internal/domain"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestRedis creates a miniredis server and returns a redisCacheService connected to it.
func setupTestRedis(t *testing.T) (*redisCacheService, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err, "failed to start miniredis")

	redisClient := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	svc := NewRedisCacheService(redisClient).(*redisCacheService)
	return svc, mr
}

func TestEnvironmentCacheService_SetAndGet(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()
	ctx := context.Background()

	sdkKey := "sdk-test-key-123"
	payload := &LeanEnvironmentConfig{
		IsActive:     true,
		TenantID:     "tenant-abc",
		TenantStatus: domain.TenantStatusActive,
		FeatureFlags: map[string]LeanFeatureFlag{
			"flag-1": {
				DefaultStatus:     true,
				RolloutPercentage: 50,
				Rules: []LeanRule{
					{AccountID: "acc-1", UserIDs: []string{"user-1", "user-2"}},
				},
			},
			"flag-2": {
				DefaultStatus:     false,
				RolloutPercentage: 0,
				Segments: []LeanSegment{
					{
						MatchType: "ANY",
						Conditions: []LeanSegmentCondition{
							{AttributeKey: "country", Operator: "eq", Value: "US"},
						},
					},
				},
			},
		},
	}

	t.Run("Set and Get - success", func(t *testing.T) {
		err := svc.SetEnvironmentConfig(ctx, sdkKey, payload)
		require.NoError(t, err)

		result, err := svc.GetEnvironmentConfig(ctx, sdkKey)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, payload.IsActive, result.IsActive)
		assert.Equal(t, payload.TenantID, result.TenantID)
		assert.Equal(t, payload.TenantStatus, result.TenantStatus)
		assert.Len(t, result.FeatureFlags, 2)

		flag1, ok := result.FeatureFlags["flag-1"]
		assert.True(t, ok)
		assert.True(t, flag1.DefaultStatus)
		assert.Equal(t, 50, flag1.RolloutPercentage)
		assert.Len(t, flag1.Rules, 1)
		assert.Equal(t, "acc-1", flag1.Rules[0].AccountID)
		assert.ElementsMatch(t, []string{"user-1", "user-2"}, flag1.Rules[0].UserIDs)

		flag2, ok := result.FeatureFlags["flag-2"]
		assert.True(t, ok)
		assert.False(t, flag2.DefaultStatus)
		assert.Len(t, flag2.Segments, 1)
		assert.Equal(t, "ANY", flag2.Segments[0].MatchType)
		assert.Len(t, flag2.Segments[0].Conditions, 1)
		assert.Equal(t, "country", flag2.Segments[0].Conditions[0].AttributeKey)
	})

	t.Run("Get - cache miss returns nil, nil", func(t *testing.T) {
		result, err := svc.GetEnvironmentConfig(ctx, "nonexistent-key")
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Get - different SDK key does not collide", func(t *testing.T) {
		result, err := svc.GetEnvironmentConfig(ctx, "another-key")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestEnvironmentCacheService_Delete(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()
	ctx := context.Background()

	sdkKey := "sdk-key-to-delete"
	payload := &LeanEnvironmentConfig{
		IsActive:     true,
		TenantID:     "tenant-xyz",
		TenantStatus: domain.TenantStatusActive,
		FeatureFlags: map[string]LeanFeatureFlag{},
	}

	t.Run("Delete - removes cached entry", func(t *testing.T) {
		// Set first
		err := svc.SetEnvironmentConfig(ctx, sdkKey, payload)
		require.NoError(t, err)

		// Verify it exists
		result, err := svc.GetEnvironmentConfig(ctx, sdkKey)
		require.NoError(t, err)
		require.NotNil(t, result)

		// Delete
		err = svc.DeleteEnvironmentConfig(ctx, sdkKey)
		require.NoError(t, err)

		// Verify it is gone
		result, err = svc.GetEnvironmentConfig(ctx, sdkKey)
		require.NoError(t, err)
		assert.Nil(t, result)
	})

	t.Run("Delete - nonexistent key does not error", func(t *testing.T) {
		err := svc.DeleteEnvironmentConfig(ctx, "key-that-never-existed")
		require.NoError(t, err)
	})

	t.Run("Set - overwrites existing value", func(t *testing.T) {
		original := &LeanEnvironmentConfig{
			IsActive:     false,
			TenantID:     "tenant-old",
			TenantStatus: domain.TenantStatusSuspended,
			FeatureFlags: map[string]LeanFeatureFlag{},
		}
		updated := &LeanEnvironmentConfig{
			IsActive:     true,
			TenantID:     "tenant-old",
			TenantStatus: domain.TenantStatusActive,
			FeatureFlags: map[string]LeanFeatureFlag{
				"flag-new": {DefaultStatus: true},
			},
		}

		err := svc.SetEnvironmentConfig(ctx, sdkKey, original)
		require.NoError(t, err)

		err = svc.SetEnvironmentConfig(ctx, sdkKey, updated)
		require.NoError(t, err)

		result, err := svc.GetEnvironmentConfig(ctx, sdkKey)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.True(t, result.IsActive)
		assert.Len(t, result.FeatureFlags, 1)
		_, ok := result.FeatureFlags["flag-new"]
		assert.True(t, ok)
	})
}

func TestTenantLevelCacheService_SetAndGetByID(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()
	ctx := context.Background()

	level := &domain.TenantLevel{
		ID:               "level-001",
		Code:             "pro",
		Name:             "Pro Plan",
		MaxEnvironments:  10,
		MaxFeatureFlags:  100,
		MaxTenantAdmins:  5,
		MaxTenantViewers: 20,
		AllowSSO:         true,
		AllowAPIAccess:   true,
		HasAuditLog:      true,
		PrioritySupport:  false,
		PriceMonthly:     99.00,
		PriceYearly:      999.00,
		IsPublic:         true,
		IsActive:         true,
	}

	t.Run("Set and GetByID - success", func(t *testing.T) {
		err := svc.SetTenantLevel(ctx, level)
		require.NoError(t, err)

		result, err := svc.GetTenantLevelByID(ctx, level.ID)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, level.ID, result.ID)
		assert.Equal(t, level.Code, result.Code)
		assert.Equal(t, level.Name, result.Name)
		assert.Equal(t, level.MaxEnvironments, result.MaxEnvironments)
		assert.Equal(t, level.MaxFeatureFlags, result.MaxFeatureFlags)
		assert.Equal(t, level.MaxTenantAdmins, result.MaxTenantAdmins)
		assert.Equal(t, level.MaxTenantViewers, result.MaxTenantViewers)
		assert.Equal(t, level.AllowSSO, result.AllowSSO)
		assert.Equal(t, level.AllowAPIAccess, result.AllowAPIAccess)
		assert.Equal(t, level.HasAuditLog, result.HasAuditLog)
		assert.Equal(t, level.PrioritySupport, result.PrioritySupport)
		assert.Equal(t, level.PriceMonthly, result.PriceMonthly)
		assert.Equal(t, level.PriceYearly, result.PriceYearly)
		assert.Equal(t, level.IsPublic, result.IsPublic)
		assert.Equal(t, level.IsActive, result.IsActive)
	})

	t.Run("GetByID - cache miss returns nil, nil", func(t *testing.T) {
		result, err := svc.GetTenantLevelByID(ctx, "nonexistent-id")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestTenantLevelCacheService_GetByCode(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()
	ctx := context.Background()

	level := &domain.TenantLevel{
		ID:              "level-002",
		Code:            "enterprise",
		Name:            "Enterprise Plan",
		MaxEnvironments: 50,
		MaxFeatureFlags: 1000,
		IsActive:        true,
	}

	t.Run("Set and GetByCode - success", func(t *testing.T) {
		err := svc.SetTenantLevel(ctx, level)
		require.NoError(t, err)

		result, err := svc.GetTenantLevelByCode(ctx, level.Code)
		require.NoError(t, err)
		require.NotNil(t, result)

		assert.Equal(t, level.ID, result.ID)
		assert.Equal(t, level.Code, result.Code)
		assert.Equal(t, level.Name, result.Name)
	})

	t.Run("GetByCode - cache miss returns nil, nil", func(t *testing.T) {
		result, err := svc.GetTenantLevelByCode(ctx, "nonexistent-code")
		require.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestTenantLevelCacheService_Delete(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()
	ctx := context.Background()

	level := &domain.TenantLevel{
		ID:              "level-003",
		Code:            "free",
		Name:            "Free Plan",
		MaxEnvironments: 2,
		MaxFeatureFlags: 10,
		IsActive:        true,
	}

	t.Run("Delete - removes both ID and Code cache entries", func(t *testing.T) {
		err := svc.SetTenantLevel(ctx, level)
		require.NoError(t, err)

		// Verify both lookups work
		byID, err := svc.GetTenantLevelByID(ctx, level.ID)
		require.NoError(t, err)
		require.NotNil(t, byID)

		byCode, err := svc.GetTenantLevelByCode(ctx, level.Code)
		require.NoError(t, err)
		require.NotNil(t, byCode)

		// Delete
		err = svc.DeleteTenantLevel(ctx, level)
		require.NoError(t, err)

		// Verify both are gone
		byID, err = svc.GetTenantLevelByID(ctx, level.ID)
		require.NoError(t, err)
		assert.Nil(t, byID)

		byCode, err = svc.GetTenantLevelByCode(ctx, level.Code)
		require.NoError(t, err)
		assert.Nil(t, byCode)
	})

	t.Run("Delete - nonexistent level does not error", func(t *testing.T) {
		nonexistent := &domain.TenantLevel{
			ID:   "no-such-id",
			Code: "no-such-code",
		}
		err := svc.DeleteTenantLevel(ctx, nonexistent)
		require.NoError(t, err)
	})
}

func TestTenantLevelCacheService_Update(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()
	ctx := context.Background()

	original := &domain.TenantLevel{
		ID:              "level-004",
		Code:            "basic",
		Name:            "Basic Plan",
		MaxEnvironments: 3,
		MaxFeatureFlags: 30,
		IsActive:        true,
	}

	updated := &domain.TenantLevel{
		ID:              "level-004",
		Code:            "basic",
		Name:            "Basic Plan (Updated)",
		MaxEnvironments: 5,
		MaxFeatureFlags: 50,
		IsActive:        true,
	}

	t.Run("Set - overwrites existing cache entry", func(t *testing.T) {
		err := svc.SetTenantLevel(ctx, original)
		require.NoError(t, err)

		// Verify original
		result, err := svc.GetTenantLevelByID(ctx, original.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Basic Plan", result.Name)

		// Overwrite
		err = svc.SetTenantLevel(ctx, updated)
		require.NoError(t, err)

		// Verify updated
		result, err = svc.GetTenantLevelByID(ctx, original.ID)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Basic Plan (Updated)", result.Name)
		assert.Equal(t, 5, result.MaxEnvironments)

		// Verify GetByCode also returns updated
		result, err = svc.GetTenantLevelByCode(ctx, original.Code)
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Equal(t, "Basic Plan (Updated)", result.Name)
	})
}

func TestEnvironmentCacheService_KeyGeneration(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()

	// Test the key generation
	sdkKey := "my-sdk-key"
	expectedKey := "env_config:my-sdk-key"
	actualKey := svc.generateEnvironmentCacheKey(sdkKey)
	assert.Equal(t, expectedKey, actualKey)
}

func TestTenantLevelCacheService_KeyGeneration(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()

	t.Run("ID key generation", func(t *testing.T) {
		levelID := "level-123"
		expectedKey := "tenant_level:id:level-123"
		actualKey := svc.generateTenantLevelIDKey(levelID)
		assert.Equal(t, expectedKey, actualKey)
	})

	t.Run("Code key generation", func(t *testing.T) {
		code := "pro-plan"
		expectedKey := "tenant_level:code:pro-plan"
		actualKey := svc.generateTenantLevelCodeKey(code)
		assert.Equal(t, expectedKey, actualKey)
	})
}

func TestEnvironmentCacheService_EmptyFeatureFlags(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()
	ctx := context.Background()

	payload := &LeanEnvironmentConfig{
		IsActive:     true,
		TenantID:     "tenant-empty",
		TenantStatus: domain.TenantStatusActive,
		FeatureFlags: map[string]LeanFeatureFlag{},
	}

	t.Run("Set and Get - empty feature flags map", func(t *testing.T) {
		err := svc.SetEnvironmentConfig(ctx, "sdk-empty", payload)
		require.NoError(t, err)

		result, err := svc.GetEnvironmentConfig(ctx, "sdk-empty")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Empty(t, result.FeatureFlags)
	})
}

func TestEnvironmentCacheService_NilFeatureFlags(t *testing.T) {
	svc, mr := setupTestRedis(t)
	defer mr.Close()
	ctx := context.Background()

	payload := &LeanEnvironmentConfig{
		IsActive:     true,
		TenantID:     "tenant-nil",
		TenantStatus: domain.TenantStatusActive,
	}

	t.Run("Set and Get - nil feature flags", func(t *testing.T) {
		err := svc.SetEnvironmentConfig(ctx, "sdk-nil", payload)
		require.NoError(t, err)

		result, err := svc.GetEnvironmentConfig(ctx, "sdk-nil")
		require.NoError(t, err)
		require.NotNil(t, result)
		assert.Nil(t, result.FeatureFlags)
	})
}