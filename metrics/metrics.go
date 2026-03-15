package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
	Requests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limiter_total_requests",
			Help: "Total number of requests processed",
		},
		[]string{"status", "algorithm"},
	)

	RequestsLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "rate_limiter_request_duration_seconds",
			Help:    "Latency of rate limiter decisions",
			Buckets: prometheus.DefBuckets,
		},
	)

	RedisErrors = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "rate_limiter_redis_errors_total",
			Help: "Total Redis errors",
		},
	)

	RedisLatency = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "rate_limiter_redis_latency_seconds",
			Help:    "Redis operation latency",
			Buckets: prometheus.DefBuckets,
		},
	)
)
