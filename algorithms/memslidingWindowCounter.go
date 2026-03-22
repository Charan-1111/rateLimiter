package algorithms

import (
	"context"
	"fmt"
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
}

type SlidingWindow struct {
	capacity int
	window   time.Duration
	tokens   map[string]*SlidingWindowStore
	mu       sync.Mutex
}

func NewSlidingWindowMem(windowStr string, capacity int) *SlidingWindow {
	window, err := time.ParseDuration(windowStr)
	if err != nil {
		fmt.Println("Error parsing the duration")
	}

	return &SlidingWindow{
		capacity: capacity,
		window:   window,
		tokens:   make(map[string]*SlidingWindowStore),
		mu:       sync.Mutex{},
	}
}

func (sw *SlidingWindow) Allow(ctx context.Context, rdb *redis.Client, cb *services.CircuitBreaker, log zerolog.Logger, scope, identifier string) (*models.LimiterResponse, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	key := utils.StringBuilder(constants.KeyRateLimit, constants.AlgorithmSlidingWindow, scope, identifier)
	now := time.Now()

	// fetch the data from the cache
	tokens, ok := sw.tokens[key]
	if !ok {
		// allocate and initialize new sliding window state
		tokens = &SlidingWindowStore{
			windowStart: now,
			currentCnt:  0,
			previousCnt: 0,
		}
	}

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
		fmt.Println("Request is rejected")
		return &models.LimiterResponse{
			Allowed:       false,
			RetryAfter:    0,
			CurrentTokens: int64(tokens.currentCnt),
		}, nil
	}

	tokens.currentCnt += 1
	fmt.Println("Request is allowed")

	// store the information in the cache
	sw.tokens[key] = tokens
	return &models.LimiterResponse{
		Allowed:       true,
		RetryAfter:    0,
		CurrentTokens: int64(tokens.currentCnt),
	}, nil
}
