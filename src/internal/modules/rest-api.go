package modules

import (
	"database/sql"
	"goproxy/application/use_cases"
	"goproxy/dal"
	"goproxy/dal/cache"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/infrastructure/api/api-http/users"
	"goproxy/infrastructure/eventhandlers/UserPasswordChangedEvent"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"strconv"
	"time"
)

type UsersApi struct {
}

func NewUsersApi() *UsersApi {
	return &UsersApi{}
}

func (api *UsersApi) Start() {
	strPort := os.Getenv("HTTP_REST_API_PORT")
	if strPort == "" {
		log.Fatalf("'HTTP_REST_API_PORT' env var must be set")
	}
	port, portErr := strconv.Atoi(strPort)
	if portErr != nil {
		log.Fatal(portErr)
	}

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

	eventHandleErr := UserPasswordChangedEvent.NewUserPasswordChangedEventProcessor[aggregates.User](domain.PROXY, userRepositoryCache).
		ProcessEvents()
	if eventHandleErr != nil {
		log.Fatal(eventHandleErr)
	}

	userRepo := repositories.NewUserRepository(db, userRepositoryCache)

	cryptoService := services.GetCryptoService()
	useCases := use_cases.NewUserUseCases(userRepo, cryptoService)

	usersController := users.NewUsersController(useCases)
	usersController.Listen(port)
}
