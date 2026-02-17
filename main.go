package main

import (
	"fmt"

	"goapp/server"
)

func main() {
	filePath := "manifest/config.json"
	_, err := server.NewApplication(filePath)
	if err != nil {
		fmt.Println("Error creating the application, retrying...")
		return
	}
}