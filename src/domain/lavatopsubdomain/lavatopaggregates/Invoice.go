package lavatopaggregates

import (
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/domain/valueobjects"
)

type Invoice struct {
	id            int
	userId        int
	extId         string
	status        lavatopvalueobjects.Status
	email         valueobjects.Email
	offer         lavatopvalueobjects.Offer
	periodicity   lavatopvalueobjects.Periodicity
	currency      lavatopvalueobjects.Currency
	paymentMethod lavatopvalueobjects.PaymentMethod
	buyerLanguage lavatopvalueobjects.BuyerLanguage
}

func NewInvoice(
	id int,
	userId int,
	externalId string, // invoice identifier in lava top system
	status lavatopvalueobjects.Status,
	email valueobjects.Email,
	offer lavatopvalueobjects.Offer,
	periodicity lavatopvalueobjects.Periodicity,
	currency lavatopvalueobjects.Currency,
	paymentMethod lavatopvalueobjects.PaymentMethod,
	buyerLanguage lavatopvalueobjects.BuyerLanguage) (Invoice, error) {

	return Invoice{
		id:            id,
		userId:        userId,
		extId:         externalId,
		status:        status,
		email:         email,
		offer:         offer,
		periodicity:   periodicity,
		currency:      currency,
		paymentMethod: paymentMethod,
		buyerLanguage: buyerLanguage,
	}, nil
}

func (i Invoice) Id() int {
	return i.id
}

func (i Invoice) UserId() int {
	return i.userId
}

func (i Invoice) ExtId() string {
	return i.extId
}

func (i Invoice) Status() lavatopvalueobjects.Status {
	return i.status
}

func (i Invoice) Email() valueobjects.Email {
	return i.email
}

func (i Invoice) Offer() lavatopvalueobjects.Offer {
	return i.offer
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
