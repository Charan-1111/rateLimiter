package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/lua"
	"goapp/store"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type LeakyTokensRedis struct {
	Tokens   float64
	LastLeak time.Time
}

type LeakyBucketRedis struct {
	MaxTokens float64
	LeakRate  float64
	mu        sync.Mutex
}

func NewLeakyBucket(maxTokens, leakRate float64) *LeakyBucketRedis {
	return &LeakyBucketRedis{
		MaxTokens: maxTokens,
		LeakRate:  leakRate,
	}
}

func (lb *LeakyBucketRedis) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	// read data from redis
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmLeakyBucket, tenantId, userId)

	leakyScript := redis.NewScript(lua.GetLeakyBucketScript())

	now := float64(time.Now().UnixNano()) / 1e9

	if store.Rdb == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	_, err := leakyScript.Run(ctx, store.Rdb, []string{redisKey}, lb.MaxTokens, lb.LeakRate, now, 1).Result()
	if err != nil {
		fmt.Println("Error running the script, rejecting the request : ", err)
		return false, err
	} else {
		fmt.Println("request accepted")
	}

	return true, nil
}
