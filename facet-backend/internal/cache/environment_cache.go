package cache

import (
	"context"
	"facet-backend/internal/domain"
)

// LeanEnvironmentConfig is an optimized environment configuration data structure used for caching.
type LeanEnvironmentConfig struct {
	IsActive     bool                       `json:"isActive"`
	TenantID     string                     `json:"tenantId"`
	TenantStatus domain.TenantStatus        `json:"tenantStatus"`
	FeatureFlags map[string]LeanFeatureFlag `json:"featureFlags"`
}

// LeanFeatureFlag only contains the information needed to evaluate a single flag.
type LeanFeatureFlag struct {
	DefaultStatus     bool                   `json:"defaultStatus"`
	RolloutPercentage int                    `json:"rolloutPercentage"`
	Rules             []LeanRule             `json:"rules"`
	Segments          []LeanSegment          `json:"segments"`
}

// LeanRule only contains the information needed to evaluate a single rule.
type LeanRule struct {
	AccountID string   `json:"accountId,omitempty"`
	UserIDs   []string `json:"userIds"`
}

// LeanSegmentCondition holds a single segment condition for evaluation.
type LeanSegmentCondition struct {
	AttributeKey string `json:"attributeKey"`
	Operator     string `json:"operator"`
	Value        string `json:"value"`
}

// LeanSegment holds a segment's conditions and match type for evaluation.
type LeanSegment struct {
	MatchType  string                 `json:"matchType"`
	Conditions []LeanSegmentCondition `json:"conditions"`
}

// EnvironmentCacheService defines the cache operation interface for environment configuration.
// This decouples business logic from specific cache implementations (e.g., Redis).
type EnvironmentCacheService interface {
	// GetEnvironmentConfig attempts to retrieve the optimized environment configuration from cache.
	// If the cache misses, it returns (nil, nil).
	GetEnvironmentConfig(ctx context.Context, sdkKey string) (*LeanEnvironmentConfig, error)

	// SetEnvironmentConfig stores an optimized environment configuration in cache with a standard expiration time.
	SetEnvironmentConfig(ctx context.Context, sdkKey string, payload *LeanEnvironmentConfig) error

	// DeleteEnvironmentConfig deletes the specified environment configuration from cache.
	DeleteEnvironmentConfig(ctx context.Context, sdkKey string) error
}
