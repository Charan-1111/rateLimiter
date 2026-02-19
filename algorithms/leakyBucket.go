package algorithms

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type LeakyBucket struct {
	maxTokens  float64
	tokens     float64
	leakRate   float64
	lastLeak   time.Time
	mu         sync.Mutex
}

func NewLeakyBucket(maxTokens, leakRate float64) *LeakyBucket {
	return &LeakyBucket{
		maxTokens:  maxTokens,
		leakRate:   leakRate,
		lastLeak:   time.Now(),
	}
}


func (lb *LeakyBucket) IsBucketFull() bool {
	return lb.tokens >= lb.maxTokens
}

func (lb *LeakyBucket) IsBucketEmpty() bool {
	return lb.tokens == 0
}

func (lb *LeakyBucket) Allow(ctx context.Context, key string) (bool, error) {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	lb.LeakTokens()

	if lb.IsBucketFull() {
		return false, fmt.Errorf("bucket is full, request is getting denied")
	}

	fmt.Println("bucket is not full, request is getting proceed")

	lb.tokens += 1

	return true, nil
}

func (lb *LeakyBucket) LeakTokens() {
	lb.mu.Lock()
	defer lb.mu.Unlock()

	now := time.Now()

	lb.tokens = lb.tokens - (now.Sub(lb.lastLeak).Seconds())*lb.leakRate

	if lb.tokens < 0 {
		lb.tokens = 0
	}

	lb.lastLeak = now
}