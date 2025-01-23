package cache_serialization

import (
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
)

type LavaTopOfferDto struct {
	ExtId  string
	Name   string
	Prices []LavaTopOfferPriceDto
}

type LavaTopOfferPriceDto struct {
	Cents       int64
	Periodicity lavatopvalueobjects.Periodicity
	Currency    lavatopvalueobjects.Currency
}

type LavaTopOfferCacheSerializer struct {
}

func NewLavaTopOfferCacheSerializer() CacheSerializer[lavatopvalueobjects.Offer, LavaTopOfferDto] {
	return &LavaTopOfferCacheSerializer{}
}

func (l *LavaTopOfferCacheSerializer) ToT(dto LavaTopOfferDto) lavatopvalueobjects.Offer {
	prices := make([]lavatopvalueobjects.Price, len(dto.Prices))
	for i, p := range dto.Prices {
		prices[i] = lavatopvalueobjects.NewPrice(p.Cents, p.Currency, p.Periodicity)
	}

	return lavatopvalueobjects.NewOffer(dto.ExtId, dto.Name, prices)
}
func (l *LavaTopOfferCacheSerializer) ToD(offer lavatopvalueobjects.Offer) LavaTopOfferDto {
	prices := make([]LavaTopOfferPriceDto, len(offer.Prices()))

	for i, p := range offer.Prices() {
		prices[i] = LavaTopOfferPriceDto{
			Cents:       p.Cents(),
			Periodicity: p.Periodicity(),
			Currency:    p.Currency(),
		}
	}

	return LavaTopOfferDto{
		ExtId:  offer.ExtId(),
		Name:   offer.Name(),
		Prices: prices,
	}
}
func (l *LavaTopOfferCacheSerializer) ToTArray(dto []LavaTopOfferDto) []lavatopvalueobjects.Offer {
	result := make([]lavatopvalueobjects.Offer, len(dto))
	for i, d := range dto {
		result[i] = l.ToT(d)
	}
	return result
}
func (l *LavaTopOfferCacheSerializer) ToDArray(offer []lavatopvalueobjects.Offer) []LavaTopOfferDto {
	result := make([]LavaTopOfferDto, len(offer))
	for i, d := range offer {
		result[i] = l.ToD(d)
	}
	return result
}
