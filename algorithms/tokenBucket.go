package algorithms

import "fmt"

type TokenBucket struct {
	maxTokens       int
	tokensAvailable []string
	tokensQueue     []string
}

func NewTokenBucket(maxTokens int) *TokenBucket {
	return &TokenBucket{
		maxTokens:       maxTokens,
		tokensAvailable: make([]string, maxTokens),
		tokensQueue:     make([]string, 0),
	}
}

func (tb *TokenBucket) AddTokens() {
	// this function is going to run periodically
	if tb.IsBucketFull() {
		fmt.Println("Bucket is full cannot add tokens")
		return
	}

}

func (tb *TokenBucket) IsBucketFull() bool {
	return len(tb.tokensAvailable) == tb.maxTokens
}

func (tb *TokenBucket) IsBucketEmpty() bool {
	return len(tb.tokensAvailable) == 0
}
