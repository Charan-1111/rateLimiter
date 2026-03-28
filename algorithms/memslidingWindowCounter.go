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

type SlidingWindowStore struct {
	currentCnt  int
	previousCnt int
	windowStart time.Time
	mu          sync.Mutex
}

type SlidingWindow struct {
	capacity int
	window   time.Duration
	tokens   sync.Map
}

func NewSlidingWindowMem(windowStr string, capacity int, log zerolog.Logger) *SlidingWindow {
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		log.Error().Err(err).Msg("Error parsing the duration")
	}

	return &SlidingWindow{
		capacity: capacity,
		window:   window,
		tokens:   sync.Map{},
	}
}

func (sw *SlidingWindow) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	key := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmSlidingWindow, scope, identifier)
	now := time.Now()

	// fetch the data from the cache
	val, ok := sw.tokens.Load(key)
	if !ok {
		// allocate and initialize new sliding window state
		val, _ = sw.tokens.LoadOrStore(key, &SlidingWindowStore{
			windowStart: now,
			currentCnt:  0,
			previousCnt: 0,
		})
	}

	tokens := val.(*SlidingWindowStore)
	tokens.mu.Lock()
	defer tokens.mu.Unlock()

	elapsed := now.Sub(tokens.windowStart)

	if elapsed >= sw.window {
		shift := int(elapsed / sw.window)

		if shift >= 2 {
			tokens.previousCnt = 0
		} else {
			tokens.previousCnt = tokens.currentCnt
		}

		tokens.currentCnt = 0
		tokens.windowStart = tokens.windowStart.Add(time.Duration(shift) * sw.window)
		elapsed = now.Sub(tokens.windowStart)
	}

	// weightage is the percentage of the current window that has elapsed
	weight := float64(sw.window-elapsed) / float64(sw.window)
	effectiveCnt := float64(tokens.currentCnt) + weight*float64(tokens.previousCnt)

	if effectiveCnt >= float64(sw.capacity) {
		log.Warn().Str("scope", scope).Msg("Request is rejected, threshold exceeded")
		return &models.LimiterResponse{
			Allowed:         false,
			RetryAfter:      0,
			RemainingTokens: int64(tokens.currentCnt),
			TotalTokens:     int64(sw.capacity),
		}, nil
	}

	tokens.currentCnt += 1
	log.Debug().Str("scope", scope).Msg("Request is allowed")

	return &models.LimiterResponse{
		Allowed:         true,
		RetryAfter:      0,
		RemainingTokens: int64(tokens.currentCnt),
		TotalTokens:     int64(sw.capacity),
	}, nil
}
