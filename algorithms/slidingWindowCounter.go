package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/lua"
	"goapp/services"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type SlidingWindowStoreRedis struct {
	CurrentCnt  int
	PreviousCnt int
	WindowStart time.Time
}
type SlidingWindowCounterRedis struct {
	window   time.Duration
	capacity int
	mu       sync.Mutex
}

func NewSlidingWindowCounter(windowStr string, capacity int) *SlidingWindowCounterRedis {
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		fmt.Println("Error parsing the duration")
	}

	return &SlidingWindowCounterRedis{
		window:   window,
		capacity: capacity,
	}
}

func (sc *SlidingWindowCounterRedis) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, tenantId, userId string) (bool, error) {
	now := time.Now().UnixMilli()

	windowMs := sc.window.Milliseconds()
	// read data from the redis
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmSlidingWindow, tenantId, userId)

	swcScript := redis.NewScript(lua.GetSlidingWindowScript())

	// _, err := swcScript.Run(ctx, rdb, []string{redisKey}, sc.capacity, windowMs, now, 1).Result()
	// if err != nil {
	// 	fmt.Println("Error calling the sliding window counter script, rejecting the request : ", err)
	// 	return false, err
	// } else {
	// 	fmt.Println("Accepting the request")
	// }

	_, err := cb.Cb.Execute(func() (any, error) {
		return swcScript.Run(ctx, rdb, []string{redisKey}, sc.capacity, windowMs, now, 1).Result()
	})

	if err != nil {
		log.Error().Err(err).Msg("Error calling the sliding window counter script, rejecting the request")
		return false, err
	} else {
		log.Info().Msg("Accepting the request")
	}
	return true, nil
}
