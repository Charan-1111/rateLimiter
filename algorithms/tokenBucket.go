package algorithms

type TokenBucket struct {
	maxTokens int
	tokens []string
}

func NewTokenBucket(maxTokens int) *TokenBucket {
	return &TokenBucket{
		maxTokens: maxTokens,
		tokens: make([]string, maxTokens),
	}
}

func (tb *TokenBucket) AddTokens() {
	
}