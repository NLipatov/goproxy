package main

import (
	"context"
	"database/sql"
	"goproxy/application"
	"goproxy/dal"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/domain/dataobjects"
	"goproxy/infrastructure"
	"goproxy/infrastructure/api/api-http"
	"goproxy/infrastructure/api/api-http/google_auth"
	"goproxy/infrastructure/api/api-ws"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/eventhandlers"
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

	trafficCache, trafficCacheErr := services.NewRedisCache[dataobjects.UserTraffic]()
	if trafficCacheErr != nil {
		log.Fatalf("failed to instantiate cache service: %s", trafficCacheErr)
	}

	planRepo := repositories.NewPlansRepository(db)
	userPlanRepo := repositories.NewUserPlanRepository(db)

	userConsumedTrafficEventProcessorErr := eventhandlers.
		NewUserConsumedTrafficEventProcessor(trafficCache, userPlanRepo, planRepo, domain.PLAN).
		ProcessEvents()

	if userConsumedTrafficEventProcessorErr != nil {
		log.Fatal(userConsumedTrafficEventProcessorErr)
	}

	bigCache, err := repositories.NewBigCacheUserRepositoryCache(15*time.Minute, 1*time.Minute, 16, 512)
	if err != nil {
		log.Fatal(err)
	}

	planCache, planCacheErr := services.NewRedisCache[dataobjects.UserPlan]()
	if planCacheErr != nil {
		log.Fatalf("failed to instantiate cache service: %s", planCacheErr)
	}

	userRepo := repositories.NewUserRepository(db, bigCache)
	userPlanInfoUseCases := application.NewUserPlanInfoUseCases(planRepo, userPlanRepo, userRepo, planCache, trafficCache)
	cryptoService := services.GetCryptoService()
	userUseCases := application.NewUserUseCases(userRepo, cryptoService)
	planController := api_ws.NewPlanController(userUseCases, userPlanInfoUseCases)
	planController.Listen(3031)

	for {
		select {
		case <-context.Background().Done():
			return
		default:
		}
	}
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

	usersController := api_http.NewUsersController(useCases)
	err = usersController.ServePort(port)
	if err != nil {
		log.Printf("Failed serving port: %v", err)
	}
}

func createBigcacheInstance() (repositories.BigCacheUserRepositoryCache, error) {
	return repositories.NewBigCacheUserRepositoryCache(15*time.Minute, 1*time.Minute, 16, 512)
}
