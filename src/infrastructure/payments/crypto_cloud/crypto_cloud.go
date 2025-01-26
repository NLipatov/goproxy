package crypto_cloud

import (
	"goproxy/application"
	crypto_cloud_commands2 "goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_configuration"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_dto"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_handlers"
	"net/http"
)

type CryptoCloudService struct {
	config     crypto_cloud_configuration.Configuration
	httpClient *http.Client
	messageBus application.MessageBusService
}

func NewCryptoCloudService(messageBus application.MessageBusService) *CryptoCloudService {
	return &CryptoCloudService{
		config:     crypto_cloud_configuration.NewConfiguration(),
		httpClient: &http.Client{},
		messageBus: messageBus,
	}
}

func (s *CryptoCloudService) IssueInvoice(command crypto_cloud_commands2.IssueInvoiceCommand) (interface{}, error) {
	request := crypto_cloud_dto.InvoiceRequest{
		Amount:  command.AmountUSD,
		ShopID:  s.config.ShopId(),
		Email:   command.Email,
		OrderID: command.OrderId,
	}

	url := s.config.BaseUrl() + s.config.CreateInvoiceUrl()

	return crypto_cloud_handlers.HandleCreateInvoice(s.httpClient, url, s.config.ApiKey(), request)
}

func (s *CryptoCloudService) HandlePostBack(command crypto_cloud_commands2.PostBackCommand) (interface{}, error) {
	return crypto_cloud_handlers.HandlePostBack(command, s.messageBus)
}
