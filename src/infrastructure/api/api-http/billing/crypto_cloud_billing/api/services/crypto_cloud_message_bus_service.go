package services

import (
	"encoding/json"
	"fmt"
	"goproxy/application/contracts"
	"goproxy/domain"
	"goproxy/domain/events"
	"time"
)

type CryptoCloudMessageBusService struct {
	messageBus contracts.MessageBusService
}

func NewCryptoCloudMessageBusService(messageBus contracts.MessageBusService) CryptoCloudMessageBusService {
	return CryptoCloudMessageBusService{
		messageBus: messageBus,
	}
}

func (m *CryptoCloudMessageBusService) ProducePlanAssignedEvent(planId int, email string) error {
	event := events.NewPlanAssigned(email, planId, time.Now().UTC())
	eventJson, eventJsonErr := json.Marshal(event)
	if eventJsonErr != nil {
		return eventJsonErr
	}

	outboxEvent, outboxEventErr := events.NewOutboxEvent(-1, string(eventJson), false, "PlanAssignedEvent")
	if outboxEventErr != nil {
		return outboxEventErr
	}

	produceErr := m.messageBus.Produce(fmt.Sprintf("%s", domain.PLAN), outboxEvent)
	if produceErr != nil {
		return produceErr
	}

	return nil
}
