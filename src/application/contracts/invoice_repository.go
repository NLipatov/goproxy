package contracts

import (
	"goproxy/domain/contracts"
)

type InvoiceRepository[T contracts.Invoice] interface {
	SaveInvoice(invoice T) (T, error)
}
