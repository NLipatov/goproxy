package main

import (
	data_access "goproxy/DataAccess"
	"goproxy/Infrastructure"
	"log"
	"os"
)

func main() {
	args := os.Args[1:]
	if len(args) == 1 && args[0] == "migrate" {
		data_access.Migrate()
		return
	}

	port := os.Getenv("HTTP_LISTENER_PORT")
	if port == "" {
		log.Fatalf("'HTTP_LISTENER_PORT' env var must be set")
	}

	listener := Infrastructure.NewHttpListener()
	err := listener.ServePort(port)
	if err != nil {
		log.Printf("Failed serving port: %v", err)
	}
}
