package services

import (
	"encoding/json"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/events"
	"log"
	"os"
	"strings"
)

type UserRestrictionService struct {
	restrictedIds map[int]bool
	messageBus    application.MessageBusService
}

func NewUserRestrictionService() *UserRestrictionService {
	bootstrapServers := os.Getenv("UP_KAFKA_BOOTSTRAP_SERVERS")
	groupId := os.Getenv("UP_KAFKA_GROUP_ID")
	autoOffsetReset := os.Getenv("UP_KAFKA_AUTO_OFFSET_RESET")

	messageBusService, err := NewKafkaService(bootstrapServers, groupId, autoOffsetReset)
	if err != nil {
		log.Fatalf("failed to initialize kafka service: %s", err)
	}

	return &UserRestrictionService{
		restrictedIds: make(map[int]bool),
		messageBus:    messageBusService,
	}
}

func (u *UserRestrictionService) IsRestricted(user aggregates.User) bool {
	val, ok := u.restrictedIds[user.Id()]
	if ok {
		return val
	}

	return false
}

func (u *UserRestrictionService) AddToRestrictionList(user aggregates.User) error {
	u.restrictedIds[user.Id()] = true
	return nil
}

func (u *UserRestrictionService) RemoveFromRestrictionList(user aggregates.User) error {
	delete(u.restrictedIds, user.Id())
	return nil
}

func (u *UserRestrictionService) ProcessEvents() {
	defer func(messageBus application.MessageBusService) {
		_ = messageBus.Close()
	}(u.messageBus)

	topics := []string{"user-plans"}
	err := u.messageBus.Subscribe(topics)
	if err != nil {
		log.Fatalf("Failed to subscribe to topics: %s", err)
	}

	log.Printf("Subscribed to topics: %s", strings.Join(topics, ", "))

	for {
		event, consumeErr := u.messageBus.Consume()
		if consumeErr != nil {
			log.Printf("failed to consume from message bus: %s", consumeErr)
		}

		if event.EventType.Value() == "UserConsumedTrafficWithoutPlan" {
			log.Printf(event.EventType.Value())
		}

		if event.EventType.Value() == "UserExceededTrafficLimitEvent" {
			var userExceededTrafficLimitEvent events.UserExceededTrafficLimitEvent
			deserializationErr := json.Unmarshal([]byte(event.Payload), &userExceededTrafficLimitEvent)
			if deserializationErr != nil {
				log.Printf("failed to deserialize user exceeded threshold event: %s", deserializationErr)
			}

			u.restrictedIds[userExceededTrafficLimitEvent.UserId] = true
		}
	}
}
