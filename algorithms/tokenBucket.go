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

type TokensRedis struct {
	Tokens     float64
	LastRefill time.Time
}

type TokenBucketRedis struct {
	MaxTokens  float64
	RefillRate float64
}

func NewTokenBucket(maxTokens, refillRate float64, log zerolog.Logger) *TokenBucketRedis {
	return &TokenBucketRedis{
		MaxTokens:  maxTokens,
		RefillRate: float64(refillRate),
	}
}

func (tb *TokenBucketRedis) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	// get the information from the redis for the key
	redisKey := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmTokenBucket, scope, identifier)

	tokenBucketScript := redis.NewScript(lua.GetTokenBucketScript())
	now := float64(time.Now().UnixNano()) / 1e9

	results, err := cb.Cb.Execute(func() (any, error) {
		results, err := tokenBucketScript.Run(ctx, rdb, []string{redisKey}, tb.MaxTokens, tb.RefillRate, now, 1).Result()
		return results, err
	})

	allowed := results.([]any)[0].(bool)
	currentTokens := results.([]any)[1].(int64)

	if err != nil {
		log.Error().Err(err).Msg("Error running the token bucket script")
		return &models.LimiterResponse{
			Allowed:       false,
			RetryAfter:    0,
			RemainingTokens: 0,
			TotalTokens : int64(tb.MaxTokens),
		}, err
	} else {
		log.Info().Msg("Accepting the request")
	}

	retryAfter := now + (tb.MaxTokens-float64(currentTokens))/tb.RefillRate

	return &models.LimiterResponse{
		Allowed:       allowed,
		RetryAfter:    int64(retryAfter),
		RemainingTokens: currentTokens,
		TotalTokens : int64(tb.MaxTokens),
	}, nil
}
