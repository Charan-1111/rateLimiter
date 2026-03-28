package algorithms

import (
	"context"
	"errors"
	"goapp/constants"
	"goapp/models"
	"goapp/services"
	"goapp/utils"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type LeakyBucketStore struct {
	tokens   float64
	lastLeak time.Time
}

type LeakyBucket struct {
	capacity float64
	leakRate float64
	tokens   map[string]*LeakyBucketStore
	mu       sync.Mutex
}

func NewLeakyBucketMem(capacity, leakRate float64, log zerolog.Logger) *LeakyBucket {
	return &LeakyBucket{
		capacity: capacity,
		leakRate: leakRate,
		tokens:   make(map[string]*LeakyBucketStore),
		mu:       sync.Mutex{},
	}
}

func (lb *LeakyBucket) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	key := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmLeakyBucket, scope, identifier)
	now := time.Now()

	// Fetch the details from the cache
	tokenStore, ok := lb.tokens[key]
	if !ok {
		// allocate a fresh store
		tokenStore = &LeakyBucketStore{
			tokens:   0,
			lastLeak: now,
		}
	}

	// leak the tokens
	tokenStore.tokens = tokenStore.tokens - (now.Sub(tokenStore.lastLeak).Seconds() * lb.leakRate)

	if tokenStore.tokens <= 0 {
		tokenStore.tokens = 0
	}
	tokenStore.lastLeak = now

	// check if the bukcet is full
	if tokenStore.tokens >= lb.capacity {
		log.Warn().Str("scope", scope).Msg("Request is getting rejected, bucket is full")
		return &models.LimiterResponse{
			Allowed:         false,
			RetryAfter:      0,
			RemainingTokens: int64(tokenStore.tokens),
			TotalTokens:     int64(lb.capacity),
		}, errors.New("Request is getting rejected")
	}

	tokenStore.tokens += 1

	// store the information in the cache
	lb.tokens[key] = tokenStore

	return &models.LimiterResponse{
		Allowed:         true,
		RetryAfter:      0,
		RemainingTokens: int64(tokenStore.tokens),
		TotalTokens:     int64(lb.capacity),
	}, nil
}
