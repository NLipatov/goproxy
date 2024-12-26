package main

import (
	"database/sql"
	"goproxy/Application"
	data_access "goproxy/DataAccess"
	"goproxy/DataAccess/Repositories"
	"goproxy/Infrastructure"
	"log"
	"os"
)

func main() {
	mode := os.Getenv("MODE")
	switch mode {
	case "migrate":
		applyMigrations()
	case "proxy":
		startHttpProxy()
	case "rest-api":
		startHttpRestApi()
	default:
		log.Fatalf("Unsupported mode: %s", mode)
	}
}

func applyMigrations() {
	db, err := data_access.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	data_access.Migrate(db)
	return
}

func startHttpProxy() {
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

func startHttpRestApi() {
	port := os.Getenv("HTTP_REST_API_PORT")
	if port == "" {
		log.Fatalf("'HTTP_REST_API_PORT' env var must be set")
	}

	db, err := data_access.ConnectDB()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err != nil {
		log.Fatal(err)
	}

	userRepository := Repositories.NewUserRepository(db)
	useCases := Application.NewUserUseCases(userRepository)
	restApiListener := Infrastructure.NewHttpRestApiListener(useCases)
	err = restApiListener.ServePort(port)
	if err != nil {
		log.Printf("Failed serving port: %v", err)
	}
}
