package crypto_cloud_handlers

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/domain"
	"goproxy/domain/events"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api_interaction/crypto_cloud_configuration"
	"goproxy/infrastructure/services"
)

type PostBackHandler struct {
	config     crypto_cloud_configuration.Configuration
	jwt        application.Jwt
	messageBus application.MessageBusService
}

func NewPostBackHandler(messageBus application.MessageBusService) PostBackHandler {
	return PostBackHandler{
		jwt:        services.NewHS256Jwt(),
		config:     crypto_cloud_configuration.NewConfiguration(),
		messageBus: messageBus,
	}
}

func (p *PostBackHandler) HandlePostBack(command crypto_cloud_commands.PostBackCommand) error {
	isTokenValid, isTokenValidErr := p.jwt.Validate(p.config.SecretKey(), command.Token)
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

	produceErr := p.messageBus.Produce(fmt.Sprintf("%s", domain.BILLING), event)
	if produceErr != nil {
		return produceErr
	}

	return nil
}
