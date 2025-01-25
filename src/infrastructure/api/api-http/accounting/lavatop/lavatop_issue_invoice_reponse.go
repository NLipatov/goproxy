package lavatop

import (
	"fmt"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/dto"
)

type IssueInvoiceResponse struct {
	lavaTopUseCases application.LavaTopUseCases
	user            aggregates.User
	offerId         string
	currency        lavatopvalueobjects.Currency
	paymentMethod   lavatopvalueobjects.PaymentMethod
}

func NewIssueInvoiceResponse(lavaTopUseCases application.LavaTopUseCases,
	user aggregates.User, currency lavatopvalueobjects.Currency, paymentMethod lavatopvalueobjects.PaymentMethod,
	offerId string) IssueInvoiceResponse {
	return IssueInvoiceResponse{
		lavaTopUseCases: lavaTopUseCases,
		user:            user,
		offerId:         offerId,
		currency:        currency,
		paymentMethod:   paymentMethod,
	}
}

func (i IssueInvoiceResponse) Build() (dto.ApiResponse[dto.PostInvoiceResponse], error) {
	email, emailErr := valueobjects.ParseEmailFromString(i.user.Email())
	if emailErr != nil {
		return dto.ApiResponse[dto.PostInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    404,
			ErrorMessage: "invalid email",
		}, emailErr
	}

	offers, offersErr := i.lavaTopUseCases.GetOffers()
	if offersErr != nil {
		return dto.ApiResponse[dto.PostInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    500,
			ErrorMessage: "could not load offers",
		}, offersErr
	}

	status, statusErr := lavatopvalueobjects.ParseInvoiceStatus("new")
	if statusErr != nil {
		return dto.ApiResponse[dto.PostInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    500,
			ErrorMessage: "failed to parse invoice status",
		}, statusErr
	}

	for _, offer := range offers {
		if offer.ExtId() == i.offerId {
			invoice, _ := lavatopaggregates.NewInvoice(
				-1,
				i.user.Id(),
				offer.ExtId(),
				status,
				email,
				offer,
				lavatopvalueobjects.ONE_TIME,
				i.currency,
				i.paymentMethod,
				lavatopvalueobjects.EN,
			)
			publishedInvoice, publishedInvoiceErr := i.lavaTopUseCases.PublishInvoice(invoice)
			if publishedInvoiceErr != nil {
				return dto.ApiResponse[dto.PostInvoiceResponse]{
					Payload:      nil,
					ErrorCode:    500,
					ErrorMessage: "could not issue invoice",
				}, publishedInvoiceErr
			}

			fmt.Println("Published Invoice", publishedInvoice)
		}
	}

	return dto.ApiResponse[dto.PostInvoiceResponse]{
		Payload:      nil,
		ErrorCode:    400,
		ErrorMessage: "invalid offer id",
	}, fmt.Errorf("invalid offer id")
}
