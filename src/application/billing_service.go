package application

import (
	"goproxy/domain/contracts"
)

type BillingService[T contracts.Invoice] interface {
	PublishInvoice(invoice T) (T, error)
}
