package logic

import (
	"context"
	"fmt"
	"goapp/algorithms"
	"goapp/services"
	"goapp/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

func GetLimiter(ctx context.Context, db *pgxpool.Pool, rdb *redis.Client, config *utils.Config, log zerolog.Logger, limiterFactory algorithms.LimiterFactory, cache *services.Cache, scope, identifier, rateLimitType string) {
	limiter, err := limiterFactory.GetLimiter(ctx, db, log, scope, identifier, rateLimitType, config.Queries.Fetch.FetchPolicyByKey, cache)
	if err != nil {
		fmt.Println("Error getting the limiter : ", err)
	}

	limiter.Allow(context.Background(), rdb, "tenant1", "user1")
}
