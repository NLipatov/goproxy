package lavatopaggregates

import (
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/domain/valueobjects"
)

type Invoice struct {
	id            int
	email         valueobjects.Email
	offerId       valueobjects.Guid
	periodicity   lavatopvalueobjects.Periodicity
	currency      lavatopvalueobjects.Currency
	paymentMethod lavatopvalueobjects.PaymentMethod
	buyerLanguage lavatopvalueobjects.BuyerLanguage
}

func NewInvoice(
	id int,
	email valueobjects.Email,
	offerId valueobjects.Guid,
	periodicity lavatopvalueobjects.Periodicity,
	currency lavatopvalueobjects.Currency,
	paymentMethod lavatopvalueobjects.PaymentMethod,
	buyerLanguage lavatopvalueobjects.BuyerLanguage) (Invoice, error) {

	return Invoice{
		id:            id,
		email:         email,
		offerId:       offerId,
		periodicity:   periodicity,
		currency:      currency,
		paymentMethod: paymentMethod,
		buyerLanguage: buyerLanguage,
	}, nil
}

func (i Invoice) Id() int {
	return i.id
}

func (i Invoice) Email() valueobjects.Email {
	return i.email
}

func (i Invoice) OfferId() valueobjects.Guid {
	return i.offerId
}

func (i Invoice) Periodicity() lavatopvalueobjects.Periodicity {
	return i.periodicity
}

func (i Invoice) Currency() lavatopvalueobjects.Currency {
	return i.currency
}

func (i Invoice) PaymentMethod() lavatopvalueobjects.PaymentMethod {
	return i.paymentMethod
}

func (i Invoice) BuyerLanguage() lavatopvalueobjects.BuyerLanguage {
	return i.buyerLanguage
}
