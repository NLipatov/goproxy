package application

import (
	"fmt"
	"goproxy/dal/cache_serialization"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"time"
)

const (
	lavaTopUseCasesCacheTtl = time.Hour * 12
)

type LavaTopUseCases struct {
	invoiceRepository    InvoiceRepository[lavatopaggregates.Invoice]
	billingService       BillingService[lavatopaggregates.Invoice, lavatopvalueobjects.Offer]
	cache                CacheWithTTL[[]cache_serialization.LavaTopOfferDto]
	offerCacheSerializer cache_serialization.CacheSerializer[lavatopvalueobjects.Offer, cache_serialization.LavaTopOfferDto]
}

func NewLavaTopUseCases(billingService BillingService[lavatopaggregates.Invoice, lavatopvalueobjects.Offer],
	cache CacheWithTTL[[]cache_serialization.LavaTopOfferDto]) LavaTopUseCases {
	return LavaTopUseCases{
		billingService:       billingService,
		cache:                cache,
		offerCacheSerializer: cache_serialization.NewLavaTopOfferCacheSerializer(),
	}
}

func (l *LavaTopUseCases) GetOffers() ([]lavatopvalueobjects.Offer, error) {
	cached, cachedErr := l.cache.Get("lavatop_use_cases:get_offers")
	if cachedErr == nil {
		return l.offerCacheSerializer.ToTArray(cached), nil
	}

	offers, offersErr := l.billingService.GetOffers()
	if offersErr != nil {
		return nil, offersErr
	}

	_ = l.cache.Set("lavatop_use_cases:get_offers", l.offerCacheSerializer.ToDArray(offers))
	_ = l.cache.Expire("lavatop_use_cases:get_offers", lavaTopUseCasesCacheTtl)

	return offers, nil
}

func (l *LavaTopUseCases) PublishInvoice(invoice lavatopaggregates.Invoice) (lavatopaggregates.Invoice, error) {
	publishedInvoice, err := l.billingService.PublishInvoice(invoice)
	if err != nil {
		return lavatopaggregates.Invoice{}, err
	}

	//save published invoice to db and associate it with userId
	fmt.Printf("Published Invoice with Id: %d, user id: %d", publishedInvoice.Id(), invoice.UserId())
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
