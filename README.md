# rateLimiter

![Build Status](https://img.shields.io/badge/build-passing-brightgreen) ![Version](https://img.shields.io/badge/version-0.1.0-blue) ![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)

A production-grade, high-performance rate limiting service written in Go, designed for both low-latency in-memory execution and distributed environments using Redis.

## What the project does

`rateLimiter` is a high-performance HTTP service that provides rate limiting capabilities. It exposes an API endpoint to evaluate limits based on dynamic policies. It features multiple rate limiting algorithms (like token bucket and sliding window) which can execute entirely in-memory for minimal latency, or use Redis with Lua scripts for distributed environments. The application utilizes the Fiber web framework, PostgreSQL for policy configuration, and provides a customizable plugin-based architecture.

## Why the project is useful

Building reliable systems requires effective traffic control. `rateLimiter` is useful because it offers:

- **Multiple Algorithms**: Supports Token Bucket, Leaky Bucket, Fixed Window Counter, and Sliding Window Counter out of the box.
- **Pluggable Storage Backends**: Choose between ultra-fast in-memory processing or robust distributed coordination via Redis.
- **Dynamic Configuration**: Rate limiting policies are effectively managed through a PostgreSQL database and internally cached, allowing limits to be updated seamlessly.
- **Resiliency Patterns**: Integrated with `gobreaker` for circuit breaking to protect related services and storage calls.

## How users can get started

### Prerequisites

- [Go](https://golang.org/) 1.25 or greater
- [Redis](https://redis.io/) (for the distributed limiting backend)
- [PostgreSQL](https://www.postgresql.org/) (for storing policies)

### Installation

1. **Clone the repository:**
   ```bash
   git clone https://github.com/your-org/rateLimiter.git
   cd rateLimiter
   ```

2. **Configuration:**
   Update the `deploy/config.json` file with your database, Redis, and server details. Ensure the `rateLimitPolicies` table exists in your Postgres database:
   ```json
   {
     "ports": { "fiberServer": ":8000" },
     "database": {
       "username": "<databaseUsername>",
       "password": "<databasePassword>",
       "host": "<databaseHost>",
       "port": "<databasePort>",
       "databaseName": "<databaseName>"
     },
     "redis": { "host": "<redisHost>", "port": "<redisPort>" }
   }
   ```

3. **Build and Run:**
   ```bash
   go mod tidy
   go build -o ratelimiter main.go
   ./ratelimiter
   ```

### Usage Example

The server exposes a main rate-limiting validation endpoint `/api/v1/limiter`. You can query it using `curl`:

```bash
curl "http://localhost:8000/api/v1/limiter?scope=api&identifier=user_123&type=memory"
```

*Note: Ensure you have populated your `rateLimitPolicies` table in Postgres for the given `scope` and `identifier` prior to making rate-limiting requests.*

## Failure Handling

To ensure high availability and stability, `rateLimiter` implements the Circuit Breaker pattern using [`gobreaker`](https://github.com/sony/gobreaker).

This mechanism protects the application and its downstream dependencies (like Redis or PostgreSQL) from cascading failures. When a dependency experiences consecutive failures beyond a configured threshold (`constants.ConsecutiveFailuresThreshold`), the Circuit Breaker "trips" and opens. 
- While **Open**, requests are failed fast without overwhelming the struggling service.
- After a configured `Timeout` (`constants.CircuitBreakerTimeout`), the circuit transitions to a **Half-Open** state, allowing a limited number of test requests (`MaxRequests: 2`) to pass through. 
- If the test requests succeed, the circuit closes and normal traffic resumes; if they fail, the circuit re-opens.

## Failure Scenarios

- Redis down → fallback to in-memory (eventual inconsistency)
- PostgreSQL down → cached policies still work
- Cache miss storm → controlled via Ristretto
- Circuit breaker open → fail fast

## Where users can get help

- Detailed algorithm implementations can be found in the [`algorithms/`](algorithms/) directory.
- Database schemas and queries are specified in [`deploy/config.json`](deploy/config.json).
- Refer to the performance benchmarks in the [`benchmarkReports/`](benchmarkReports/) folder for memory vs Redis latency and throughput comparisons.

If you encounter issues or have questions, please check our Issue Tracker or open a new issue.

## Trade-offs and Architecture

![Architecture Diagram](documentation/Architecture.png)

When configuring and using `rateLimiter`, consider the following architectural choices based on the algorithms and storage backends:

- **Accuracy vs Memory**: Storing exact timestamps for every request (e.g., Sliding Window Log) provides 100% accuracy but consumes significantly more memory. Counter-based approaches (e.g., Sliding Window Counter or Token Bucket) approximate the rate and are highly memory-efficient, making them better suited for high-throughput scenarios.
- **Latency vs Consistency (Algorithms)**: Using the local in-memory store for algorithms offers ultra-low, sub-millisecond latency but sacrifices strict global consistency in a distributed, multi-instance deployment. Conversely, using Redis algorithms ensures strict global consistency across all application instances but introduces network latency for every rate-limit evaluation.
- **Policy Engine Caching**: To dynamically fetch rate limit configurations without hammering the PostgreSQL database, `rateLimiter` aggressively caches the database `PolicySchemas` locally in memory using [**Dgraph's Ristretto**](https://github.com/dgraph-io/ristretto). *Note: Ristretto is strictly used to cache the database rules, it does not store the algorithmic request counters.* Read our deep-dive on [Why We Chose Ristretto](documentation/ristretto.md) for more details on resolving lock-contention, TTL management, and TinyLFU eviction rules!

## Performance Benchmarks

The following benchmarks illustrate the Max RPS and latency parameters of the different rate limiting algorithms when stored purely in-memory vs using Redis. **These benchmarks were conducted with 10,000 total requests and 100 concurrent users.** For more details, refer to the raw benchmarks in the [`benchmarkReports/`](benchmarkReports/) folder.

### Fixed Window Counter

| Backend      | Max RPS | P95 Latency | P99 Latency |
|--------------|---------|-------------|-------------|
| In-Memory    | 3,518   | 78.1 ms     | 173.3 ms    |
| Redis        | 1,011   | 232.5 ms    | 352.9 ms    |

### Leaky Bucket

| Backend      | Max RPS | P95 Latency | P99 Latency |
|--------------|---------|-------------|-------------|
| In-Memory    | 22,328  | 10.6 ms     | 55.4 ms     |
| Redis        | 1,376   | 182.2 ms    | 245.0 ms    |

### Sliding Window Counter

| Backend      | Max RPS | P95 Latency | P99 Latency |
|--------------|---------|-------------|-------------|
| In-Memory    | 970     | 216.3 ms    | 342.4 ms    |
| Redis        | 777     | 331.4 ms    | 474.5 ms    |

### Token Bucket

| Backend      | Max RPS | P95 Latency | P99 Latency |
|--------------|---------|-------------|-------------|
| In-Memory    | 4,029   | 68.6 ms     | 130.0 ms    |
| Redis        | 930     | 271.5 ms    | 432.8 ms    |

## Future Work

See the [Future Work](documentation/futureWork.md) document for a list of planned features and enhancements.

## Who maintains and contributes

This project is maintained by **Leela Guru Charan Avvaru** (charanavvaru11@gmail.com) and our open-source community.

We welcome contributions! Please review our [Contributing Guidelines](documentation/CONTRIBUTING.md) to learn how to propose bugfixes and improvements, and how to format your code before submitting a Pull Request.
