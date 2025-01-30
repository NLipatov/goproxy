package modules

import (
	"database/sql"
	"goproxy/dal"
	"goproxy/dal/cache_serialization"
	"goproxy/dal/repositories"
	"goproxy/infrastructure/api/api-http/billing/generic"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"strconv"
)

type BillingAPI struct{}

func NewBillingAPI() *BillingAPI {
	return &BillingAPI{}
}

func (b *BillingAPI) Start() {
	strPort := os.Getenv("HTTP_PORT")
	if strPort == "" {
		log.Fatalf("'HTTP_PORT' env var must be set")
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

	planPriceRepoCache, planPriceRepoCacheErr := services.NewRedisCache[[]cache_serialization.PriceDto]()
	if planPriceRepoCacheErr != nil {
		log.Fatalf("failed to instantiate cache service: %s", planPriceRepoCacheErr)
	}
	planPriceRepository := repositories.NewPlanPriceRepository(db, planPriceRepoCache)
	controller := generic.NewBillingController(planPriceRepository)
	controller.Listen(port)
}
