package crypto_cloud_commands

import (
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api_interaction/crypto_cloud_currencies"
)

type IssueInvoiceCommand struct {
	Currency crypto_cloud_currencies.CryptoCloudCurrency
	Amount   float64
	Email    string
	OrderId  int
}
