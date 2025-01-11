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

	eurPrice := lavatopvalueobjects.NewPrice(48, lavatopvalueobjects.EUR, lavatopvalueobjects.ONE_TIME)
	rubPrice := lavatopvalueobjects.NewPrice(5000, lavatopvalueobjects.RUB, lavatopvalueobjects.ONE_TIME)
	usdPrice := lavatopvalueobjects.NewPrice(49, lavatopvalueobjects.USD, lavatopvalueobjects.ONE_TIME)
	prices := []lavatopvalueobjects.Price{eurPrice, rubPrice, usdPrice}

	offer := lavatopvalueobjects.NewOffer(
		"6c0cf730-3432-4755-941b-ca23b419d6df",
		"1 Month Plan",
		prices,
	)

	status, statusErr := lavatopvalueobjects.ParseInvoiceStatus("new")
	if statusErr != nil {
		t.Fatal(statusErr)
	}

	invoice, invoiceErr := lavatopaggregates.NewInvoice(
		1,
		2,
		"e624e74b-a109-4775-b8e2-be27ce89a0b8",
		status,
		email,
		offer,
		lavatopvalueobjects.ONE_TIME,
		lavatopvalueobjects.USD,
		lavatopvalueobjects.STRIPE,
		lavatopvalueobjects.EN,
	)
	if invoiceErr != nil {
		t.Fatal(invoiceErr)
	}

	expected := PostInvoiceCommand{
		Email:         "example@example.com",
		OfferId:       "6c0cf730-3432-4755-941b-ca23b419d6df",
		Periodicity:   "ONE_TIME",
		Currency:      "USD",
		PaymentMethod: "STRIPE",
		BuyerLanguage: "EN",
	}

	cmd, err := ToInvoiceDTO(invoice)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if cmd != expected {
		t.Errorf("Unexpected result: got %+v, expected %+v", cmd, expected)
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
