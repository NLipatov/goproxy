package crypto_cloud_billing_dto

type IssueInvoiceCommandDto struct {
	PlanId   int    `json:"plan_id"`
	Currency string `json:"currency"`
	Email    string `json:"email"`
}
