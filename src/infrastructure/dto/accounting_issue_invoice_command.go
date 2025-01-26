package dto

type AccountingIssueInvoiceCommand struct {
	OfferId       string `json:"offer_id"`
	Currency      string `json:"currency"`
	PaymentMethod string `json:"payment_method"`
}
