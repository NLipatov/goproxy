package dto

import (
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
)

type PostInvoiceCommand struct {
	Email         string `json:"email"`
	OfferId       string `json:"offerId"`
	Periodicity   string `json:"periodicity"`
	Currency      string `json:"currency"`
	PaymentMethod string `json:"paymentMethod"`
	BuyerLanguage string `json:"buyerLanguage"`
}

func ToInvoiceDTO(invoice lavatopaggregates.Invoice) (PostInvoiceCommand, error) {
	email := invoice.Email().String()

	periodicity := invoice.Periodicity().String()

	currency := invoice.Currency().String()

	paymentMethod := invoice.PaymentMethod().String()

	buyerLanguage := invoice.BuyerLanguage().String()

	offerId := invoice.Offer().ExtId()

	return PostInvoiceCommand{
		Email:         email,
		OfferId:       offerId,
		Periodicity:   periodicity,
		Currency:      currency,
		PaymentMethod: paymentMethod,
		BuyerLanguage: buyerLanguage,
	}, nil
}
