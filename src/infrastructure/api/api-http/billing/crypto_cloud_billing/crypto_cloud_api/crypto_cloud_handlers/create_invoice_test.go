package crypto_cloud_handlers

import (
	"github.com/stretchr/testify/assert"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/dal/repositories/mocks"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api"
	"os"
	"testing"
)

func TestCreateInvoiceIntegration(t *testing.T) {
	apiKey := os.Getenv("API_KEY")
	if apiKey == "" {
		t.Logf("Skipping crypto cloud test as API key environment variable not set")
		return
	}
	shopID := os.Getenv("SHOP_ID")
	if shopID == "" {
		t.Logf("Skipping crypto cloud test as ShopId key environment variable not set")
		return
	}

	mockMessageBus := mocks.NewMockMessageBusService()
	service := crypto_cloud_api.NewCryptoCloudService(mockMessageBus)

	invoiceRequest := crypto_cloud_commands.IssueInvoiceCommand{
		AmountUSD: 100.0,
		Email:     "test@test.com",
	}

	response, err := service.IssueInvoice(invoiceRequest)

	assert.NoError(t, err, "Invoice creation should not return an error")
	assert.NotNil(t, response, "Response should not be nil")

	t.Logf("Invoice created successfully: %+v", response)
}
