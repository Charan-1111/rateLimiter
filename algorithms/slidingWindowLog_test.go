package algorithms

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestSlidingWindowLog_Basic(t *testing.T) {
	window := 100 * time.Millisecond
	capacity := int64(3)
	limiter := GetNewSlidingWindowLog(window, capacity)

	ctx := context.Background()
	key := "user1"

	// Should allow exactly 3 requests
	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, key)
		if err != nil {
			t.Fatalf("Unexpected error: %v", err)
		}
		if !allowed {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}

	// 4th request should be rejected
	allowed, err := limiter.Allow(ctx, key)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}
	if allowed {
		t.Error("Expected 4th request to be rejected, but it was allowed")
	}
}

func TestSlidingWindowLog_Sliding(t *testing.T) {
	window := 100 * time.Millisecond
	capacity := int64(2)
	limiter := GetNewSlidingWindowLog(window, capacity)

	ctx := context.Background()
	key := "user2"

	// Send 2 requests, both should be allowed
	limiter.Allow(ctx, key)
	limiter.Allow(ctx, key)

	// 3rd request should be rejected
	allowed, _ := limiter.Allow(ctx, key)
	if allowed {
		t.Error("Expected 3rd request to be rejected")
	}

	// Wait for the window to pass
	time.Sleep(150 * time.Millisecond)

	// Now a new request should be allowed again
	allowed, _ = limiter.Allow(ctx, key)
	if !allowed {
		t.Error("Expected request to be allowed after window passed")
	}
}

func TestSlidingWindowLog_Concurrent(t *testing.T) {
	window := 200 * time.Millisecond
	capacity := int64(50)
	limiter := GetNewSlidingWindowLog(window, capacity)

	ctx := context.Background()
	key := "user3"

	var wg sync.WaitGroup
	var allowedCount int64
	var mu sync.Mutex

	// Fire 100 concurrent requests
	totalRequests := 100
	for i := 0; i < totalRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _ := limiter.Allow(ctx, key)
			if allowed {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	if allowedCount != capacity {
		t.Errorf("Expected exactly %d requests to be allowed, but got %d", capacity, allowedCount)
	}
}
