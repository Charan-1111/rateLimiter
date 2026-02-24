package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/store"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}

type Tokens struct {
	tokens     float64
	lastRefill time.Time
}

type TokenBucket struct {
	maxTokens  float64
	keys       map[string]*Tokens
	refillRate float64
	mu         sync.Mutex
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		maxTokens:  maxTokens,
		keys:       make(map[string]*Tokens),
		refillRate: float64(refillRate),
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	// get the information from the redis for the key
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmTokenBucket, tenantId, userId)

	val, err := store.Rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		tb.keys[userId] = &Tokens{
			tokens:     tb.maxTokens,
			lastRefill: time.Now(),
		}
	} else if err != nil {
		fmt.Println("error getting the key from redis", err)
		return false, err
	} else {
		fmt.Println(val)
	}

	// refill the tokens for this key

	if _, ok := tb.keys[userId]; !ok {
		tb.keys[userId] = &Tokens{
			tokens:     tb.maxTokens,
			lastRefill: time.Now(),
		}
	}

	tb.keys[userId].tokens = tb.keys[userId].tokens + (time.Since(tb.keys[userId].lastRefill).Seconds() * tb.refillRate)
	if tb.keys[userId].tokens >= tb.maxTokens {
		tb.keys[userId].tokens = tb.maxTokens
	}
	tb.keys[userId].lastRefill = time.Now()

	if tb.keys[userId].tokens < 1 {
		return false, fmt.Errorf("bucket is empty, request is getting denied")
	}

	fmt.Println("token available, request is getting proceed")

	tb.keys[userId].tokens -= 1

	return true, nil
}
