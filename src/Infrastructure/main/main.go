package main

import (
	"log"
	"os"
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		log.Fatalf("'PORT' env var must be set")
	}

	listener := newHttpListener()
	err := listener.servePort(port)
	if err != nil {
		log.Printf("Failed serving port: %v", err)
	}
}
