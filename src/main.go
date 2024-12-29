package main

import (
	"database/sql"
	"goproxy/application"
	"goproxy/dal"
	"goproxy/dal/repositories"
	"goproxy/infrastructure"
	"goproxy/infrastructure/services"
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
	case "kafka-relay":
		startKafkaRelay()
	case "traffic-collector":
		startTrafficCollector()
	default:
		log.Fatalf("Unsupported mode: %s", mode)
	}
}

func startTrafficCollector() {
	collector, err := infrastructure.NewTrafficCollector()
	if err != nil {
		log.Fatal(err)
	}

	collector.ProcessEvents()
}

func startKafkaRelay() {
	infrastructure.StartOutboxProcessing()
}

func applyMigrations() {
	db, err := dal.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	dal.Migrate(db)
	return
}

func startHttpProxy() {
	port := os.Getenv("HTTP_LISTENER_PORT")
	if port == "" {
		log.Fatalf("'HTTP_LISTENER_PORT' env var must be set")
	}

	db, err := dal.ConnectDB()
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
	authUseCases := application.NewAuthUseCases(authService, userRepository)

	listener := infrastructure.NewHttpListener(authUseCases)
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

	db, err := dal.ConnectDB()
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
	useCases := application.NewUserUseCases(userRepository, cryptoService)

	restApiListener := infrastructure.NewHttpRestApiListener(useCases)
	err = restApiListener.ServePort(port)
	if err != nil {
		log.Printf("Failed serving port: %v", err)
	}
}

func createBigcacheInstance() (repositories.BigCacheUserRepositoryCache, error) {
	return repositories.NewBigCacheUserRepositoryCache(15*time.Minute, 1*time.Minute, 16, 512)
}
