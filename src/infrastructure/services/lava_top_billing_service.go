package services

import (
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/domain/valueobjects"
)

type CreateInvoiceCommand struct {
	Email         valueobjects.Email
	OfferId       string
	Period        lavatopvalueobjects.Periodicity
	Currency      lavatopvalueobjects.Currency
	PaymentMethod lavatopvalueobjects.PaymentMethod
}

type LavaTopBillingService struct {
}

func NewLavaTopBillingService() LavaTopBillingService {
	return LavaTopBillingService{}
}
