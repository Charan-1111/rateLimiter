package algorithms

import (
	"context"
	"fmt"
	"goapp/constants"
	"goapp/store"
	"sync"
	"time"

	"github.com/bytedance/sonic"
	"github.com/redis/go-redis/v9"
)

type SlidingWindowStore struct {
	CurrentCnt  int
	PreviousCnt int
	WindowStart time.Time
}
type SlidingWindowCounter struct {
	window   time.Duration
	capacity int
	mu       sync.Mutex
}

func NewSlidingWindowCounter(window time.Duration, capacity int) *SlidingWindowCounter {
	return &SlidingWindowCounter{
		window:   window,
		capacity: capacity,
	}
}

func (sc *SlidingWindowCounter) Allow(ctx context.Context, tenantId , userId string) (bool, error) {
	now := time.Now()

	tokens := &SlidingWindowStore{}

	// read data from the redis
	redisKey := fmt.Sprintf("%s:%s:%s:%s", constants.KeyRateLimit, constants.AlgorithmSlidingWindow, tenantId, userId)

	val, err := store.Rdb.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		tokens.WindowStart = now.Truncate(sc.window)
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

	elapsed := now.Sub(tokens.WindowStart)

	// shift window if needed
	if elapsed >= sc.window {
		shift := int(elapsed / sc.window)

		if shift >= 2 {
			tokens.PreviousCnt = 0
		} else {
			tokens.PreviousCnt = tokens.CurrentCnt
		}
		tokens.CurrentCnt = 0
		tokens.WindowStart = now.Truncate(sc.window)
		elapsed = now.Sub(tokens.WindowStart)
	}

	// weighted count
	weight := float64(sc.window-elapsed) / float64(sc.window)
	effectiveCnt := float64(tokens.CurrentCnt) + float64(tokens.PreviousCnt)*weight

	if effectiveCnt >= float64(sc.capacity) {
		fmt.Println("Requests rejected")
		return false, nil
	}

	tokens.CurrentCnt++
	fmt.Println("Request accepted")

	// Store the data into redis
	marshedVal, err := sonic.Marshal(tokens)
	if err != nil {
		fmt.Println("Error marshalling data to redis : ", err)
		return false, err
	}

	store.Rdb.Set(ctx, redisKey, marshedVal, 2*sc.window)
	return true, nil
}
