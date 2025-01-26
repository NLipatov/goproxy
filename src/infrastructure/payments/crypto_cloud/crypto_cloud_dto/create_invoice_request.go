package crypto_cloud_dto

type InvoiceRequest struct {
	ShopID    string            `json:"shop_id"`
	Amount    float64           `json:"amount"`
	Currency  string            `json:"currency,omitempty"`
	Email     string            `json:"email,omitempty"`
	AddFields map[string]string `json:"add_fields,omitempty"`
	OrderID   string            `json:"order_id,omitempty"`
}
