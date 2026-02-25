package algorithms

import (
	"context"
	"testing"
	"time"
)

func TestTokenBucket_FractionalTokens(t *testing.T) {
	// Setup: Max 10 tokens, refill rate 1/sec.
	// Initial state: Full (10 tokens).
	tb := NewTokenBucket(10, 1)

	// Consume all 10 tokens.
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		allowed, err := tb.Allow(ctx, "test", "test")
		if !allowed || err != nil {
			t.Fatalf("Iteration %d: Expected allowed=true, got %v, %v", i, allowed, err)
		}
	}

	// Now bucket should be empty (or close to 0).
	// Wait 0.5 seconds. Should have ~0.5 tokens.
	time.Sleep(500 * time.Millisecond)

	// Try to consume 1 token. Should FAIL because we only have 0.5.
	allowed, _ := tb.Allow(ctx, "test", "test")
	if allowed {
		t.Errorf("Fractional check failed: Allowed request with < 1 token available. Current tokens: %f", "")
	} else {
		t.Log("Correctly denied request with partial token.")
	}
}
