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

type TokenBucketStore struct {
	tokens   float64
	lastFill time.Time
}

type TokenBucket struct {
	capacity float64
	fillRate float64
	tokens   map[string]*TokenBucketStore
	mu       sync.Mutex
}

func NewTokenBucketMem(capacity, fillRate float64, log zerolog.Logger) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		fillRate: fillRate,
		tokens:   make(map[string]*TokenBucketStore),
		mu:       sync.Mutex{},
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// make the key
	key := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmTokenBucket, scope, identifier)
	now := time.Now()

	// Fetch from the cache
	tokenStore, ok := tb.tokens[key]
	if !ok {
		// create a new store if this key hasn't been seen before
		log.Debug().Msg("Token bucket not found in cache, creating new one")
		tokenStore = &TokenBucketStore{
			tokens:   tb.capacity,
			lastFill: now,
		}
	}

	// fill the tokens
	tokenStore.tokens = tokenStore.tokens + (now.Sub(tokenStore.lastFill).Seconds() * tb.fillRate)
	if tokenStore.tokens >= tb.capacity {
		tokenStore.tokens = tb.capacity
	}

	tokenStore.lastFill = now

	// check if the bucket is empty
	if tokenStore.tokens == 0 {
		log.Warn().Str("scope", scope).Msg("Request is getting rejected, bucket is empty")
		return &models.LimiterResponse{
			Allowed:         false,
			RetryAfter:      0,
			RemainingTokens: int64(tokenStore.tokens),
			TotalTokens:     int64(tb.capacity),
		}, errors.New("Request is getting rejected")
	}

	tokenStore.tokens -= 1

	// store this in the cache
	tb.tokens[key] = tokenStore

	return &models.LimiterResponse{
		Allowed:         true,
		RetryAfter:      0,
		RemainingTokens: int64(tokenStore.tokens),
		TotalTokens:     int64(tb.capacity),
	}, nil
}
