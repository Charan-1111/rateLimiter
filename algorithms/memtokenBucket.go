package algorithms

import (
	"context"
	"errors"
	"fmt"
	"goapp/constants"
	"sync"
	"time"
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

func NewTokenBucketMem(capacity, fillRate float64) *TokenBucket {
	return &TokenBucket{
		capacity: capacity,
		fillRate: fillRate,
		tokens:   make(map[string]*TokenBucketStore),
		mu:       sync.Mutex{},
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// make the key
	key := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmTokenBucket, tenantId, userId)
	now := time.Now()

	// Fetch from the cache
	tokenStore, ok := tb.tokens[key]
	if !ok {
		// create a new store if this key hasn't been seen before
		fmt.Println("Not found in the cache")
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
		fmt.Println("Request is getting rejected, bucket is empty")
		return false, errors.New("Request is getting rejected")
	}

	tokenStore.tokens -= 1

	// store this in the cache
	tb.tokens[key] = tokenStore

	return true, nil
}
