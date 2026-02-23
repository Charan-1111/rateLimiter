package algorithms

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	window      time.Duration
	capacity    int
	currentCnt  int
	previousCnt int
	windowStart time.Time
	mu          sync.Mutex
}

func NewSlidingWindowCounter(window time.Duration, capacity int) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		window:      window,
		capacity:    capacity,
		windowStart: time.Now(),
	}
}

func (sc *SlidingWindowCounter) Allow(ctx context.Context, key string) (bool, error) {
	now := time.Now()

	sc.mu.Lock()
	defer sc.mu.Unlock()

	elapsed := now.Sub(sc.windowStart)

	// shift window if needed
	if elapsed >= sc.window {
		shift := int(elapsed / sc.window)

		if shift >= 2 {
			sc.previousCnt = 0
		} else {
			sc.previousCnt = sc.currentCnt
		}
		sc.currentCnt = 0
		sc.windowStart = now.Truncate(sc.window)
		elapsed = now.Sub(sc.windowStart)
	}

	// weighted count
	weight := float64(sc.window-elapsed) / float64(sc.window)
	effectiveCnt := float64(sc.currentCnt) + float64(sc.previousCnt)*weight

	if effectiveCnt >= float64(sc.capacity) {
		fmt.Println("Requests rejected")
		return false, nil
	}

	sc.currentCnt++
	fmt.Println("Request accepted")
	return true, nil
}
