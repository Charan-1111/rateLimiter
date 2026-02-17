package main

func main() {
	app, err := server.NewApplication(); err != nil {
		log.Error("Error creating the application, retrying...")
		return
	}
}