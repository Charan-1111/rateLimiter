package server

import (
	"goapp/store"
	"goapp/utils"

	"github.com/redis/go-redis/v9"
)

type Application struct {
	config *utils.Config
	rdb    *redis.Client
}

func NewApplication(filePath string) (*Application, error) {
	// Load the configuration file
	config := &utils.Config{}
	if err := config.LoadConfig(filePath); err != nil {
		return nil, err
	}

	// Initialize Redis
	rdb := store.InitRedis(&config.Redis)

	return &Application{
		config: config,
		rdb:    rdb,
	}, nil
}
