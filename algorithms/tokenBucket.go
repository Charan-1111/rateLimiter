package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/store"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

type RateLimiter interface {
	Allow(ctx context.Context, tenantId string, userId string) (bool, error)
}

type Tokens struct {
	Tokens     float64
	LastRefill time.Time
}

type TokenBucket struct {
	MaxTokens  float64
	RefillRate float64
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucket {
	return &TokenBucket{
		MaxTokens:  maxTokens,
		RefillRate: float64(refillRate),
	}
}

func (tb *TokenBucket) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	tokens := &Tokens{}

	// get the information from the redis for the key
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmTokenBucket, tenantId, userId)

	val, err := store.Rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		fmt.Println("redis is null")
		tokens.Tokens = tb.MaxTokens
		tokens.LastRefill = time.Now()
	} else if err != nil {
		fmt.Println("error getting the key from redis", err)
		return false, err
	} else {
		// unmarshal the value
		err = sonic.Unmarshal([]byte(val), &tokens)
		if err != nil {
			fmt.Println("error unmarshalling the value", err)
			return false, err
		}
	}

	// refill the tokens for this key
	tokens.Tokens = tokens.Tokens + (time.Since(tokens.LastRefill).Minutes() * tb.RefillRate)
	if tokens.Tokens >= tb.MaxTokens {
		tokens.Tokens = tb.MaxTokens
	}
	tokens.LastRefill = time.Now()

	fmt.Println("tokens : ", tokens)

	if tokens.Tokens < 1 {
		return false, fmt.Errorf("bucket is empty, request is getting denied")
	}

	fmt.Println("token available, request is getting proceed")

	tokens.Tokens = tokens.Tokens - 1

	// Set the information in the redis
	fmt.Println("tokens : ", tokens)

	marshaledVal, err := sonic.Marshal(tokens)
	if err != nil {
		fmt.Println("Error marshaling the error : ", err)
		return false, err
	}

	err = store.Rdb.Set(ctx, redisKey, marshaledVal, time.Second*60).Err()
	if err != nil {
		fmt.Println("error setting the key in redis", err)
		return false, err
	}

	return true, nil
}
