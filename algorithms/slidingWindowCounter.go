package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/lua"
	"goapp/models"
	"goapp/services"
	"goapp/utils"
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

func (sc *SlidingWindowCounterRedis) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	now := time.Now().UnixMilli()

	windowMs := sc.window.Milliseconds()
	// read data from the redis
	redisKey := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmSlidingWindow, scope, identifier)

	swcScript := redis.NewScript(lua.GetSlidingWindowScript())

	results, err := cb.Cb.Execute(func() (any, error) {
		results, err := swcScript.Run(ctx, rdb, []string{redisKey}, sc.capacity, windowMs, now, 1).Result()

		return results, err
	})

	allowed := results.([]any)[0].(bool)
	tokens := results.([]any)[1].(int64)

	retryAfter := now + (windowMs - tokens)

	if err != nil {
		log.Error().Err(err).Msg("Error calling the sliding window counter script, rejecting the request")
		return &models.LimiterResponse{
			Allowed:       false,
			RetryAfter:    int64(retryAfter),
			CurrentTokens: int64(tokens),
		}, err
	} else {
		log.Info().Msg("Accepting the request")
	}
	return &models.LimiterResponse{
		Allowed:       allowed,
		RetryAfter:    int64(retryAfter),
		CurrentTokens: int64(tokens),
	}, nil
}
