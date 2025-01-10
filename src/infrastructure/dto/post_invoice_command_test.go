package dto

import (
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/domain/valueobjects"
	"testing"
)

func TestToInvoiceDTO(t *testing.T) {
	email, emailErr := valueobjects.ParseEmailFromString("example@example.com")
	if emailErr != nil {
		t.Fatal(emailErr)
	}

	offerId, offerIdErr := valueobjects.ParseGuidFromString("6c0cf730-3432-4755-941b-ca23b419d6df")
	if offerIdErr != nil {
		t.Fatal(offerIdErr)
	}

	invoice, invoiceErr := lavatopaggregates.NewInvoice(
		1,
		email,
		offerId,
		lavatopvalueobjects.ONE_TIME,
		lavatopvalueobjects.USD,
		lavatopvalueobjects.STRIPE,
		lavatopvalueobjects.EN,
	)
	if invoiceErr != nil {
		t.Fatal(invoiceErr)
	}

	cmd, err := ToInvoiceDTO(invoice)
	if err != nil {
		t.Fatal(err)
	}

	if cmd.Email != invoice.Email().String() {
		t.Errorf("Email expected %s, got %s", invoice.Email().String(), cmd.Email)
	}

	if cmd.OfferId != invoice.OfferId().String() {
		t.Errorf("OfferId expected %s, got %s", invoice.OfferId().String(), cmd.OfferId)
	}

	if cmd.Periodicity != invoice.Periodicity().String() {
		t.Errorf("Periodicity expected %s, got %s", invoice.Periodicity().String(), cmd.Periodicity)
	}

	tests := []struct {
		name        string
		invoice     lavatopaggregates.Invoice
		expected    PostInvoiceCommand
		expectError bool
	}{
		{
			name:    "Valid Invoice",
			invoice: invoice,
			expected: PostInvoiceCommand{
				Email:         "example@example.com",
				OfferId:       "6c0cf730-3432-4755-941b-ca23b419d6df",
				Periodicity:   "ONE_TIME",
				Currency:      "USD",
				PaymentMethod: "STRIPE",
				BuyerLanguage: "EN",
			},
			expectError: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			invoiceDTO, invoiceDTOErr := ToInvoiceDTO(test.invoice)
			if (invoiceDTOErr != nil) != test.expectError {
				t.Fatalf("Unexpected error: got %v, expected error: %v", invoiceDTOErr, test.expectError)
			}

			if !test.expectError && invoiceDTO != test.expected {
				t.Errorf("Unexpected invoiceDTO: got %+v, expected %+v", invoiceDTO, test.expected)
			}
		})
	}
}
