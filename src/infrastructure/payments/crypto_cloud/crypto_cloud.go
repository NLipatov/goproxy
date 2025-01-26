package crypto_cloud

import (
	"goproxy/application/payments/crypto_cloud_commands"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_configuration"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_dto"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_handlers"
	"net/http"
)

type CryptoCloudService struct {
	config     crypto_cloud_configuration.Configuration
	httpClient *http.Client
}

func NewCryptoCloudService() *CryptoCloudService {
	return &CryptoCloudService{
		config:     crypto_cloud_configuration.NewConfiguration(),
		httpClient: &http.Client{},
	}
}
func (s *CryptoCloudService) IssueInvoice(command crypto_cloud_commands.IssueInvoiceCommand) (interface{}, error) {
	request := crypto_cloud_dto.InvoiceRequest{
		Amount: command.AmountUSD,
		ShopID: s.config.ShopId(),
		Email:  command.Email,
	}

	url := s.config.BaseUrl() + s.config.CreateInvoiceUrl()

	return crypto_cloud_handlers.HandleCreateInvoice(s.httpClient, url, s.config.ApiKey(), request)
}
