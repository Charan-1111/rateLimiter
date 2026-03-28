package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/metrics"
	"goapp/models"
	"goapp/services"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type RateLimiter interface {
	Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, tenantId string, userId string) (*models.LimiterResponse, error)
}

type metricsLimiter struct {
	base RateLimiter
	algo string
}

func (m *metricsLimiter) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, tenantId string, userId string) (*models.LimiterResponse, error) {
	start := time.Now()
	allowed, err := m.base.Allow(ctx, rdb, cb, log, tenantId, userId)
	duration := time.Since(start).Seconds()

	metrics.RequestsLatency.Observe(duration)

	status := "allowed"
	if err != nil {
		status = "error"
	} else if !allowed.Allowed {
		status = "denied"
	}

	metrics.Requests.WithLabelValues(status, m.algo).Inc()

	return allowed, err
}

type LimiterFactory interface {
	GetLimiter(ctx context.Context, db *pgxpool.Pool, log zerolog.Logger, scope, identifier, rateLimitType, query string, cache *services.Cache) (RateLimiter, error)
}

type DefaultLimiterFactory struct{}

type constructor func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter

var registry = map[string]map[string]constructor{
	constants.AlgorithmTokenBucket: {
		constants.ValeTypeMemory: func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter {
			return NewTokenBucketMem(float64(policy.Limit), float64(policy.Burst), log)
		},
		constants.ValueTypeRedis: func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter {
			return NewTokenBucket(float64(policy.Limit), float64(policy.Burst), log)
		},
	},
	constants.AlgorithmLeakyBucket: {
		constants.ValeTypeMemory: func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter {
			return NewLeakyBucketMem(float64(policy.Limit), float64(policy.Burst), log)
		},
		constants.ValueTypeRedis: func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter {
			return NewLeakyBucket(float64(policy.Limit), float64(policy.Burst), log)
		},
	},
	constants.AlgorithmFixedWindow: {
		constants.ValeTypeMemory: func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter {
			return NewFixedWindowMem(policy.Window, policy.Limit, log)
		},
		constants.ValueTypeRedis: func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter {
			return NewFixedWindowCounter(policy.Window, int64(policy.Limit), log)
		},
	},
	constants.AlgorithmSlidingWindow: {
		constants.ValeTypeMemory: func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter {
			return NewSlidingWindowMem(policy.Window, policy.Limit, log)
		},
		constants.ValueTypeRedis: func(policy *services.PolicySchema, log zerolog.Logger) RateLimiter {
			return NewSlidingWindowCounter(policy.Window, policy.Limit, log)
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

	limiter := constructor(policy, log)
	return &metricsLimiter{
		base: limiter,
		algo: policy.Algorithm,
	}, nil
}
