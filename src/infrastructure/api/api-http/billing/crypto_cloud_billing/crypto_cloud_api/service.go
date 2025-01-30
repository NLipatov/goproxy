package crypto_cloud_api

import (
	"goproxy/application/contracts"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_handlers"
)

type CryptoCloudService struct {
	createInvoiceHandler crypto_cloud_handlers.CreateInvoiceHandler
	postBackHandler      crypto_cloud_handlers.PostBackHandler
	messageBus           contracts.MessageBusService
}

func NewCryptoCloudService(messageBus contracts.MessageBusService) *CryptoCloudService {
	return &CryptoCloudService{
		messageBus:           messageBus,
		postBackHandler:      crypto_cloud_handlers.NewPostBackHandler(messageBus),
		createInvoiceHandler: crypto_cloud_handlers.NewCreateInvoiceHandler(),
	}
}

func (s *CryptoCloudService) IssueInvoice(command crypto_cloud_commands.IssueInvoiceCommand) (interface{}, error) {
	return s.createInvoiceHandler.HandleCreateInvoice(command)
}

func (s *CryptoCloudService) HandlePostBack(command crypto_cloud_commands.PostBackCommand) error {
	return s.postBackHandler.HandlePostBack(command)
}
