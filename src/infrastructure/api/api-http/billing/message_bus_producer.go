package billing

import (
	"encoding/json"
	"fmt"
	"goproxy/application/contracts"
	"goproxy/domain"
	"goproxy/domain/events"
	"time"
)

type MessageBusProducer struct {
	messageBus contracts.MessageBusService
}

func NewMessageBusProducer(messageBus contracts.MessageBusService) MessageBusProducer {
	return MessageBusProducer{
		messageBus: messageBus,
	}
}

func (m *MessageBusProducer) ProducePlanAssignedEvent(planId int, email string) error {
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
