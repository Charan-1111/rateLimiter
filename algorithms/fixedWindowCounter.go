package algorithms

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type FixedCounter struct {
	windowIndex int64
	window      time.Duration
	capacity    int64
	allowed     int64
	mu          sync.Mutex
}

func GetNewFixedWindowCounter(window time.Duration, capacity int64) *FixedCounter {
	return &FixedCounter{
		windowIndex: time.Now().UnixNano() / int64(window),
		window:      window,
		capacity:    capacity,
		allowed:     capacity,
	}
}

func (fc *FixedCounter) Allow(ctx context.Context, key string) (bool, error) {
	// reset if the window is crossed
	now := time.Now()

	currentWindowIndex := now.UnixNano() / int64(fc.window)

	fc.mu.Lock()
	defer fc.mu.Unlock()

	if currentWindowIndex > fc.windowIndex {
		fc.windowIndex = currentWindowIndex
		fc.allowed = fc.capacity
	}

	if fc.allowed == 0 {
		fmt.Println("Request is denied")
		return false, nil
	}

	fc.allowed -= 1
	fmt.Println("Request is allowd")
	return true, nil
}
