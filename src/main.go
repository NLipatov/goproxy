package main

import (
	"database/sql"
	"goproxy/application"
	"goproxy/dal"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/infrastructure"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/dto"
	"goproxy/infrastructure/eventhandlers"
	"goproxy/infrastructure/restapi"
	"goproxy/infrastructure/restapi/google_auth"
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
		startPlanController()
	case "google-auth":
		startGoogleAuthController()
	default:
		log.Fatalf("Unsupported mode: %s", mode)
	}
}

func startGoogleAuthController() {
	oauthConfig := config.NewGoogleOauthConfig()
	db, err := dal.ConnectDB()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err != nil {
		log.Fatal(err)
	}

	cache, err := repositories.NewBigCacheUserRepositoryCache(15*time.Minute, 1*time.Minute, 16, 512)
	if err != nil {
		log.Fatal(err)
	}

	kafkaConf, kafkaConfErr := config.NewKafkaConfig(domain.PROXY)
	if kafkaConfErr != nil {
		log.Fatal(kafkaConfErr)
	}
	kafkaConf.GroupID = "google-auth"

	messageBusService, err := services.NewKafkaService(kafkaConf)
	if err != nil {
		log.Fatal(err)
	}

	eventHandleErr := eventhandlers.NewUserPasswordChangedEventProcessor[aggregates.User](domain.PROXY, cache).
		ProcessEvents()
	if eventHandleErr != nil {
		log.Fatal(eventHandleErr)
	}

	userRepo := repositories.NewUserRepository(db, cache)
	cryptoService := services.GetCryptoService()
	userUseCases := application.NewUserUseCases(userRepo, cryptoService)
	authService := google_auth.NewGoogleAuthService(userUseCases, cryptoService, messageBusService)
	controller := google_auth.NewGoogleAuthController(authService)

	controller.Listen(oauthConfig.Port)
}

func startPlanController() {
	db, err := dal.ConnectDB()
	if err != nil {
		log.Fatal(err)
	}
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)

	kafkaConf, kafkaConfErr := config.NewKafkaConfig(domain.PLAN)
	if kafkaConfErr != nil {
		log.Fatal(kafkaConfErr)
	}

	messageBusService, err := services.NewKafkaService(kafkaConf)
	if err != nil {
		log.Fatal(err)
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
		UseMessageBus(messageBusService).
		UseCache(redisCache).
		Build()

	if err != nil {
		log.Fatalf("failed to instantiate plan controller: %s", err)
	}

	controller.ProcessEvents()
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

	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(domain.PROXY)
	if kafkaConfigErr != nil {
		log.Fatal(kafkaConfigErr)
	}
	kafkaConfig.GroupID = "proxy"

	eventHandleErr := eventhandlers.NewUserPasswordChangedEventProcessor[aggregates.User](domain.PROXY, cache).
		ProcessEvents()
	if eventHandleErr != nil {
		log.Fatal(eventHandleErr)
	}

	userRepo := repositories.NewUserRepository(db, cache)

	userRestrictionService := services.NewUserRestrictionService()
	cryptoService := services.GetCryptoService()
	authCache := services.NewMapCacheWithTTL[services.ValidateResult]()

	authCacheEventHandlerErr := eventhandlers.NewUserPasswordChangedEventProcessor[services.ValidateResult](domain.PROXY, authCache).
		ProcessEvents()
	if authCacheEventHandlerErr != nil {
		log.Fatal(eventHandleErr)
	}

	authService := services.NewAuthService(cryptoService, authCache)
	authUseCases := application.NewAuthUseCases(authService, userRepo, userRestrictionService)

	go userRestrictionService.ProcessEvents()

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

	eventHandleErr := eventhandlers.NewUserPasswordChangedEventProcessor[aggregates.User](domain.PROXY, cache).
		ProcessEvents()
	if eventHandleErr != nil {
		log.Fatal(eventHandleErr)
	}

	userRepo := repositories.NewUserRepository(db, cache)

	cryptoService := services.GetCryptoService()
	useCases := application.NewUserUseCases(userRepo, cryptoService)

	usersController := restapi.NewUsersController(useCases)
	err = usersController.ServePort(port)
	if err != nil {
		log.Printf("Failed serving port: %v", err)
	}
}

func createBigcacheInstance() (repositories.BigCacheUserRepositoryCache, error) {
	return repositories.NewBigCacheUserRepositoryCache(15*time.Minute, 1*time.Minute, 16, 512)
}
