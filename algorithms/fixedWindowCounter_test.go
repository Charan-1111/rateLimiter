package algorithms

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestFixedWindowCounter_AllowWithinCapacity(t *testing.T) {
	window := 100 * time.Millisecond
	capacity := int64(3)
	fc := GetNewFixedWindowCounter(window, capacity)
	ctx := context.Background()

	for i := int64(0); i < capacity; i++ {
		allowed, err := fc.Allow(ctx, "test-user")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !allowed {
			t.Errorf("expected request %d to be allowed", i+1)
		}
	}
}

func TestFixedWindowCounter_DenyOverCapacity(t *testing.T) {
	window := 1 * time.Second
	capacity := int64(2)
	fc := GetNewFixedWindowCounter(window, capacity)
	ctx := context.Background()

	// Exhaust capacity
	for i := int64(0); i < capacity; i++ {
		fc.Allow(ctx, "test-user")
	}

	// This one should be denied
	allowed, err := fc.Allow(ctx, "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if allowed {
		t.Error("expected request to be denied after exceeding capacity")
	}
}

func TestFixedWindowCounter_ResetOnNewWindow(t *testing.T) {
	window := 50 * time.Millisecond
	capacity := int64(1)
	fc := GetNewFixedWindowCounter(window, capacity)
	ctx := context.Background()

	// Use up the capacity for current window
	fc.Allow(ctx, "test-user")

	// Ensure next immediately fails
	allowed, _ := fc.Allow(ctx, "test-user")
	if allowed {
		t.Fatal("expected request to be denied initially")
	}

	// Wait for the window to roll over
	time.Sleep(window + 10*time.Millisecond)

	// Now it should be allowed again
	allowed, err := fc.Allow(ctx, "test-user")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !allowed {
		t.Error("expected request to be allowed after new window started")
	}
}

func TestFixedWindowCounter_Concurrency(t *testing.T) {
	window := 1 * time.Second
	capacity := int64(100)
	fc := GetNewFixedWindowCounter(window, capacity)
	ctx := context.Background()

	var wg sync.WaitGroup
	var allowedCount int64
	var mu sync.Mutex

	// Fire 200 concurrent requests
	numRequests := 200
	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			allowed, _ := fc.Allow(ctx, "test-user")
			if allowed {
				mu.Lock()
				allowedCount++
				mu.Unlock()
			}
		}()
	}

	wg.Wait()

	// Even though we fired 200 requests, exactly `capacity` (100) should be allowed within this window.
	if allowedCount != capacity {
		t.Errorf("expected exactly %d allowed requests, got %d", capacity, allowedCount)
	}
}
