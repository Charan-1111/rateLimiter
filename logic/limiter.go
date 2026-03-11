package logic

import (
	"context"
	"fmt"
	"goapp/algorithms"
	"goapp/services"

	"github.com/redis/go-redis/v9"
)

func GetLimiter(rdb *redis.Client, limiterFactory algorithms.LimiterFactory, cache *services.Cache, scope, identifier string) {
	limiter, err := limiterFactory.GetLimiter(scope, identifier, cache)
	if err != nil {
		fmt.Println("Error getting the limiter : ", err)
	}

	limiter.Allow(context.Background(), rdb, "tenant1", "user1")
}
