package algorithms

import (
	"context"
	"testing"
	"time"
)

func TestLeakyBucket_Allow(t *testing.T) {
	// Capacity 5, leaks 1 token per second
	lb := NewLeakyBucket(5, 1)

	ctx := context.Background()

	// Fill the bucket
	for i := 0; i < 5; i++ {
		allowed, err := lb.Allow(ctx, "test")
		if !allowed || err != nil {
			t.Fatalf("Expected request %d to be allowed, got error: %v", i+1, err)
		}
	}

	// Next request should be denied (bucket full)
	allowed, err := lb.Allow(ctx, "test")
	if allowed || err == nil {
		t.Error("Expected request to be denied (bucket full), but it was allowed")
	}

	// Wait for 2 seconds (should leak 2 tokens) - theoretical check
	// Since inside Allow call it uses time.Now(), we need to sleep
	time.Sleep(2100 * time.Millisecond)

	// Should allow 2 more requests
	allowed, err = lb.Allow(ctx, "test")
	if !allowed || err != nil {
		t.Errorf("Expected request after waiting to be allowed")
	}

	allowed, err = lb.Allow(ctx, "test")
	if !allowed || err != nil {
		t.Errorf("Expected second request after waiting to be allowed")
	}

	// Should be full again (5 initial - 2 leaked + 2 new = 5)
	allowed, err = lb.Allow(ctx, "test")
	if allowed || err == nil {
		t.Error("Expected request to be denied after filling again")
	}
}

func TestLeakyBucket_Leak(t *testing.T) {
	lb := NewLeakyBucket(10, 100) // Leaks 100 per second

	ctx := context.Background()
	lb.Allow(ctx, "test") // 1 token

	time.Sleep(50 * time.Millisecond) // leaks ~5 tokens (but we only have 1)

	// Should be empty now
	lb.mu.Lock()
	lb.LeakTokens() // internal update
	if lb.tokens != 0 {
		t.Errorf("Expected bucket to be empty (0 tokens), got %f", lb.tokens)
	}
	lb.mu.Unlock()
}
