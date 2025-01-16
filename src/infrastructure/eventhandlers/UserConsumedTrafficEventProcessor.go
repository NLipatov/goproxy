package eventhandlers

import (
	"context"
	"fmt"
	"goproxy/application"
	"goproxy/domain"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/dto"
	"goproxy/infrastructure/services"
	"log"
)

type UserConsumedTrafficEventProcessor struct {
	boundedContext     domain.BoundedContexts
	cache              application.CacheWithTTL[dto.UserTraffic]
	userPlanRepository application.UserPlanRepository
	planRepository     application.PlanRepository
	messageBus         application.MessageBusService
}

func NewUserConsumedTrafficEventProcessor(cache application.CacheWithTTL[dto.UserTraffic],
	userPlanRepository application.UserPlanRepository,
	planRepository application.PlanRepository,
	boundedContext domain.BoundedContexts) *UserConsumedTrafficEventProcessor {

	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(boundedContext)
	if kafkaConfigErr != nil {
		log.Fatalf("failed to create kafka congif: %s", kafkaConfigErr)
	}

	kafkaConfig.GroupID = "UserConsumedTrafficEventProcessor"

	kafka, kafkaErr := services.NewKafkaService(kafkaConfig)
	if kafkaErr != nil {
		log.Fatal(kafkaErr)
	}

	return &UserConsumedTrafficEventProcessor{
		cache:              cache,
		userPlanRepository: userPlanRepository,
		planRepository:     planRepository,
		boundedContext:     boundedContext,
		messageBus:         kafka,
	}
}
func (u *UserConsumedTrafficEventProcessor) ProcessEvents() error {
	eventHandler := NewUserConsumedTrafficEventHandler(u.cache, u.userPlanRepository, u.planRepository, u.messageBus)
	eventProcessor := application.NewEventProcessor(u.messageBus).
		RegisterTopic(fmt.Sprintf("%s", u.boundedContext)).
		RegisterHandler("UserConsumedTrafficEvent", eventHandler)

	if eventProcessorErr := eventProcessor.Build(); eventProcessorErr != nil {
		log.Fatal(eventProcessorErr)
	}

	processingErr := eventProcessor.Start(context.Background())
	if processingErr != nil {
		log.Fatal(processingErr)
	}

	return nil
}
