package crypto_cloud_handlers

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud_commands"
	"goproxy/domain"
	"goproxy/domain/events"
)

func HandlePostBack(command crypto_cloud_commands.PostBackCommand, messageBus application.MessageBusService) (interface{}, error) {
	orderPaidEvent := events.NewOrderPaidEvent(command.OrderID)
	serializedEvent, serializedEventErr := json.Marshal(orderPaidEvent)
	if serializedEventErr != nil {
		return nil, serializedEventErr
	}

	event, eventErr := events.NewOutboxEvent(-1, string(serializedEvent), false, "OrderPaidEvent")
	if eventErr != nil {
		return nil, eventErr
	}

	produceErr := messageBus.Produce(fmt.Sprintf("%s", domain.BILLING), event)
	if produceErr != nil {
		return nil, produceErr
	}

	return nil, nil
}
