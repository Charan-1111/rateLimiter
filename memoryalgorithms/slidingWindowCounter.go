package memoryalgorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"sync"
	"time"
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

func (sw *SlidingWindow) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	sw.mu.Lock()
	defer sw.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmSlidingWindow, tenantId, userId)
	now := time.Now()

	// fetch the data from the cache
	tokens, ok := sw.tokens[key]
	if !ok {
		tokens.windowStart = now
		tokens.currentCnt = 0
		tokens.previousCnt = 0
	}

	elapsed := now.Sub(tokens.windowStart)

	if elapsed >= sw.window {
		shift := int(elapsed / sw.window)

		if shift >=2 {
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
		return false, nil
	}

	tokens.currentCnt += 1
	fmt.Println("Request is allowed")

	// store the information in the cache
	sw.tokens[key] = tokens
	return true, nil
}
