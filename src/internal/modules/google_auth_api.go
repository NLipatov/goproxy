package modules

import (
	"database/sql"
	"goproxy/application/use_cases"
	"goproxy/dal"
	"goproxy/dal/cache"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/infrastructure/api/api-http/google_auth"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/eventhandlers/UserPasswordChangedEvent"
	"goproxy/infrastructure/services"
	"log"
	"time"
)

type GoogleAuthAPI struct{}

func NewGoogleAuthAPI() *GoogleAuthAPI {
	return &GoogleAuthAPI{}
}

func (api *GoogleAuthAPI) Start() {
	oauthConfig := config.NewGoogleOauthConfig()
	db, err := dal.ConnectDB()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err != nil {
		log.Fatal(err)
	}

	userRepositoryCache, err := cache.NewBigCacheUserRepositoryCache(15*time.Minute, 1*time.Minute, 16, 512)
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

	eventHandleErr := UserPasswordChangedEvent.NewUserPasswordChangedEventProcessor[aggregates.User](domain.PROXY, userRepositoryCache).
		ProcessEvents()
	if eventHandleErr != nil {
		log.Fatal(eventHandleErr)
	}

	userRepo := repositories.NewUserRepository(db, userRepositoryCache)
	cryptoService := services.GetCryptoService()
	userUseCases := use_cases.NewUserUseCases(userRepo, cryptoService)
	authService := google_auth.NewGoogleAuthService(userUseCases, cryptoService, messageBusService)
	controller := google_auth.NewGoogleAuthController(authService)
	controller.Listen(oauthConfig.Port)
}
