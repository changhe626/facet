package cache

import (
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	// cacheTTL is the standard expiration time for all caches in our system
	cacheTTL = 24 * time.Hour
)

// redisCacheService is the Redis implementation of the cache.Service interface.
// It composes its functionality by implementing parts of the interface across different files.
type redisCacheService struct {
	redisClient *redis.Client
}

// NewRedisCacheService creates a new redisCacheService instance that implements all cache interfaces.
func NewRedisCacheService(redisClient *redis.Client) Service {
	return &redisCacheService{
		redisClient: redisClient,
	}
}
