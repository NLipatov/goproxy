package commands

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
