package memoryalgorithms

import (
	"context"
	"errors"
	"fmt"
	"goapp/constants"
	"sync"
	"time"
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

func (lb *LeakyBucket) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmLeakyBucket, tenantId, userId)
	now := time.Now()

	// Fetch the details from the cache
	tokenStore, ok := lb.tokens[key]
	if !ok {
		tokenStore.tokens = 0
		tokenStore.lastLeak = now
	}

	// leak the tokens
	tokenStore.tokens = tokenStore.tokens - (now.Sub(tokenStore.lastLeak).Seconds() * lb.leakRate)

	if tokenStore.tokens <= 0 {
		tokenStore.tokens = 0
	}
	tokenStore.lastLeak = now

	// check if the bukcet is full
	if tokenStore.tokens >= lb.capacity {
		fmt.Println("request is getting rejected")
		return false, errors.New("Request is getting rejected")
	}

	tokenStore.tokens += 1

	// store the information in the cache
	lb.tokens[key] = tokenStore

	return true, nil
}
