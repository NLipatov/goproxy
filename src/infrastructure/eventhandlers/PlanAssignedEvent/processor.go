package PlanAssignedEvent

import (
	"context"
	"fmt"
	"goproxy/application"
	"goproxy/application/contracts"
	"goproxy/domain"
	"goproxy/domain/dataobjects"
	"goproxy/infrastructure/config"
	"goproxy/infrastructure/services"
	"log"
)

type PlanAssignedProcessor struct {
	boundedContext     domain.BoundedContexts
	messageBus         contracts.MessageBusService
	userPlanCache      contracts.CacheWithTTL[dataobjects.UserPlan]
	userPlanRepository contracts.UserPlanRepository
	userRepository     contracts.UserRepository
	planRepository     contracts.PlanRepository
	trafficCache       contracts.CacheWithTTL[dataobjects.UserTraffic]
	userApiHost        string
}

func NewPlanAssignedProcessor(boundedContext domain.BoundedContexts,
	userPlanCache contracts.CacheWithTTL[dataobjects.UserPlan],
	userPlanRepository contracts.UserPlanRepository,
	userRepository contracts.UserRepository,
	planRepository contracts.PlanRepository,
	trafficCache contracts.CacheWithTTL[dataobjects.UserTraffic],
	userApiHost string) *PlanAssignedProcessor {
	kafkaConfig, kafkaConfigErr := config.NewKafkaConfig(boundedContext)
	if kafkaConfigErr != nil {
		log.Fatalf("failed to create kafka congif: %s", kafkaConfigErr)
	}

	kafkaConfig.GroupID = "PlanAssignedEventProcessor"

	kafka, kafkaErr := services.NewKafkaService(kafkaConfig)
	if kafkaErr != nil {
		log.Fatalf("failed to create kafka service: %s", kafkaErr)
	}

	return &PlanAssignedProcessor{
		boundedContext:     boundedContext,
		messageBus:         kafka,
		userPlanCache:      userPlanCache,
		userPlanRepository: userPlanRepository,
		userRepository:     userRepository,
		planRepository:     planRepository,
		trafficCache:       trafficCache,
		userApiHost:        userApiHost,
	}
}

func (p *PlanAssignedProcessor) ProcessEvents() error {
	eventHandler := NewPlanAssignedHandler(p.messageBus, p.userPlanCache, p.userPlanRepository, p.userRepository,
		p.planRepository, p.trafficCache, p.userApiHost)

	eventProcessor := application.NewEventProcessor(p.messageBus).
		RegisterTopic(fmt.Sprintf("%s", p.boundedContext)).
		RegisterHandler("PlanAssignedEvent", eventHandler)

	processingErr := eventProcessor.Start(context.Background())
	if processingErr != nil {
		log.Fatal(processingErr)
	}

	return nil
}
