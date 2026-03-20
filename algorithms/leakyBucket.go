package algorithms

import (
	"context"
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

func (lb *LeakyBucketRedis) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	// read data from redis
	redisKey := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmLeakyBucket, scope, identifier)

	leakyScript := redis.NewScript(lua.GetLeakyBucketScript())

	now := float64(time.Now().UnixNano()) / 1e9

	results, err := cb.Cb.Execute(func() (any, error) {
		results, err := leakyScript.Run(ctx, rdb, []string{redisKey}, lb.MaxTokens, lb.LeakRate, now, 1).Result()
		return results, err
	})

	allowed := results.([]any)[0].(bool)
	currentTokens := results.([]any)[1].(float64)

	retryAfter := now + (lb.MaxTokens / lb.LeakRate)

	if err != nil {
		log.Error().Err(err).Msg("Error running the script, rejecting the request")
		return &models.LimiterResponse{
			Allowed:       false,
			RetryAfter:    0,
			CurrentTokens: 0,
		}, err
	} else {
		log.Info().Msg("Accepting the request")
	}

	return &models.LimiterResponse{
		Allowed:       allowed,
		RetryAfter:    int64(retryAfter),
		CurrentTokens: int64(currentTokens),
	}, nil
}
