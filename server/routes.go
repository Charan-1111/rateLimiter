package server

import (
	"goapp/handlers"

	"github.com/gofiber/adaptor/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/requestid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func (app *Application) SetupRoutes() *fiber.App {
	appServer := fiber.New()
	appServer.Use(requestid.New())

	configHandler := handlers.NewConfigHandler(app.ctx, app.config, app.log, app.db, app.rdb, app.factory, app.cache, app.cb)

	// Defining the routes
	appServer.Get("/api/v1/limiter", configHandler.GetLimiter)

	appServer.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

	appServer.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})
	return appServer
}
