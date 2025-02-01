package modules

import (
	"database/sql"
	"goproxy/dal"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/infrastructure/api/api-http/billing/free_plan_billing"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"strconv"
)

type FreePlanBillingAPI struct{}

func NewFreePlanBillingAPI() *FreePlanBillingAPI {
	return &FreePlanBillingAPI{}
}

func (f *FreePlanBillingAPI) Start() {
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
	kafkaConf.GroupID = "free-plan-billing"

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

	orderRepository := repositories.NewOrderRepository(db)
	service := free_plan_billing.NewService(orderRepository, messageBusService)
	controller := free_plan_billing.NewController(service)
	controller.Listen(port)
}
