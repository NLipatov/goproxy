package modules

import (
	"context"
	"database/sql"
	"goproxy/application/use_cases"
	"goproxy/dal"
	"goproxy/dal/cache"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/domain/aggregates"
	"goproxy/infrastructure"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/eventhandlers/UserPasswordChangedEvent"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"strconv"
	"time"
)

type Proxy struct {
}

func NewProxy() *Proxy {
	return &Proxy{}
}

func (p *Proxy) Start() {
	strPort := os.Getenv("HTTP_LISTENER_PORT")
	if strPort == "" {
		log.Fatalf("'HTTP_LISTENER_PORT' env var must be set")
	}

	port, err := strconv.Atoi(strPort)
	if err != nil {
		log.Fatalf("failed to parse HTTP_LISTENER_PORT: %s", err)
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

	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(domain.PROXY)
	if kafkaConfigErr != nil {
		log.Fatal(kafkaConfigErr)
	}
	kafkaConfig.GroupID = "proxy"

	eventHandleErr := UserPasswordChangedEvent.NewUserPasswordChangedEventProcessor[aggregates.User](domain.PROXY, userRepositoryCache).
		ProcessEvents()
	if eventHandleErr != nil {
		log.Fatal(eventHandleErr)
	}

	userRepo := repositories.NewUserRepository(db, userRepositoryCache)

	userRestrictionService := services.NewUserRestrictionService()
	cryptoService := services.GetCryptoService()
	authCache := services.NewMapCacheWithTTL[services.ValidateResult]()

	authCacheEventHandlerErr := UserPasswordChangedEvent.NewUserPasswordChangedEventProcessor[services.ValidateResult](domain.PROXY, authCache).
		ProcessEvents()
	if authCacheEventHandlerErr != nil {
		log.Fatal(eventHandleErr)
	}

	authService := services.NewAuthService(cryptoService, authCache)
	authUseCases := use_cases.NewAuthUseCases(authService, userRepo, userRestrictionService)

	go userRestrictionService.ProcessEvents()

	dialerPool := services.NewDialerPool(services.NewIPResolver())
	dialerPool.StartExploringNewPublicIps(context.Background(), time.Hour*8)
	proxy := services.NewProxy(dialerPool)
	listener := infrastructure.NewHttpListener(proxy)
	proxyUseCases := use_cases.NewProxyUseCases(proxy, listener, authUseCases)
	proxyUseCases.ServeOnPort(port)
}
