package algorithms

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SlidingWindowLog struct {
	window   time.Duration
	capacity int64
	logs     []time.Time
	mu       sync.Mutex
}

func GetNewSlidingWindowLog(window time.Duration, capacity int64) *SlidingWindowLog {
	return &SlidingWindowLog{
		window:   window,
		capacity: capacity,
		logs:     make([]time.Time, 0),
	}
}

func (sl *SlidingWindowLog) Allow(ctx context.Context, key string) (bool, error) {
	sl.mu.Lock()
	defer sl.mu.Unlock()

	now := time.Now()
	threshold := now.Add(-sl.window)

	// 1. Drop timestamps older than the window
	dropCnt := 0
	for _, t := range sl.logs {
		if t.After(threshold) {
			break
		}
		dropCnt++
	}

	// 2. Remove the old timestamps
	if dropCnt > 0 {
		sl.logs = sl.logs[dropCnt:]
	}

	// 3. Accept the request if the current log count is less than capacity
	if int64(len(sl.logs)) < sl.capacity {
		sl.logs = append(sl.logs, now)
		fmt.Println("Request is accepted") // You might want to remove print statements in production!
		return true, nil
	}
	fmt.Println("Request is rejected")
	return false, nil
}
