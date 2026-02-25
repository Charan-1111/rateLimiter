package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/store"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

type FixedCounterStore struct {
	WindowIndex int64
	Allowed     int64
}

type FixedCounter struct {
	window   time.Duration
	capacity int64
}

func GetNewFixedWindowCounter(window time.Duration, capacity int64) *FixedCounter {
	return &FixedCounter{
		window:   window,
		capacity: capacity,
	}
}

func (fc *FixedCounter) Allow(ctx context.Context, tenandId, userId string) (bool, error) {
	tokens := &FixedCounterStore{}
	// reset if the window is crossed
	now := time.Now()

	currentWindowIndex := now.UnixNano() / int64(fc.window)

	// read data from the redis
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmFixedWindow, tenandId, userId)

	val, err := store.Rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		tokens.WindowIndex = currentWindowIndex
		tokens.Allowed = fc.capacity
	} else if err != nil {
		fmt.Println("Error reading data from redis : ", err)
		return false, err
	} else {
		// unmarshal the data
		err := sonic.Unmarshal([]byte(val), tokens)
		if err != nil {
			fmt.Println("Error unmarshalling data from redis : ", err)
			return false, err
		}
	}

	if currentWindowIndex > tokens.WindowIndex {
		tokens.WindowIndex = currentWindowIndex
		tokens.Allowed = fc.capacity
	}

	if tokens.Allowed == 0 {
		fmt.Println("Request is denied")
		return false, nil
	}

	tokens.Allowed -= 1
	fmt.Println("Request is allowd")

	// Store the data into redis
	marshedVal, err := sonic.Marshal(tokens)
	if err != nil {
		fmt.Println("Error marshalling data to redis : ", err)
		return false, err
	}

	store.Rdb.Set(ctx, redisKey, marshedVal, fc.window)
	return true, nil
}
