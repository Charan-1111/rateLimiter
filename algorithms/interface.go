package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
)

type RateLimiter interface {
	Allow(ctx context.Context, tenantId string, userId string) (bool, error)
}

type LimiterFactory interface {
	GetLimiter(limiterType, algorithm string) (RateLimiter, error)
}

type DefaultLimiterFactory struct{}

type constructor func() RateLimiter

var registry = map[string]map[string]constructor{
	constants.AlgorithmTokenBucket: {
		constants.ValeTypeMemory: func() RateLimiter { return NewTokenBucketMem(10, 1) },
		constants.ValueTypeRedis: func() RateLimiter { return NewTokenBucket(10, 1) },
	},
	constants.AlgorithmLeakyBucket: {
		constants.ValeTypeMemory: func() RateLimiter { return NewLeakyBucketMem(10, 1) },
		constants.ValueTypeRedis: func() RateLimiter { return NewLeakyBucket(10, 1) },
	},
	constants.AlgorithmFixedWindow: {
		constants.ValeTypeMemory: func() RateLimiter { return NewFixedWindowMem(1, 10) },
		constants.ValueTypeRedis: func() RateLimiter { return NewFixedWindowCounter(1, 10) },
	},
	constants.AlgorithmSlidingWindow: {
		constants.ValeTypeMemory: func() RateLimiter { return NewSlidingWindowMem(1, 10) },
		constants.ValueTypeRedis: func() RateLimiter { return NewSlidingWindowCounter(1, 10) },
	},
}

func (f *DefaultLimiterFactory) GetLimiter(limiterType, algorithm string) (RateLimiter, error) {
	// Implement logic to create and return the appropriate limiter based on the type and algorithm
	algo, ok := registry[algorithm]
	if !ok {
		return nil, fmt.Errorf("unsupported algorithm: %s", algorithm)
	}

	constructor, ok := algo[limiterType]
	if !ok {
		return nil, fmt.Errorf("unsupported limiter type: %s", limiterType)
	}

	return constructor(), nil
}
