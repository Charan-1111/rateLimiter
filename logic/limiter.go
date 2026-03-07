package logic

import (
	"context"
	"fmt"
	"goapp/algorithms"
)

func GetLimiter(limiterFactory algorithms.LimiterFactory, limiterType, algorithm string) {
	limiter, err := limiterFactory.GetLimiter(limiterType, algorithm)
	if err != nil {
		fmt.Println("Error getting the limiter : ", err)
	}

	limiter.Allow(context.Background(), "tenant1", "user1")
}
