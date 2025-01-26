package crypto_cloud

import (
	crypto_cloud_commands2 "goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
)

type PaymentProvider interface {
	IssueInvoice(command crypto_cloud_commands2.IssueInvoiceCommand) (interface{}, error)
	HandlePostBack(command crypto_cloud_commands2.PostBackCommand) (interface{}, error)
}
