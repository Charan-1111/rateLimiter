package memoryalgorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"sync"
	"time"
)

type FixedWindowStore struct {
	windowIndex int
	tokens      int
}

type FixedWindow struct {
	capacity int
	window   time.Duration
	tokens   map[string]*FixedWindowStore
	mu       sync.Mutex
}

func (fw *FixedWindow) Allow(ctx context.Context, tenantId, userId string) (bool, error) {
	fw.mu.Lock()
	defer fw.mu.Unlock()

	key := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmFixedWindow, tenantId, userId)
	now := time.Now()

	currentWindowIdx := int(now.UnixNano() / int64(fw.window))

	// fetch the data from the cache

	tokenStore, ok := fw.tokens[key]
	if !ok {
		tokenStore.windowIndex = currentWindowIdx
		tokenStore.tokens = fw.capacity
	}

	if currentWindowIdx > tokenStore.windowIndex {
		tokenStore.windowIndex = currentWindowIdx
		tokenStore.tokens = fw.capacity
	}

	if tokenStore.tokens <= 0 {
		fmt.Println("Request is rejected")
		return false, nil
	}

	tokenStore.tokens -= 1
	fmt.Println("Request is allowed")

	// store the information in the cache
	fw.tokens[key] = tokenStore

	return true, nil
}
