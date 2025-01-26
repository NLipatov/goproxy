package crypto_cloud_commands

type PostBackCommand struct {
	OrderID string `json:"order_id"`
	Token   string `json:"token"`
	Secret  string `json:"secret"`
}
