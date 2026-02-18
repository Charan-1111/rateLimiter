package algorithms

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type TokenBucket struct {
	maxTokens  float64
	tokens     float64
	lastRefill time.Time
	refillRate float64
	mu         sync.Mutex
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		maxTokens:  maxTokens,
		tokens:     maxTokens,
		lastRefill: time.Now(),
		refillRate: float64(refillRate),
	}
}

func (tb *TokenBucket) IsBucketFull() bool {
	return tb.tokens == tb.maxTokens
}

func (tb *TokenBucket) IsBucketEmpty() bool {
	return tb.tokens == 0
}

func (tb *TokenBucket) RefillTokens() {
	tb.tokens = tb.tokens + (time.Since(tb.lastRefill).Seconds() * tb.refillRate)
	if tb.tokens >= tb.maxTokens {
		tb.tokens = tb.maxTokens
	}
	tb.lastRefill = time.Now()
}


func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	tb.RefillTokens()

	if tb.IsBucketEmpty() {
		return false, fmt.Errorf("bucket is empty, request is getting denied")
	}

	fmt.Println("token available, request is getting proceed")

	tb.tokens -= 1

	return true, nil
}
