package handlers

import (
	"context"
	"goapp/algorithms"
	"goapp/services"
	"goapp/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type ConfigHandler struct {
	ctx     context.Context
	config  *utils.Config
	log     zerolog.Logger
	db      *pgxpool.Pool
	rdb     *redis.Client
	cache   *services.Cache
	factory algorithms.LimiterFactory
}

func NewConfigHandler(ctx context.Context, config *utils.Config, log zerolog.Logger, db *pgxpool.Pool, rdb *redis.Client, factory algorithms.LimiterFactory, cache *services.Cache) *ConfigHandler {
	return &ConfigHandler{
		ctx:     ctx,
		config:  config,
		log:     log,
		db:      db,
		rdb:     rdb,
		cache:   cache,
		factory: factory,
	}
}
