package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/store"
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(ctx context.Context, tenantId string, userId string) (bool, error)
}

type TokenBucket struct {
	maxTokens  float64
	refillRate float64
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		maxTokens:  maxTokens,
		refillRate: float64(refillRate),
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	var (
		tokens     float64
		lastRefill time.Time
	)

	// get the information from the redis for the key
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmTokenBucket, tenantId, userId)

	val, err := store.Rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		fmt.Println("redis is null")
		tokens = tb.maxTokens
		lastRefill = time.Now()
	} else if err != nil {
		fmt.Println("error getting the key from redis", err)
		return false, err
	} else {
		fmt.Println(val)
	}

	// refill the tokens for this key
	tokens = tokens + (time.Since(lastRefill).Minutes() * tb.refillRate)
	if tokens >= tb.maxTokens {
		tokens = tb.maxTokens
	}
	lastRefill = time.Now()

	fmt.Println("tokens : ", tokens)

	if tokens < 1 {
		return false, fmt.Errorf("bucket is empty, request is getting denied")
	}

	fmt.Println("token available, request is getting proceed")

	tokens = tokens - 1

	// Set the information in the redis
	fmt.Println("tokens : ", tokens)

	err = store.Rdb.Set(ctx, redisKey, tokens, time.Second*60).Err()
	if err != nil {
		fmt.Println("error setting the key in redis", err)
		return false, err
	}

	return true, nil
}
