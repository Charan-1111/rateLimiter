package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/services"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type RateLimiter interface {
	Allow(ctx context.Context, rdb *redis.Client, tenantId string, userId string) (bool, error)
}

type LimiterFactory interface {
	GetLimiter(ctx context.Context, db *pgxpool.Pool, log zerolog.Logger, scope, identifer, rateLimitType, query string, cache *services.Cache) (RateLimiter, error)
}

type DefaultLimiterFactory struct{}

type constructor func(policy *services.PolicySchema) RateLimiter

var registry = map[string]map[string]constructor{
	constants.AlgorithmTokenBucket: {
		constants.ValeTypeMemory: func(policy *services.PolicySchema) RateLimiter {
			return NewTokenBucketMem(float64(policy.Limit), float64(policy.Burst))
		},
		constants.ValueTypeRedis: func(policy *services.PolicySchema) RateLimiter {
			return NewTokenBucket(float64(policy.Limit), float64(policy.Burst))
		},
	},
	constants.AlgorithmLeakyBucket: {
		constants.ValeTypeMemory: func(policy *services.PolicySchema) RateLimiter {
			return NewLeakyBucketMem(float64(policy.Limit), float64(policy.Burst))
		},
		constants.ValueTypeRedis: func(policy *services.PolicySchema) RateLimiter {
			return NewLeakyBucket(float64(policy.Limit), float64(policy.Burst))
		},
	},
	constants.AlgorithmFixedWindow: {
		constants.ValeTypeMemory: func(policy *services.PolicySchema) RateLimiter {
			return NewFixedWindowMem(policy.Window, policy.Limit)
		},
		constants.ValueTypeRedis: func(policy *services.PolicySchema) RateLimiter {
			return NewFixedWindowCounter(policy.Window, int64(policy.Limit))
		},
	},
	constants.AlgorithmSlidingWindow: {
		constants.ValeTypeMemory: func(policy *services.PolicySchema) RateLimiter {
			return NewSlidingWindowMem(policy.Window, policy.Limit)
		},
		constants.ValueTypeRedis: func(policy *services.PolicySchema) RateLimiter {
			return NewSlidingWindowCounter(policy.Window, policy.Limit)
		},
	},
}

func (f *DefaultLimiterFactory) GetLimiter(ctx context.Context, db *pgxpool.Pool, log zerolog.Logger, scope, identifier, rateLimitType, query string, cache *services.Cache) (RateLimiter, error) {
	policy, exists := cache.GetPolicy(ctx, db, log, scope, identifier, query)
	if !exists {
		// call the database to get the policy and update the cache
		return nil, fmt.Errorf("no policy found for scope : %s and identifier : %s", scope, identifier)
	}

	// // Implement logic to create and return the appropriate limiter based on the type and algorithm
	algo, ok := registry[policy.Algorithm]
	if !ok {
		return nil, fmt.Errorf("unsupported algorithm: %s", policy.Algorithm)
	}

	constructor, ok := algo[rateLimitType]
	if !ok {
		return nil, fmt.Errorf("unsupported limiter type: %s", rateLimitType)
	}

	return constructor(policy), nil
}
