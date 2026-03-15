package main

import (
	"goapp/metrics"
	"goapp/server"

	"github.com/gofiber/fiber/v2/log"
)

func main() {
	metrics.InitMetrics()

	filePath := "manifest/config.json"
	app, err := server.NewApplication(filePath)
	if err != nil {
		log.Error("Errro creating the application, retrying...")
		return
	}

	err = app.StartServer()
	if err != nil {
		log.Error("Error starting the server")
	}
}
