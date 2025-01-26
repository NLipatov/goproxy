package crypto_cloud

import (
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
)

type PaymentProvider interface {
	IssueInvoice(command crypto_cloud_commands.IssueInvoiceCommand) (interface{}, error)
	HandlePostBack(command crypto_cloud_commands.PostBackCommand) error
}
