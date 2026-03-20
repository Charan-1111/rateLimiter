package lua

func GetLeakyBucketScript() string {
	script := `
	local key = KEYS[1]

	local capacity = tonumber(ARGV[1])
	local leak_rate = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])

	-- Fetch bucket
	local bucket = redis.call("HMGET", key, "tokens", "last_leak")
	local tokens = tonumber(bucket[1])
	local last_leak = tonumber(bucket[2])

	if tokens == nil then
		tokens = 0
		last_leak = now
	end

	-- Leak the bucket
	local delta = math.max(0, now-last_leak)
	tokens = tokens - (delta * leak_rate)
	tokens = math.max(tokens, 0)
	last_leak = now

	local allowed = 0
	local retry_after

	if tokens + requested > capacity then
		-- reject the requests at this stage
		allowed = 0
	else
		tokens = tokens + requested
		allowed = 1
	end

	-- Save the state
	redis.call("HMSET", key, "tokens", tokens, "last_leak", now)

	-- TTL = time to fully leak * 2
	local ttl = math.ceil((capacity / leak_rate) * 2)
	redis.call("EXPIRE", key, ttl)

	return {allowed, tokens}
	`

	return script
}
