package crypto_cloud_dto

type InvoiceResponse struct {
	Status string `json:"status"`
	Result struct {
		UUID                  string  `json:"uuid"`
		Created               string  `json:"created"`
		Address               string  `json:"address,omitempty"`
		ExpiryDate            string  `json:"expiry_date"`
		SideCommission        string  `json:"side_commission"`
		SideCommissionService string  `json:"side_commission_service"`
		TypePayments          string  `json:"type_payments"`
		Amount                float64 `json:"amount"`
		AmountUSD             float64 `json:"amount_usd"`
		AmountInFiat          float64 `json:"amount_in_fiat"`
		Fee                   float64 `json:"fee"`
		FeeUSD                float64 `json:"fee_usd"`
		ServiceFee            float64 `json:"service_fee"`
		ServiceFeeUSD         float64 `json:"service_fee_usd"`
		FiatCurrency          string  `json:"fiat_currency"`
		Status                string  `json:"status"`
		IsEmailRequired       bool    `json:"is_email_required"`
		Link                  string  `json:"link"`
		Currency              struct {
			ID       int    `json:"id"`
			Code     string `json:"code"`
			FullCode string `json:"fullcode"`
			Network  struct {
				Code     string `json:"code"`
				ID       int    `json:"id"`
				Icon     string `json:"icon"`
				FullName string `json:"fullname"`
			} `json:"network"`
			Name            string `json:"name"`
			IsEmailRequired bool   `json:"is_email_required"`
			StableCoin      bool   `json:"stablecoin"`
			IconBase        string `json:"icon_base"`
			IconNetwork     string `json:"icon_network"`
			IconQR          string `json:"icon_qr"`
			Order           int    `json:"order"`
		} `json:"currency"`
		Project struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			FailURL    string `json:"fail"`
			SuccessURL string `json:"success"`
			Logo       string `json:"logo"`
		} `json:"project"`
		TestMode bool `json:"test_mode"`
	} `json:"result"`
}
