package application

import (
	"goproxy/domain/contracts"
)

type BillingUseCases[T contracts.Invoice] interface {
	CreateInvoice(email, offerID, currency, language string) (T, error)
	GetInvoice(invoiceID string) (T, error)
	HandleWebhook(payload []byte) error
}
