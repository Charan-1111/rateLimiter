package handlers

import (
	"goapp/algorithms"
	"goapp/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type ConfigHandler struct {
	config  *utils.Config
	log     zerolog.Logger
	db      *pgxpool.Pool
	rdb     *redis.Client
	factory algorithms.LimiterFactory
}

func NewConfigHandler(config *utils.Config, log zerolog.Logger, db *pgxpool.Pool, rdb *redis.Client, factory algorithms.LimiterFactory) *ConfigHandler {
	return &ConfigHandler{
		config:  config,
		log:     log,
		db:      db,
		rdb:     rdb,
		factory: factory,
	}
}
