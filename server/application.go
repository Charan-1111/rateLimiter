package server

import (
	"fmt"
	"goapp/algorithms"
	"goapp/store"
	"goapp/utils"

	"github.com/redis/go-redis/v9"
)

type Application struct {
	config  *utils.Config
	rdb     *redis.Client
	factory algorithms.LimiterFactory
}

func NewApplication(filePath string) (*Application, error) {
	// Load the configuration file
	config := &utils.Config{}
	if err := config.LoadConfig(filePath); err != nil {
		return nil, err
	}

	// Initialize Redis
	rdb := store.InitRedis(&config.Redis)

	// creating the defautl limiter factory
	factory := &algorithms.DefaultLimiterFactory{}

	return &Application{
		config:  config,
		rdb:     rdb,
		factory: factory,
	}, nil
}

func (app *Application) StartServer() error {
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
