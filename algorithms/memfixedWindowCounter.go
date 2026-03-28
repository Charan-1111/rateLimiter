package algorithms

import (
	"context"
	"goapp/constants"
	"goapp/models"
	"goapp/services"
	"goapp/utils"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type FixedWindowStore struct {
	windowIndex int
	tokens      int
	mu          sync.Mutex
}

type FixedWindow struct {
	capacity int
	window   time.Duration
	tokens   sync.Map
}

func NewFixedWindowMem(windowStr string, capacity int, log zerolog.Logger) *FixedWindow {
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing the duration")
	}

	return &FixedWindow{
		capacity: capacity,
		window:   window,
		tokens:   sync.Map{},
	}
}

func (fw *FixedWindow) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	key := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmFixedWindow, scope, identifier)
	now := time.Now()

	currentWindowIdx := int(now.UnixNano() / int64(fw.window))

	// fetch the data from the cache
	val, ok := fw.tokens.Load(key)
	if !ok {
		// initialize a new window store
		val, _ = fw.tokens.LoadOrStore(key, &FixedWindowStore{
			windowIndex: currentWindowIdx,
			tokens:      fw.capacity,
		})
	}

	tokenStore := val.(*FixedWindowStore)
	tokenStore.mu.Lock()
	defer tokenStore.mu.Unlock()

	if currentWindowIdx > tokenStore.windowIndex {
		tokenStore.windowIndex = currentWindowIdx
		tokenStore.tokens = fw.capacity
	}

	if tokenStore.tokens <= 0 {
		log.Warn().Str("scope", scope).Msg("Request is rejected, bucket empty")
		return &models.LimiterResponse{
			Allowed:         false,
			RetryAfter:      0,
			RemainingTokens: int64(tokenStore.tokens),
			TotalTokens:     int64(fw.capacity),
		}, nil
	}

	tokenStore.tokens -= 1
	log.Debug().Str("scope", scope).Msg("Request is allowed")

	return &models.LimiterResponse{
		Allowed:         true,
		RetryAfter:      0,
		RemainingTokens: int64(tokenStore.tokens),
		TotalTokens:     int64(fw.capacity),
	}, nil
}
