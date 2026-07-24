package service

import (
	"context"
	"facet-backend/internal/cache"
	"facet-backend/internal/domain"
	"facet-backend/internal/observability"
	"fmt"
	"log/slog"
	"time"

	"golang.org/x/sync/singleflight"
)

// EvaluationDataSource defines the interface for the evaluation data source,
// whose sole responsibility is to provide stable and efficient config data for the evaluation system.
type EvaluationDataSource interface {
	GetLeanConfig(ctx context.Context, sdkKey string) (*cache.LeanEnvironmentConfig, error)
}

// evaluationDataSource is the implementation of the interface, encapsulating all data fetching complexity.
type evaluationDataSource struct {
	cacheService   cache.Service
	envRepo        domain.EnvironmentRepository
	tenantRepo     domain.TenantRepository
	ffRepo         domain.FeatureFlagRepository
	ruleRepo       domain.FeatureFlagRuleRepository
	segRepo        domain.SegmentRepository
	condRepo       domain.SegmentConditionRepository
	ffSegRepo      domain.FeatureFlagSegmentRepository
	sfGroup        singleflight.Group
	logger         *slog.Logger
	metrics        *observability.MetricsLogger
}

// NewEvaluationDataSource creates a new data source instance
func NewEvaluationDataSource(
	cacheService cache.Service,
	envRepo domain.EnvironmentRepository,
	tenantRepo domain.TenantRepository,
	ffRepo domain.FeatureFlagRepository,
	ruleRepo domain.FeatureFlagRuleRepository,
	segRepo domain.SegmentRepository,
	condRepo domain.SegmentConditionRepository,
	ffSegRepo domain.FeatureFlagSegmentRepository,
	logger *slog.Logger,
) EvaluationDataSource {
	return &evaluationDataSource{
		cacheService: cacheService,
		envRepo:      envRepo,
		tenantRepo:   tenantRepo,
		ffRepo:       ffRepo,
		ruleRepo:     ruleRepo,
		segRepo:      segRepo,
		condRepo:     condRepo,
		ffSegRepo:    ffSegRepo,
		logger:       logger.With("component", "EvaluationDataSource"),
		metrics:      observability.NewMetricsLogger(logger),
	}
}

// GetLeanConfig is the core method that orchestrates cache, database, and singleflight logic
func (s *evaluationDataSource) GetLeanConfig(ctx context.Context, sdkKey string) (*cache.LeanEnvironmentConfig, error) {
	// 1. Record total duration
	startTime := time.Now()
	defer func() {
		s.metrics.Observe("evaluation_datasource_get_config_duration_seconds", time.Since(startTime).Seconds(), nil)
	}()

	// 2. Check cache
	payload, err := s.cacheService.GetEnvironmentConfig(ctx, sdkKey)
	if err != nil {
		s.logger.Error("failed to get from cache", "sdkKey", sdkKey, "error", err)
	}
	if payload != nil {
		s.metrics.Incr("evaluation_datasource_cache_requests_total", map[string]string{"result": "hit"}) // <-- cache hit
		return payload, nil
	}

	s.metrics.Incr("evaluation_datasource_cache_requests_total", map[string]string{"result": "miss"}) // <-- cache miss

	// 3. Cache miss, use singleflight
	v, err, shared := s.sfGroup.Do(sdkKey, func() (interface{}, error) {
		// This function is only called by the "executor", so record directly here
		s.metrics.Incr("evaluation_datasource_singleflight_requests_total", map[string]string{"type": "executed"}) // <-- singleflight executor

		// Record DB load duration
		dbStartTime := time.Now()
		defer func() {
			s.metrics.Observe("evaluation_datasource_db_load_duration_seconds", time.Since(dbStartTime).Seconds(), nil)
		}()

		s.logger.Info("rebuilding cache", "sdkKey", sdkKey)
		return s.loadFromDBAndCache(ctx, sdkKey)
	})

	// If the result is shared, the current request is a "waiter"
	if shared {
		s.metrics.Incr("evaluation_datasource_singleflight_requests_total", map[string]string{"type": "shared"}) // <-- singleflight waiter
	}

	if err != nil {
		return nil, fmt.Errorf("singleflight loadFromDBAndCache execution failed: %w", err)
	}

	// Type assert the singleflight return value
	payload, ok := v.(*cache.LeanEnvironmentConfig)
	if !ok {
		// This theoretically should not happen, but as a defensive programming measure, we handle it
		return nil, fmt.Errorf("singleflight returned unexpected type")
	}

	return payload, nil
}

// loadFromDBAndCache is a private method that loads data from the database,
// transforms it into an optimized "lean" structure, and then caches it.
func (s *evaluationDataSource) loadFromDBAndCache(ctx context.Context, sdkKey string) (*cache.LeanEnvironmentConfig, error) {
	// 1. Retrieve and validate environment
	env, err := s.envRepo.FindBySDKKey(ctx, sdkKey)
	if err != nil {
		return nil, err
	}

	// 2. Retrieve and validate tenant
	tenant, err := s.tenantRepo.GetByID(ctx, env.TenantID)
	if err != nil {
		return nil, err
	}

	// Defensive check
	if env.TenantID != tenant.ID {
		return nil, fmt.Errorf("data inconsistency: environment tenant ID (%s) does not match tenant ID (%s)", env.TenantID, tenant.ID)
	}

	// 3. Get all feature flags for the environment
	flags, err := s.ffRepo.FindByEnvironmentID(ctx, env.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get feature flags for environment %s: %w", env.ID, err)
	}

	// --- Data transformation ---
	featureFlagsMap := make(map[string]cache.LeanFeatureFlag, len(flags))
	for _, flag := range flags {
		rules, err := s.ruleRepo.GetRulesByFlagID(ctx, flag.ID)
		if err != nil {
			return nil, fmt.Errorf("failed to get rules for flag %s: %w", flag.ID, err)
		}

		leanRules := make([]cache.LeanRule, len(rules))
		for i, rule := range rules {
			leanRules[i] = cache.LeanRule{
				AccountID: rule.AccountID.String,
				UserIDs:   rule.UserIDs,
			}
		}

		segIDs, _ := s.ffSegRepo.GetSegmentIDs(ctx, flag.ID)
		var leanSegments []cache.LeanSegment
		for _, segID := range segIDs {
			seg, err := s.segRepo.GetByID(ctx, segID, tenant.ID)
			if err != nil {
				continue
			}
			conds, err := s.condRepo.ListBySegmentID(ctx, segID)
			if err != nil {
				continue
			}
			leanConds := make([]cache.LeanSegmentCondition, len(conds))
			for j, c := range conds {
				leanConds[j] = cache.LeanSegmentCondition{
					AttributeKey: c.AttributeKey,
					Operator:     c.Operator,
					Value:        c.Value,
				}
			}
			leanSegments = append(leanSegments, cache.LeanSegment{
				MatchType:  string(seg.MatchType),
				Conditions: leanConds,
			})
		}

		featureFlagsMap[flag.Name] = cache.LeanFeatureFlag{
			DefaultStatus:     flag.Enabled,
			RolloutPercentage: flag.RolloutPercentage,
			Rules:             leanRules,
			Segments:          leanSegments,
		}
	}

	payload := &cache.LeanEnvironmentConfig{
		IsActive:     env.IsActive,
		TenantID:     tenant.ID,
		TenantStatus: tenant.Status,
		FeatureFlags: featureFlagsMap,
	}

	// 4. Store the transformed "lean" structure in cache
	if err := s.cacheService.SetEnvironmentConfig(ctx, sdkKey, payload); err != nil {
		slog.Warn("cache store failed", "sdkKey", sdkKey, "error", err)
	}

	return payload, nil
}
