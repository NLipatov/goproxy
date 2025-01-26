package crypto_cloud_billing_dto

type IssueInvoiceCommandDto struct {
	AmountUSD float64 `json:"amount_usd"`
	Email     string  `json:"email"`
	OrderId   string  `json:"order_id"`
}
