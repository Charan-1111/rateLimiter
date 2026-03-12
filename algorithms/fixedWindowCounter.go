package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/lua"
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

func NewFixedWindowCounter(windowStr string, capacity int64) *FixedCounterRedis {
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		fmt.Println("Error parsing the duration")
	}
	
	return &FixedCounterRedis{
		window:   window,
		capacity: capacity,
	}
}

func (fc *FixedCounterRedis) Allow(ctx context.Context, rdb *redis.Client, tenandId, userId string) (bool, error) {
	now := time.Now().UnixNano()

	window := fc.window.Microseconds()

	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmFixedWindow, tenandId, userId)

	fwcScript := redis.NewScript(lua.GetFixedWindowCounterScript())

	_, err := fwcScript.Run(ctx, rdb, []string{redisKey}, fc.capacity, window, now, 1).Result()
	if err != nil {
		fmt.Println("Error calling the fixed window counter script, rejecting the request : ", err)
		return false, err
	} else {
		fmt.Println("Accepting the request")
	}
	return true, nil
}
