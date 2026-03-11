package server

import (
	"context"
	"fmt"
	"goapp/algorithms"
	"goapp/logger"
	"goapp/services"
	"goapp/store"
	"goapp/utils"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Application struct {
	config  *utils.Config
	log     zerolog.Logger
	db      *pgxpool.Pool
	rdb     *redis.Client
	cache   *services.Cache
	factory algorithms.LimiterFactory
}

func NewApplication(filePath string) (*Application, error) {
	// Initiating the log
	log := logger.InitLogger()

	// Load the configuration file
	config := &utils.Config{}
	if err := config.LoadConfig(filePath); err != nil {
		log.Error().Err(err).Msg("Error loading the config file")
		return nil, err
	}

	// Initialize the database

	db, _ := config.Database.InitDb(context.Background(), log)

	// Initialize Redis
	rdb := store.InitRedis(&config.Redis, log)

	// creating the defautl limiter factory
	factory := &algorithms.DefaultLimiterFactory{}

	// create the cache variable

	cache := services.NewCache()

	return &Application{
		config:  config,
		log:     log,
		db:      db,
		rdb:     rdb,
		cache:   cache,
		factory: factory,
	}, nil
}

func (app *Application) StartServer() error {
	// create the tables
	store.CreateTables(context.Background(), app.db, app.log, app.config.Tables)

	// Start fiber server
	app.StartFiberServer()

	return nil
}

func (app *Application) StartFiberServer() {
	appServer := app.SetupRoutes()

	if err := appServer.Listen(app.config.Ports.FiberServer); err != nil {
		fmt.Println("Error starting fiber server:", err)
	} else {
		fmt.Println("Fiber server started on port", app.config.Ports.FiberServer)
	}
}
