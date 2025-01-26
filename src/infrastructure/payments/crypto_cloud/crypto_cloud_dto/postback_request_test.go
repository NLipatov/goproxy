package crypto_cloud_dto

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_PostBackRequest_Deserialization(t *testing.T) {
	requestJson := "{\n    \"status\": \"success\",\n    \"invoice_id\": \"H5LY5T6R\",\n    \"amount_crypto\": 100,\n    \"currency\": \"USDT_TRC20\",\n    \"order_id\": \"order_id\",\n    \"token\": \"eyJ0eXAiOiJKV1QiLCJhbGciOiJIAcI1NiJ9.eyJpZCI6MTMsImV4cCI6MTYzMTc4NjQyNn0.HQavV3z8dFnk56bX3MSY5X9lR6qVa9YhAoeTEHkaAzs\",\n    \"invoice_info\": {\n        \"uuid\": \"INV-H5LY5T6R\",\n        \"created\": \"2024-08-22 11:49:59.756692\",\n        \"address\": \"address\",\n        \"currency\": {\n            \"id\": 4,\n            \"code\": \"USDT\",\n            \"fullcode\": \"USDT_TRC20\",\n            \"network\": {\n                \"code\": \"TRC20\",\n                \"id\": 4,\n                \"icon\": \"https://cdn.cryptocloud.plus/currency/crypto/TRX.svg\",\n                \"fullname\": \"Tron\"\n            },\n            \"name\": \"Tether\",\n            \"is_email_required\": true,\n            \"stablecoin\": true,\n            \"icon_base\": \"https://cdn.cryptocloud.plus/currency/icons/main/usdt.svg\",\n            \"icon_network\": \"https://cdn.cryptocloud.plus/icons-currency/USDT-TRC20.svg\",\n            \"icon_qr\": \"https://cdn.cryptocloud.plus/currency/icons/stroke/usdt.svg\",\n            \"order\": 1\n        },\n        \"date_finished\": \"2024-08-22 11:51:53.753528\",\n        \"expiry_date\": \"2024-08-23 11:49:59.746385\",\n        \"side_commission\": \"client\",\n        \"type_payments\": \"crypto\",\n        \"amount\": 100,\n        \"amount_\": 100,\n        \"status\": \"overpaid\",\n        \"invoice_status\": \"success\",\n        \"is_email_required\": true,\n        \"project\": {\n            \"id\": 7,\n            \"name\": \"My Project\",\n            \"fail\": \"fail.com\",\n            \"success\": \"success.com\",\n            \"logo\": \"logo.com\"\n        },\n        \"tx_list\": [\n            \"\"\n        ],\n        \"amount_in_crypto\": null,\n        \"amount_in_fiat\": 100.0,\n        \"amount_usd\": 100.0,\n        \"amount_to_pay\": 102.0,\n        \"amount_to_pay_usd\": 102.0,\n        \"amount_paid\": 102.0,\n        \"amount_paid_usd\": 102.0,\n        \"fee\": 1.4,\n        \"fee_usd\": 1.4,\n        \"service_fee\": 0.8048,\n        \"service_fee_usd\": 0.8,\n        \"received\": 99.7952,\n        \"received_usd\": 99.8,\n        \"to_surcharge\": 0.2,\n        \"to_surcharge_usd\": 0.2,\n        \"total_rub\": 0,\n        \"step\": 3,\n        \"test_mode\": true,\n        \"type\": \"up\",\n        \"aml_enabled\": true,\n        \"aml_side\": \"merchant\",\n        \"aml_checks\": [\"one\", \"two\"],\n        \"links_invoice\": null\n    }\n}"

	var postbackRequest PostbackRequest
	err := json.Unmarshal([]byte(requestJson), &postbackRequest)

	assert.NoError(t, err, "Deserialization should not produce an error")

	// asserts all 'higher-level' fields
	assert.Equal(t, "success", postbackRequest.Status, "Status should match")
	assert.Equal(t, "H5LY5T6R", postbackRequest.InvoiceID, "InvoiceID should match")
	assert.Equal(t, 100.0, postbackRequest.AmountCrypto, "AmountCrypto should match")
	assert.Equal(t, "USDT_TRC20", postbackRequest.Currency, "Currency should match")
	assert.Equal(t, "order_id", postbackRequest.OrderID, "Order Id should match")
	assert.Equal(t, "eyJ0eXAiOiJKV1QiLCJhbGciOiJIAcI1NiJ9.eyJpZCI6MTMsImV4cCI6MTYzMTc4NjQyNn0.HQavV3z8dFnk56bX3MSY5X9lR6qVa9YhAoeTEHkaAzs", postbackRequest.Token, "Token should match")

	// asserts all InvoiceInfo fields
	invoiceInfo := postbackRequest.InvoiceInfo
	assert.Equal(t, "INV-H5LY5T6R", invoiceInfo.UUID, "InvoiceInfo UUID should match")
	assert.Equal(t, "2024-08-22 11:49:59.756692", invoiceInfo.Created, "Created date should match")
	assert.Equal(t, "address", invoiceInfo.Address, "Address should match")
	assert.Equal(t, "2024-08-22 11:51:53.753528", invoiceInfo.DateFinished, "Date finished should match")
	assert.Equal(t, "2024-08-23 11:49:59.746385", invoiceInfo.ExpiryDate, "Expiry Date should match")
	assert.Equal(t, "client", invoiceInfo.SideCommission, "SideCommission should match")
	assert.Equal(t, "crypto", invoiceInfo.TypePayments, "TypePayments should match")
	assert.Equal(t, 100.0, invoiceInfo.Amount, "Amount should match")
	assert.Equal(t, 100.0, invoiceInfo.Amount_, "Amount_ should match")
	assert.Equal(t, "overpaid", invoiceInfo.Status, "Status should match")
	assert.Equal(t, "success", invoiceInfo.InvoiceStatus, "InvoiceStatus should match")
	assert.Equal(t, true, invoiceInfo.IsEmailRequired, "Is Email Required should match")
	assert.Equal(t, 100.0, invoiceInfo.AmountInFiat, "Amount In Fiat should match")
	assert.Equal(t, 102.0, invoiceInfo.AmountToPay, "Amount To Pay should match")
	assert.Equal(t, 102.0, invoiceInfo.AmountToPayUSD, "Amount To Pay USD should match")
	assert.Equal(t, 1.4, invoiceInfo.Fee, "Fee should match")
	assert.Equal(t, 1.4, invoiceInfo.FeeUSD, "Fee USD should match")
	assert.Equal(t, 0.8048, invoiceInfo.ServiceFee, "Service Fee should match")
	assert.Equal(t, 0.8, invoiceInfo.ServiceFeeUSD, "Service Fee USD should match")
	assert.Equal(t, 99.7952, invoiceInfo.Received, "Received should match")
	assert.Equal(t, 99.8, invoiceInfo.ReceivedUSD, "Received USD should match")
	assert.Equal(t, 0.2, invoiceInfo.ToSurcharge, "To Surcharge should match")
	assert.Equal(t, 0.2, invoiceInfo.ToSurchargeUSD, "To Surcharge USD should match")
	assert.Equal(t, 0, invoiceInfo.TotalRub, "Total Rub should match")
	assert.Equal(t, 3, invoiceInfo.Step, "Step should match")
	assert.Equal(t, true, invoiceInfo.TestMode, "TestMode should match")
	assert.Equal(t, "up", invoiceInfo.Type, "Type should match")
	assert.Equal(t, true, invoiceInfo.AMLEnabled, "AML Enabled should match")
	assert.Equal(t, "merchant", invoiceInfo.AMLSide, "AML Side should match")
	assert.Equal(t, []string{"one", "two"}, invoiceInfo.AMLChecks, "AML Checks should match")

	// assert all Currency fields
	currency := invoiceInfo.Currency
	assert.Equal(t, 4, currency.ID, "Currency ID should match")
	assert.Equal(t, "USDT", currency.Code, "Currency Code should match")
	assert.Equal(t, "USDT_TRC20", currency.FullCode, "Currency FullCode should match")
	assert.Equal(t, "Tether", currency.Name, "Currency Name should match")
	assert.Equal(t, true, currency.IsEmailRequired, "Currency IsEmailRequired should match")
	assert.Equal(t, true, currency.StableCoin, "Currency StableCoin should match")
	assert.Equal(t, "https://cdn.cryptocloud.plus/currency/icons/main/usdt.svg", currency.IconBase, "Icon Base should match")
	assert.Equal(t, "https://cdn.cryptocloud.plus/icons-currency/USDT-TRC20.svg", currency.IconNetwork, "Icon Network should match")
	assert.Equal(t, "https://cdn.cryptocloud.plus/currency/icons/stroke/usdt.svg", currency.IconQR, "Icon QR should match")
	assert.Equal(t, 1, currency.Order, "Order should match")

	// assert all Network fields
	network := currency.Network
	assert.Equal(t, "TRC20", network.Code, "Network Code should match")
	assert.Equal(t, 4, network.ID, "Network ID should match")
	assert.Equal(t, "https://cdn.cryptocloud.plus/currency/crypto/TRX.svg", network.Icon, "Network Icon should match")
	assert.Equal(t, "Tron", network.FullName, "Network FullName should match")

	// assert all Project fields
	project := invoiceInfo.Project
	assert.Equal(t, 7, project.ID, "Project ID should match")
	assert.Equal(t, "My Project", project.Name, "Project Name should match")
	assert.Equal(t, "fail.com", project.FailURL, "Project FailURL should match")
	assert.Equal(t, "success.com", project.SuccessURL, "Project SuccessURL should match")
	assert.Equal(t, "logo.com", project.Logo, "Project Logo should match")
}
