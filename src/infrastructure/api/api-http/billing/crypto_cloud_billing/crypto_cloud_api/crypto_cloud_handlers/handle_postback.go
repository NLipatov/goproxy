package crypto_cloud_handlers

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/domain"
	"goproxy/domain/events"
)

type PostBackHandler struct {
	jwt application.Jwt
}

func NewPostBackHandler(jwt application.Jwt) PostBackHandler {
	return PostBackHandler{
		jwt: jwt,
	}
}

func (p *PostBackHandler) HandlePostBack(command crypto_cloud_commands.PostBackCommand, messageBus application.MessageBusService) error {
	isTokenValid, isTokenValidErr := p.jwt.Validate(command.Secret, command.Token)
	if isTokenValidErr != nil {
		return fmt.Errorf("could not validate token: %s", isTokenValidErr)
	}

	if !isTokenValid {
		return fmt.Errorf("invalid token")
	}

	orderPaidEvent := events.NewOrderPaidEvent(command.OrderID)
	serializedEvent, serializedEventErr := json.Marshal(orderPaidEvent)
	if serializedEventErr != nil {
		return serializedEventErr
	}

	event, eventErr := events.NewOutboxEvent(-1, string(serializedEvent), false, "OrderPaidEvent")
	if eventErr != nil {
		return eventErr
	}

	produceErr := messageBus.Produce(fmt.Sprintf("%s", domain.BILLING), event)
	if produceErr != nil {
		return produceErr
	}

	return nil
}
