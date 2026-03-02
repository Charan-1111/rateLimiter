package lua

func GetFixedWindowCounterScript() string {
	script := `
	local key = KEYS[1]

	local capacity = tonumber(ARGV[1])
	local window = tonumber(ARGV[2])
	local now = tonumber(ARGV[3])
	local requested = tonumber(ARGV[4])

	-- Fetch the data
	local result = redis.call("HMGET", key, "tokens", "windowIndex")

	local tokens = tonumber(result[1])
	local windowIndex = tonumber(result[2])

	local currentWindowIndex = math.ceil(now / window)

	if tokens == nil or currentWindowIndex > windowIndex then
		tokens = capacity
		windowIndex = currentWindowIndex
	end

	local allowed = 0

	if tokens - requested >= 0 then
		allowed = 1
		tokens = tokens - requested
	end

	-- Store the values
	redis.call("HMSET", key, "tokens", tokens, "windowIndex", windowIndex)

	-- SET TTL ( PEXPIRE is crucial as it accepts the time in milliseconds)
	redis.call("PEXPIRE", key, window)

	return allowed, tokens
	`

	return script
}
