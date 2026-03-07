package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/lua"
	"goapp/store"
	"time"

	"github.com/redis/go-redis/v9"
)

type TokensRedis struct {
	Tokens     float64
	LastRefill time.Time
}

type TokenBucketRedis struct {
	MaxTokens  float64
	RefillRate float64
}

func NewTokenBucket(maxTokens, refillRate float64) *TokenBucketRedis {
	return &TokenBucketRedis{
		MaxTokens:  maxTokens,
		RefillRate: float64(refillRate),
	}
}

func (tb *TokenBucketRedis) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	// tokens := &Tokens{}

	// get the information from the redis for the key
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmTokenBucket, tenantId, userId)

	tokenBucketScript := redis.NewScript(lua.GetTokenBucketScript())
	now := float64(time.Now().UnixNano()) / 1e9

	if store.Rdb == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	_, err := tokenBucketScript.Run(
		ctx,
		store.Rdb,
		[]string{redisKey},
		tb.MaxTokens,
		tb.RefillRate,
		now,
		1,
	).Result()

	if err != nil {
		fmt.Println("Error running the token bucket script : ", err)
		return false, err
	}

	return true, nil
}
