package handlers

import (
	"goapp/algorithms"
	"goapp/utils"

	"github.com/redis/go-redis/v9"
)

type ConfigHandler struct {
	config  *utils.Config
	rdb     *redis.Client
	factory algorithms.LimiterFactory
}

func NewConfigHandler(config *utils.Config, rdb *redis.Client, factory algorithms.LimiterFactory) *ConfigHandler {
	return &ConfigHandler{
		config:  config,
		rdb:     rdb,
		factory: factory,
	}
}
