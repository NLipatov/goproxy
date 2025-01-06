package main

import (
	"database/sql"
	"fmt"
	"goproxy/application"
	"goproxy/dal"
	"goproxy/dal/repositories"
	"goproxy/infrastructure"
	"goproxy/infrastructure/dto"
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
	case "plan-controller":
		startTrafficCollector()
	default:
		log.Fatalf("Unsupported mode: %s", mode)
	}
}

func startTrafficCollector() {
	db, err := dal.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	messageBus, newMBErr := instantiateMessageBusService()
	if newMBErr != nil {
		log.Fatalf("failed to instantiate message bus service: %s", newMBErr)
	}

	redisCache, newRedisClientErr := services.NewRedisCache[dto.UserTraffic]()
	if newRedisClientErr != nil {
		log.Fatalf("failed to instantiate redis cache service: %s", newRedisClientErr)
	}

	planRepo := repositories.NewPlansRepository(db)
	userPlanRepo := repositories.NewUserPlanRepository(db)

	controller, err := infrastructure.NewTrafficCollector().
		UsePlanRepository(planRepo).
		UseUserPlanRepository(userPlanRepo).
		UseMessageBus(messageBus).
		UseCache(redisCache).
		Build()

	if err != nil {
		log.Fatalf("failed to instantiate plan controller: %s", err)
	}

	controller.ProcessEvents()
}

func instantiateMessageBusService() (application.MessageBusService, error) {
	bootstrapServers := os.Getenv("KAFKA_BOOTSTRAP_SERVERS")
	groupId := os.Getenv("KAFKA_GROUP_ID")
	autoOffsetReset := os.Getenv("KAFKA_AUTO_OFFSET_RESET")
	topic := os.Getenv("KAFKA_TOPIC")
	plansTopic := os.Getenv("PLANS_KAFKA_TOPIC")

	if groupId == "" || autoOffsetReset == "" || topic == "" || bootstrapServers == "" || plansTopic == "" {
		return nil, fmt.Errorf("invalid configuration")
	}

	messageBusService, err := services.NewKafkaService(bootstrapServers, groupId, autoOffsetReset)
	if err != nil {
		return nil, err
	}

	return messageBusService, nil
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
