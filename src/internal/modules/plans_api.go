package modules

import (
	"database/sql"
	"goproxy/dal"
	"goproxy/dal/cache_serialization"
	"goproxy/dal/repositories"
	"goproxy/infrastructure/api/api-http/plans"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"strconv"
)

type PlansAPI struct{}

func NewPlansAPI() *PlansAPI {
	return &PlansAPI{}
}

func (p *PlansAPI) Start() {
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

	planRepoCache, planRepoCacheErr := services.NewRedisCache[[]cache_serialization.PlanDto]()
	if planRepoCacheErr != nil {
		log.Fatalf("failed to instantiate cache service: %s", planRepoCacheErr)
	}
	planRepository := repositories.NewPlansRepository(db, planRepoCache)
	plansController := plans.NewPlansController(planRepository)
	plansController.Listen(port)
}
