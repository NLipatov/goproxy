package application

import (
	"fmt"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
)

type LavaTopUseCases struct {
	invoiceRepository InvoiceRepository[lavatopaggregates.Invoice]
	billingService    BillingService[lavatopaggregates.Invoice, lavatopvalueobjects.Offer]
}

func NewLavaTopUseCases(billingService BillingService[lavatopaggregates.Invoice, lavatopvalueobjects.Offer]) LavaTopUseCases {
	return LavaTopUseCases{
		billingService: billingService,
	}
}

func (l *LavaTopUseCases) GetOffers() ([]lavatopvalueobjects.Offer, error) {
	return l.billingService.GetOffers()
}

func (l *LavaTopUseCases) PublishInvoice(invoice lavatopaggregates.Invoice) (lavatopaggregates.Invoice, error) {
	publishedInvoice, err := l.billingService.PublishInvoice(invoice)
	if err != nil {
		return lavatopaggregates.Invoice{}, err
	}

	//save published invoice to db and associate it with userId
	fmt.Printf("Published Invoice with ID: %d, user id: %d", publishedInvoice.Id(), invoice.UserId())
	panic("not implemented")
}

func (l *LavaTopUseCases) GetInvoiceStatus(offerId string) (string, error) {
	status, statusErr := l.billingService.GetInvoiceStatus(offerId)
	if statusErr != nil {
		return "", statusErr
	}

	//update db status
	return status, nil
}
