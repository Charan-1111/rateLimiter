package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/store"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

type LeakyTokens struct {
	Tokens   float64
	LastLeak time.Time
}

type LeakyBucket struct {
	MaxTokens float64
	LeakRate  float64
	mu        sync.Mutex
}

func NewLeakyBucket(maxTokens, leakRate float64) *LeakyBucket {
	return &LeakyBucket{
		MaxTokens: maxTokens,
		LeakRate:  leakRate,
	}
}

func (lb *LeakyBucket) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	tokens := &LeakyTokens{}

	// read data from redis
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmLeakyBucket, tenantId, userId)
	val, err := store.Rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		tokens.Tokens = 0
		tokens.LastLeak = time.Now()
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

	// leak the tokens from the bucket
	now := time.Now()
	
	tokens.Tokens = tokens.Tokens - (now.Sub(tokens.LastLeak).Seconds() * lb.LeakRate)

	if tokens.Tokens < 0 {
		tokens.Tokens = 0
	}
	tokens.LastLeak = now

	if tokens.Tokens >= lb.MaxTokens {
		return false, fmt.Errorf("Bucket is full, request is getting rejected")
	}

	tokens.Tokens = tokens.Tokens + 1
	fmt.Println("Request accepted")

	// store the details in the redis
	marshalVal, err := sonic.Marshal(tokens)
	if err != nil {
		fmt.Println("Error marshaling leaky bucket token details : ", err)
		return false, err
	}

	err = store.Rdb.Set(ctx, redisKey, marshalVal, time.Second*60).Err()
	if err != nil {
		fmt.Println("Error setting the key in redis : ", err)
		return false, err
	}

	return true, nil
}
