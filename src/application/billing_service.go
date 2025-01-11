package application

import (
	"goproxy/domain/contracts"
)

type BillingService[T contracts.Invoice] interface {
	GetInvoiceStatus(offerId string) (string, error)
	PublishInvoice(invoice T) (T, error)
}
