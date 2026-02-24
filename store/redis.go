package store

import (
	"goapp/models"

	"github.com/redis/go-redis/v9"
)

var Rdb *redis.Client

func InitRedis(redisDetails *models.RedisConfig) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr: redisDetails.Host + ":" + redisDetails.Port,
	})

	return rdb
}
