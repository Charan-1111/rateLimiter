package server

import (
	"goapp/handlers"

	"github.com/gofiber/fiber/v2"
)

func (app *Application) SetupRoutes() *fiber.App {
	appServer := fiber.New()

	configHandler := handlers.NewConfigHandler(app.config, app.rdb, app.factory)

	// Defining the routes
	appServer.Get("/api/v1/limiter", configHandler.GetLimiter)

	return appServer
}
