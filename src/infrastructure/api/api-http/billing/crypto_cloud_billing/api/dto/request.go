package dto

type Request struct {
	PlanId   int    `json:"plan_id"`
	Currency string `json:"currency"`
}
