package application

import (
	"fmt"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
)

type LavaTopUseCases struct {
	billingService BillingService[lavatopaggregates.Invoice, lavatopaggregates.Offer]
}

func NewLavaTopUseCases(billingService BillingService[lavatopaggregates.Invoice, lavatopaggregates.Offer]) LavaTopUseCases {
	return LavaTopUseCases{
		billingService: billingService,
	}
}

func (l *LavaTopUseCases) GetOffers() ([]lavatopaggregates.Offer, error) {
	return l.billingService.GetOffers()
}

func (l *LavaTopUseCases) PublishInvoice(invoice lavatopaggregates.Invoice, userId int) (lavatopaggregates.Invoice, error) {
	publishedInvoice, err := l.billingService.PublishInvoice(invoice)
	if err != nil {
		return lavatopaggregates.Invoice{}, err
	}

	//save published invoice to db and associate it with userId
	fmt.Printf("Published Invoice with ID: %d, user id: %d", publishedInvoice.Id(), userId)
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
