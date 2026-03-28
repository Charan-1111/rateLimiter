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
}

type FixedWindow struct {
	capacity int
	window   time.Duration
	tokens   map[string]*FixedWindowStore
	mu       sync.Mutex
}

func NewFixedWindowMem(windowStr string, capacity int, log zerolog.Logger) *FixedWindow {
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing the duration")
	}

	return &FixedWindow{
		capacity: capacity,
		window:   window,
		tokens:   make(map[string]*FixedWindowStore),
		mu:       sync.Mutex{},
	}
}

func (fw *FixedWindow) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	key := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmFixedWindow, scope, identifier)
	now := time.Now()

	currentWindowIdx := int(now.UnixNano() / int64(fw.window))

	// fetch the data from the cache

	tokenStore, ok := fw.tokens[key]
	if !ok {
		// initialize a new window store
		tokenStore = &FixedWindowStore{
			windowIndex: currentWindowIdx,
			tokens:      fw.capacity,
		}
	}

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

	// store the information in the cache
	fw.tokens[key] = tokenStore

	return &models.LimiterResponse{
		Allowed:         true,
		RetryAfter:      0,
		RemainingTokens: int64(tokenStore.tokens),
		TotalTokens:     int64(fw.capacity),
	}, nil
}
