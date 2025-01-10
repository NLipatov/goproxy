package application

import (
	"fmt"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
)

type LavaTopUseCases struct {
	billingService BillingService[lavatopaggregates.Invoice]
}

func NewLavaTopUseCases(billingService BillingService[lavatopaggregates.Invoice]) LavaTopUseCases {
	return LavaTopUseCases{
		billingService: billingService,
	}
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
