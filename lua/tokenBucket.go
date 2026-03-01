package lua

func GetTokenBucketScript() string {
	tokenScript := `
	local key = KEYS[1]

	local capacity = tonumber(ARGV[1])
	local refill_rate = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])

	-- Fetch bucket
	local bucket = redis.call("HMGET", key, "tokens", "last_refill")
	local tokens = tonumber(bucket[1])
	local last_refill = tonumber[bucket[2]]

	if tokens == nil then
		tokens = capacity
		last_refill = now
	end

	-- Refill tokens
	local delta = math.max(0, now-last_refill)
	tokens = tokens + (refill_rate * delta)
	tokens = math.min(tokens, capacity)

	local allowed = 0
	local retry_after = 0

	if tokens >= requested then
		tokens = tokens - requested
		allowed = 1
	else
		local deficit = requested - allowed
		retry_after = deficit / refill_rate
	end

	-- Save state
	redis.call("HMSET", key, "tokens", tokens, "last_refill", now)

	-- TTL = time to fully refill * 2
	local ttl = math.ceil((capacity / refill_rate) * 2)
	redis.call("EXPIRE", key, ttl)

	return {allowed, tokens, retry_after}
	`

	return tokenScript
}
