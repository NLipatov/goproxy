package dto

type InvoicePaymentParamsResponse struct {
	Id          string         `json:"id"`
	Status      string         `json:"status"`
	AmountTotal AmountTotalDto `json:"amountTotal"`
	PaymentURL  string         `json:"paymentUrl,omitempty"`
}

type AmountTotalDto struct {
	Currency string `json:"currency"`
	Amount   int    `json:"amount"`
}

type ErrorResponse struct {
	Error     string            `json:"error"`
	Details   map[string]string `json:"details"`
	Timestamp string            `json:"timestamp"`
}
