package dto

import "goproxy/domain/lavatopsubdomain/lavatopvalueobjects"

type PriceDto struct {
	Currency    string  `json:"currency"`
	Amount      float64 `json:"amount"`
	Periodicity string  `json:"periodicity"`
}

type OfferResponse struct {
	Id     string     `json:"id"`
	Name   string     `json:"name"`
	Prices []PriceDto `json:"prices"`
}
type ProductResponse struct {
	ID     string          `json:"id"`
	Title  string          `json:"title"`
	Offers []OfferResponse `json:"offers"`
}
type GetOffersResponse struct {
	Items []ProductResponse `json:"items"`
}

func ToOfferResponse(offer lavatopvalueobjects.Offer) OfferResponse {
	prices := make([]PriceDto, len(offer.Prices()))
	for i, price := range offer.Prices() {
		prices[i] = ToPriceDto(price)
	}

	return OfferResponse{
		Id:     offer.ExtId(),
		Name:   offer.Name(),
		Prices: prices,
	}
}

func ToPriceDto(price lavatopvalueobjects.Price) PriceDto {
	return PriceDto{
		Currency:    price.Currency().String(),
		Amount:      float64(price.Cents()) / 100.0,
		Periodicity: price.Periodicity().String(),
	}
}
