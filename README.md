# rateLimiter

![Build Status](https://img.shields.io/badge/build-passing-brightgreen) ![Version](https://img.shields.io/badge/version-0.1.0-blue) ![Go Version](https://img.shields.io/badge/go-1.25+-blue.svg)

A flexible, scalable, and extensible rate limiting service written in Go.

## What the project does

`rateLimiter` is a high-performance HTTP service that provides rate limiting capabilities. It exposes an API endpoint to evaluate limits based on dynamic policies. It features multiple rate limiting algorithms (like token bucket and sliding window) which can execute entirely in-memory for minimal latency, or use Redis with Lua scripts for distributed environments. The application utilizes the Fiber web framework, PostgreSQL for policy configuration, and provides a customizable plugin-based architecture.

## Why the project is useful

Building reliable systems requires effective traffic control. `rateLimiter` is useful because it offers:

- **Multiple Algorithms**: Supports Token Bucket, Leaky Bucket, Fixed Window Counter, and Sliding Window Counter out of the box.
- **Pluggable Storage Backends**: Choose between ultra-fast in-memory processing or robust distributed coordination via Redis.
- **Dynamic Configuration  **: Rate limiting policies are effectively managed through a PostgreSQL database and internally cached, allowing limits to be updated seamlessly.
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

## Where users can get help

- Detailed algorithm implementations can be found in the [`algorithms/`](algorithms/) directory.
- Database schemas and queries are specified in [`deploy/config.json`](deploy/config.json).
- Refer to the performance benchmarks in the [`benchmarkReports/`](benchmarkReports/) folder for memory vs Redis latency and throughput comparisons.

If you encounter issues or have questions, please check our Issue Tracker or open a new issue.

## Trade-offs

When configuring and using `rateLimiter`, consider the following trade-offs based on the chosen algorithms and storage backends:

- **Accuracy vs Memory**: Storing exact timestamps for every request (e.g., Sliding Window Log) provides 100% accuracy but consumes significantly more memory. Counter-based approaches (e.g., Sliding Window Counter or Token Bucket) approximate the rate and are highly memory-efficient, making them better suited for high-throughput scenarios.
- **Latency vs Consistency**: Using the local in-memory store offers ultra-low, sub-millisecond latency but sacrifices strict global consistency in a distributed, multi-instance deployment. Conversely, using Redis ensures strict global consistency across all application instances but introduces network latency for every rate-limit evaluation.

## Future Work

See the [Future Work](documentation/futureWork.md) document for a list of planned features and enhancements.

## Who maintains and contributes

This project is maintained by **Leela Guru Charan Avvaru** (charanavvaru11@gmail.com) and our open-source community.

We welcome contributions! Please review our [Contributing Guidelines](documentation/CONTRIBUTING.md) to learn how to propose bugfixes and improvements, and how to format your code before submitting a Pull Request.
