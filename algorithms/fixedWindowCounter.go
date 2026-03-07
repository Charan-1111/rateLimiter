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

type FixedCounterRedisStore struct {
	WindowIndex int64
	Allowed     int64
}

type FixedCounterRedis struct {
	window   time.Duration
	capacity int64
}

func NewFixedWindowCounter(window time.Duration, capacity int64) *FixedCounterRedis {
	return &FixedCounterRedis{
		window:   window,
		capacity: capacity,
	}
}

func (fc *FixedCounterRedis) Allow(ctx context.Context, tenandId, userId string) (bool, error) {
	now := time.Now().UnixNano()

	window := fc.window.Microseconds()

	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmFixedWindow, tenandId, userId)

	fwcScript := redis.NewScript(lua.GetFixedWindowCounterScript())

	if store.Rdb == nil {
		return false, fmt.Errorf("redis client not initialized")
	}

	_, err := fwcScript.Run(ctx, store.Rdb, []string{redisKey}, fc.capacity, window, now, 1).Result()
	if err != nil {
		fmt.Println("Error calling the fixed window counter script, rejecting the request : ", err)
		return false, err
	} else {
		fmt.Println("Accepting the request")
	}
	return true, nil
}
