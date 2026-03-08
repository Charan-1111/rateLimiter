# rateLimiter

![Build Status](https://img.shields.io/badge/build-passing-brightgreen) ![Version](https://img.shields.io/badge/version-0.1.0-blue) ![License](https://img.shields.io/badge/license-MIT-lightgrey)

A simple, extensible rate limiting service written in Go.  It supports multiple algorithms (token bucket, leaky bucket, fixed and sliding windows) and can operate either in‑memory or backed by Redis (with Lua scripts).

## 🚀 What the project does

`rateLimiter` exposes an HTTP endpoint that selects and exercises a rate limiting algorithm based on query parameters.  The core library defines a `RateLimiter` interface and a factory for creating implementations; the service layers wire up configuration, Redis, and HTTP handlers.

## 💡 Why it's useful

- **Flexible algorithms:** Token bucket, leaky bucket, fixed‑window and sliding‑window counters are all available.
- **Pluggable storage:** Choose in‑process memory for ultra‑low latency or Redis for distributed usage with atomic Lua scripts.
- **Simple API:** One endpoint lets you request a limiter by type and algorithm, making it easy to integrate in tests or as a library.
- **Educational:** Great reference for learning rate limiting techniques and how to combine Go, Fiber, and Redis.

## 🛠️ Getting started

### Prerequisites

- Go 1.20+ installed (`go env`)
- Redis server if you plan to use the Redis-backed limiters

### Clone the repo

```bash
git clone https://github.com/your-org/rateLimiter.git
cd rateLimiter
```

### Configuration

Edit `manifest/config.json` to point at your Redis host/port and set desired token parameters:

```json
{
  "ports": { "fiberServer": ":8080" },
  "redis": { "host": "localhost", "port": "6379" },
  "maxTokens": 10,
  "refillRate": 1
}
```

### Build and run

```bash
go build -o ratelimiter
./ratelimiter
```

The server will start on the configured port.

### Usage example

Query the `/api/v1/limiter` endpoint with `type` and `algo` parameters:

```bash
curl "http://localhost:8080/api/v1/limiter?type=memory&algo=token_bucket"
```

> The handler currently just invokes the limiter once and prints results to stdout; extend it for production use.

### Running benchmarks

Benchmark scripts under `benchmarkReports` show performance when using in‑memory versus Redis.  See `token_bucket_mem.txt` and `token_bucket_redis.txt` for sample output.

#### Sample results

##### In‑memory limiter (`token_bucket_mem.txt`)

```
Summary:
  Total:	2.4820 secs
  Slowest:	0.2402 secs
  Fastest:	0.0002 secs
  Average:	0.0238 secs
  Requests/sec:	4029.0574
  

... (truncated for brevity; full report is available in `benchmarkReports/token_bucket_mem.txt`)
```

##### Redis + Lua limiter (`token_bucket_redis.txt`)

```
Summary:
  Total:	12.6388 secs
  Slowest:	0.8464 secs
  Fastest:	0.0016 secs
  Average:	0.1238 secs
  Requests/sec:	791.2162
  

... (truncated for brevity; full report is available in `benchmarkReports/token_bucket_redis.txt`)
```

These numbers illustrate the performance gap between local and networked implementations; run the benchmarks yourself against your environment to validate.

#### High‑level comparison

| Algorithm       | Metric          | In‑Memory         | Redis + Lua       |
|----------------|-----------------|-------------------|-------------------|
| Token bucket   | Avg latency     | 23 ms             | 124 ms            |
|                | Requests/sec    | 4 029             | 791               |
| Leaky bucket   | Avg latency     | 4 ms              | 71 ms             |
|                | Requests/sec    | 22 329            | 1 377             |
| Fixed window   | Avg latency     | 28 ms             | 95 ms             |
|                | Requests/sec    | 3 518             | 1 012             |
| Sliding window | Avg latency     | 98 ms             | 125 ms            |
|                | Requests/sec    | 970               | 777               |

*Latency rounded to nearest millisecond; see individual reports in `benchmarkReports/` for full distributions.*

These figures make it clear that the in‑memory limiters consistently outperform their Redis‑backed counterparts, often by an order of magnitude.  Use the table as a quick reference when deciding which backend to deploy.

## 📚 Documentation & support

- See the `algorithms/` package for implementation details of each limiter.
- Lua scripts used with Redis are in `lua/`.
- Configuration loading is handled by `utils/`.

For questions or help, open an issue or check the project wiki (link to your repo).

## 🤝 Contributing

Contributions are welcome!  Please read [`CONTRIBUTING.md`](CONTRIBUTING.md) for guidelines on reporting issues, writing tests, and submitting PRs.

## 🧑‍💻 Maintainers

Maintained by **Leela Guru Charan Avvaru** (charanavvaru11@gmail.com) and the open‑source community.
