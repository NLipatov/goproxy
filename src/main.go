package main

import (
	"database/sql"
	"goproxy/Application"
	data_access "goproxy/DataAccess"
	"goproxy/DataAccess/repositories"
	"goproxy/Infrastructure"
	"goproxy/Infrastructure/services"
	"log"
	"os"
	"time"
)

func main() {
	mode := os.Getenv("MODE")
	switch mode {
	case "migrator":
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

	db, err := data_access.ConnectDB()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err != nil {
		log.Fatal(err)
	}

	cache, err := createBigcacheInstance()
	if err != nil {
		log.Fatal(err)
	}

	userRepository := repositories.NewUserRepository(db, cache)
	cryptoService := services.NewCryptoService(32)
	authService := services.NewAuthService(cryptoService)
	authUseCases := Application.NewAuthUseCases(authService, userRepository)

	listener := Infrastructure.NewHttpListener(authUseCases)
	err = listener.ServePort(port)
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

	cache, err := createBigcacheInstance()
	if err != nil {
		log.Fatal(err)
	}

	userRepository := repositories.NewUserRepository(db, cache)
	cryptoService := services.NewCryptoService(32)
	useCases := Application.NewUserUseCases(userRepository, cryptoService)

	restApiListener := Infrastructure.NewHttpRestApiListener(useCases)
	err = restApiListener.ServePort(port)
	if err != nil {
		log.Printf("Failed serving port: %v", err)
	}
}

func createBigcacheInstance() (repositories.BigCacheUserRepositoryCache, error) {
	return repositories.NewBigCacheUserRepositoryCache(15*time.Minute, 1*time.Minute, 16, 512)
}
