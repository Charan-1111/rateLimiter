package server

import (
	"context"
	"goapp/algorithms"
	"goapp/constants"
	"goapp/logger"
	"goapp/services"
	"goapp/store"
	"goapp/utils"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

type Application struct {
	ctx     context.Context
	config  *utils.Config
	log     zerolog.Logger
	db      *pgxpool.Pool
	rdb     *redis.Client
	cache   *services.Cache
	factory algorithms.LimiterFactory
	cb      *services.CircuitBreaker
}

func NewApplication(filePath string) (*Application, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ContextTimeout)
	defer cancel()

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

	// Initializing circuit breaker
	cb := services.NewCircuitBreaker()

	return &Application{
		ctx:     ctx,
		config:  config,
		log:     log,
		db:      db,
		rdb:     rdb,
		cache:   cache,
		factory: factory,
		cb:      cb,
	}, nil
}

func (app *Application) StartServer() error {
	// create the tables
	store.CreateTables(app.ctx, app.db, app.log, app.config.Tables)

	// Load the cache
	app.cache.LoadCache(app.ctx, app.log, app.db, app.config.Queries.Fetch.FetchPolicies)

	// Start fiber server
	app.StartFiberServer()

	return nil
}

func (app *Application) StartFiberServer() {
	appServer := app.SetupRoutes()

	// using the channel to handle the graceful shutdown
	listenErr := make(chan error)
	go func() {
		listenErr <- appServer.Listen(app.config.Ports.FiberServer)
	}()

	app.log.Info().Msg("Server started listening at port : " + app.config.Ports.FiberServer)

	// waiting for the interrupt signal for graceful shutdowning the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-listenErr:
		app.log.Error().Err(err).Msg("Error starting fiber server")
	case <-quit:
		app.log.Info().Msg("Graceful shutdown initiated")

	}
}
