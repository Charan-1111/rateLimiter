package constants

const (
	// ratelimit keys
	KeyRateLimit = "rate_limit"
	KeyRateLimitType = "type"
	KeyAlgo = "algo"
	KeyTenantId = "tenantId"
	KeyUserId = "userId"
	KeyScope = "scope"
	KeyIdentifier = "identifier"

	// Values
	ValeTypeMemory = "memory"
	ValueTypeRedis = "redis"
	

	// Algorithms
	AlgorithmTokenBucket = "token_bucket"
	AlgorithmLeakyBucket = "leaky_bucket"
	AlgorithmFixedWindow = "fixed_window"
	AlgorithmSlidingWindow = "sliding_window"
)