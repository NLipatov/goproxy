package modules

import (
	"context"
	"database/sql"
	"goproxy/application/use_cases"
	"goproxy/dal"
	"goproxy/dal/cache"
	"goproxy/dal/cache_serialization"
	"goproxy/dal/repositories"
	"goproxy/domain"
	"goproxy/domain/dataobjects"
	apiws "goproxy/infrastructure/api/api-ws"
	"goproxy/infrastructure/eventhandlers/PlanAssignedEvent"
	"goproxy/infrastructure/eventhandlers/UserConsumedTrafficEvent"
	"goproxy/infrastructure/services"
	"log"
	"os"
	"time"
)

type PlanController struct {
}

func NewPlanController() *PlanController {
	return &PlanController{}
}

func (p *PlanController) Start() {
	usersRestApiHost := os.Getenv("USERS_API_HOST")
	if usersRestApiHost == "" {
		log.Fatalf("USERS_API_HOST environment variable not set")
	}

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

	planRepoCache, planRepoCacheErr := services.NewRedisCache[[]cache_serialization.PlanDto]()
	if planRepoCacheErr != nil {
		log.Fatalf("failed to instantiate cache service: %s", planRepoCacheErr)
	}
	planRepo := repositories.NewPlansRepository(db, planRepoCache)
	userPlanRepo := repositories.NewUserPlanRepository(db)

	userConsumedTrafficEventProcessorErr := UserConsumedTrafficEvent.
		NewUserConsumedTrafficEventProcessor(trafficCache, userPlanRepo, planRepo, domain.PLAN).
		ProcessEvents()

	if userConsumedTrafficEventProcessorErr != nil {
		log.Fatal(userConsumedTrafficEventProcessorErr)
	}

	bigCache, err := cache.NewBigCacheUserRepositoryCache(15*time.Minute, 1*time.Minute, 16, 512)
	if err != nil {
		log.Fatal(err)
	}

	planCache, planCacheErr := services.NewRedisCache[dataobjects.UserPlan]()
	if planCacheErr != nil {
		log.Fatalf("failed to instantiate cache service: %s", planCacheErr)
	}

	userRepo := repositories.NewUserRepository(db, bigCache)
	userPlanInfoUseCases := use_cases.NewUserPlanInfoUseCases(planRepo, userPlanRepo, userRepo, planCache, trafficCache)

	planAssignedEventProcessorErr := PlanAssignedEvent.
		NewPlanAssignedProcessor(domain.PLAN, planCache, userPlanRepo, userRepo, planRepo, trafficCache, usersRestApiHost).
		ProcessEvents()

	if planAssignedEventProcessorErr != nil {
		log.Fatal(planAssignedEventProcessorErr)
	}

	planController := apiws.NewPlanController(userPlanInfoUseCases, usersRestApiHost)
	planController.Listen(3031)

	for {
		select {
		case <-context.Background().Done():
			return
		default:
		}
	}
}
