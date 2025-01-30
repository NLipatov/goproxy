package modules

import (
	"database/sql"
	"goproxy/dal"
	"goproxy/dal/cache_serialization"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"strconv"
)

type CloudBillingAPI struct{}

func NewCryptoCloudBillingAPI() *CloudBillingAPI {
	return &CloudBillingAPI{}
}

func (*CloudBillingAPI) Start() {
	strPort := os.Getenv("HTTP_PORT")
	if strPort == "" {
		log.Fatalf("'HTTP_PORT' env var must be set")
	}
	port, portErr := strconv.Atoi(strPort)

	if portErr != nil {
		log.Fatal(portErr)
	}
	kafkaConf, kafkaConfErr := config.NewKafkaConfig(domain.BILLING)
	if kafkaConfErr != nil {
		log.Fatal(kafkaConfErr)
	}
	kafkaConf.GroupID = "crypto-cloud-billing"

	messageBusService, err := services.NewKafkaService(kafkaConf)
	if err != nil {
		log.Fatal(err)
	}
	db, err := dal.ConnectDB()
	defer func(db *sql.DB) {
		_ = db.Close()
	}(db)
	if err != nil {
		log.Fatal(err)
	}

	planPriceRepoCache, planPriceRepoCacheErr := services.NewRedisCache[[]cache_serialization.PriceDto]()
	if planPriceRepoCacheErr != nil {
		log.Fatalf("failed to instantiate cache service: %s", planPriceRepoCacheErr)
	}
	planPriceRepository := repositories.NewPlanPriceRepository(db, planPriceRepoCache)
	orderRepository := repositories.NewOrderRepository(db)

	cryptoCloudService := crypto_cloud_api.NewCryptoCloudService(messageBusService)
	restController := api.NewController(cryptoCloudService, planPriceRepository, orderRepository, messageBusService)
	restController.Listen(port)
}
