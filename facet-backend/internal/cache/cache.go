package cache

// Service is a top-level cache service interface that aggregates all domain-specific cache services.
// This design follows the Interface Segregation Principle, allowing dependent services to only care about the cache capabilities they need.
type Service interface {
	EnvironmentCacheService
	TenantLevelCacheService
}
