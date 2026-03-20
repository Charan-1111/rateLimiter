package constants

import "time"

const (
	// ratelimit keys
	KeyRateLimit     = "rate_limit"
	KeyRateLimitType = "type"
	KeyAlgo          = "algo"
	KeyTenantId      = "tenantId"
	KeyUserId        = "userId"
	KeyScope         = "scope"
	KeyIdentifier    = "identifier"

	// Values
	ValeTypeMemory = "memory"
	ValueTypeRedis = "redis"

	// Algorithms
	AlgorithmTokenBucket   = "token_bucket"
	AlgorithmLeakyBucket   = "leaky_bucket"
	AlgorithmFixedWindow   = "fixed_window"
	AlgorithmSlidingWindow = "sliding_window"

	// Timeouts
	ContextTimeout               = 5 * time.Second
	CircuitBreakerInterval       = 10 * time.Second
	CircuitBreakerTimeout        = 30 * time.Second
	ConsecutiveFailuresThreshold = 5

	// Cache details
	PolicyCacheDuration = 5 * time.Minute
)
