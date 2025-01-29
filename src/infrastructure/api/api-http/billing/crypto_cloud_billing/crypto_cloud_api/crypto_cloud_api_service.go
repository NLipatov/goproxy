package crypto_cloud_api

import (
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_api_dto"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_configuration"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_handlers"
	"goproxy/infrastructure/services"
	"net/http"
)

type CryptoCloudService struct {
	createInvoiceHandler crypto_cloud_handlers.CreateInvoiceHandler
	postBackHandler      crypto_cloud_handlers.PostBackHandler
	config               crypto_cloud_configuration.Configuration
	httpClient           *http.Client
	messageBus           application.MessageBusService
}

func NewCryptoCloudService(messageBus application.MessageBusService) *CryptoCloudService {
	return &CryptoCloudService{
		config:               crypto_cloud_configuration.NewConfiguration(),
		httpClient:           &http.Client{},
		messageBus:           messageBus,
		postBackHandler:      crypto_cloud_handlers.NewPostBackHandler(services.NewHS256Jwt()),
		createInvoiceHandler: crypto_cloud_handlers.NewCreateInvoiceHandler(),
	}
}

func (s *CryptoCloudService) IssueInvoice(command crypto_cloud_commands.IssueInvoiceCommand) (interface{}, error) {
	currencyCode, currencyCodeErr := command.Currency.String()
	if currencyCodeErr != nil {
		return nil, currencyCodeErr
	}

	request := crypto_cloud_api_dto.InvoiceRequest{
		Amount:   command.Amount,
		Currency: currencyCode,
		ShopID:   s.config.ShopId(),
		Email:    command.Email,
		OrderID:  fmt.Sprintf("N_%d", command.OrderId),
	}

	url := s.config.BaseUrl() + s.config.CreateInvoiceUrl()

	return s.createInvoiceHandler.HandleCreateInvoice(s.httpClient, url, s.config.ApiKey(), request)
}

func (s *CryptoCloudService) HandlePostBack(command crypto_cloud_commands.PostBackCommand) error {
	command.Secret = s.config.SecretKey()
	return s.postBackHandler.HandlePostBack(command, s.messageBus)
}
