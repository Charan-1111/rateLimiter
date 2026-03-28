package algorithms

import (
	"context"
	"goapp/constants"
	"goapp/lua"
	"goapp/models"
	"goapp/services"
	"goapp/utils"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type FixedCounterRedisStore struct {
	WindowIndex int64
	Allowed     int64
}

type FixedCounterRedis struct {
	window   time.Duration
	capacity int64
}

func NewFixedWindowCounter(windowStr string, capacity int64, log zerolog.Logger) *FixedCounterRedis {
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing the duration")
	}

	return &FixedCounterRedis{
		window:   window,
		capacity: capacity,
	}
}

func (fc *FixedCounterRedis) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	now := time.Now().UnixNano()

	window := fc.window.Microseconds()

	redisKey := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmFixedWindow, scope, identifier)

	fwcScript := redis.NewScript(lua.GetFixedWindowCounterScript())

	results, err := cb.Cb.Execute(func() (any, error) {
		results, err := fwcScript.Run(ctx, rdb, []string{redisKey}, fc.capacity, window, now, 1).Result()
		return results, err
	})

	allowed := results.([]any)[0].(bool)
	tokens := results.([]any)[1].(int64)

	retryAfter := now + (fc.window.Microseconds() - tokens)

	if err != nil {
		log.Error().Err(err).Msg("Error calling the fixed window counter script, rejecting the request")
		return &models.LimiterResponse{
			Allowed:       false,
			RetryAfter:    int64(retryAfter),
			RemainingTokens: int64(tokens),
			TotalTokens : fc.capacity,
		}, err
	} else {
		log.Info().Msg("Accepting the request")
	}

	return &models.LimiterResponse{
		Allowed:       allowed,
		RetryAfter:    int64(retryAfter),
		RemainingTokens: int64(tokens),
		TotalTokens : fc.capacity,
	}, nil
}
