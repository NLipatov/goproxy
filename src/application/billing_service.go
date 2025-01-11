package application

import (
	"goproxy/domain/contracts"
)

type BillingService[I contracts.Invoice, O contracts.Offer] interface {
	GetOffers() ([]O, error)
	PublishInvoice(invoice I) (I, error)
	GetInvoiceStatus(offerId string) (string, error)
}
