package crypto_cloud_api_dto

// PostbackRequest comes from CryptoCloud when, for example, invoice was paid
type PostbackRequest struct {
	Status       string  `json:"status"`
	InvoiceID    string  `json:"invoice_id"`
	AmountCrypto float64 `json:"amount_crypto"`
	Currency     string  `json:"currency"`
	OrderID      string  `json:"order_id"`
	Token        string  `json:"token"`
	InvoiceInfo  struct {
		UUID     string `json:"uuid"`
		Created  string `json:"created"`
		Address  string `json:"address"`
		Currency struct {
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
		DateFinished    string  `json:"date_finished"`
		ExpiryDate      string  `json:"expiry_date"`
		SideCommission  string  `json:"side_commission"`
		TypePayments    string  `json:"type_payments"`
		Amount          float64 `json:"amount"`
		Amount_         float64 `json:"amount_"`
		Status          string  `json:"status"`
		InvoiceStatus   string  `json:"invoice_status"`
		IsEmailRequired bool    `json:"is_email_required"`
		Project         struct {
			ID         int    `json:"id"`
			Name       string `json:"name"`
			FailURL    string `json:"fail"`
			SuccessURL string `json:"success"`
			Logo       string `json:"logo"`
		} `json:"project"`
		TxList         []string `json:"tx_list"`
		AmountInCrypto *float64 `json:"amount_in_crypto"`
		AmountInFiat   float64  `json:"amount_in_fiat"`
		AmountUSD      float64  `json:"amount_usd"`
		AmountToPay    float64  `json:"amount_to_pay"`
		AmountToPayUSD float64  `json:"amount_to_pay_usd"`
		AmountPaid     float64  `json:"amount_paid"`
		AmountPaidUSD  float64  `json:"amount_paid_usd"`
		Fee            float64  `json:"fee"`
		FeeUSD         float64  `json:"fee_usd"`
		ServiceFee     float64  `json:"service_fee"`
		ServiceFeeUSD  float64  `json:"service_fee_usd"`
		Received       float64  `json:"received"`
		ReceivedUSD    float64  `json:"received_usd"`
		ToSurcharge    float64  `json:"to_surcharge"`
		ToSurchargeUSD float64  `json:"to_surcharge_usd"`
		TotalRub       int      `json:"total_rub"`
		Step           int      `json:"step"`
		TestMode       bool     `json:"test_mode"`
		Type           string   `json:"type"`
		AMLEnabled     bool     `json:"aml_enabled"`
		AMLSide        string   `json:"aml_side"`
		AMLChecks      []string `json:"aml_checks"`
		LinksInvoice   *string  `json:"links_invoice"`
	} `json:"invoice_info"`
}
