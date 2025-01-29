package crypto_cloud_api_dto

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CreateInvoice_Deserialization(t *testing.T) {
	responseJSON := "{\"status\":\"success\",\"result\":{\"uuid\":\"INV-H5LY5T6R\",\"created\":\"2025-01-26 13:05:50.381108\",\"address\":\"1600 Pennsylvania Avenue NW in Washington, D.C., USA\",\"expiry_date\":\"2025-01-27 13:05:50.367824\",\"side_commission\":\"merchant\",\"side_commission_service\":\"merchant\",\"type_payments\":\"crypto\",\"amount\":100.0,\"amount_usd\":100.0,\"amount_in_fiat\":100.0,\"fee\":1.4,\"fee_usd\":1.4,\"service_fee\":1.9,\"service_fee_usd\":1.9,\"fiat_currency\":\"USD\",\"status\":\"created\",\"is_email_required\":false,\"link\":\"https://pay.cryptocloud.plus/H5LY5T6R\",\"invoice_id\":null,\"currency\":{\"id\":4,\"code\":\"USDT\",\"fullcode\":\"USDT_TRC20\",\"network\":{\"code\":\"TRC20\",\"id\":4,\"icon\":\"https://cdn.cryptocloud.plus/img/network/TRC20.svg\",\"fullname\":\"Tron\"},\"name\":\"Tether\",\"is_email_required\":false,\"stablecoin\":true,\"icon_base\":\"https://cdn.cryptocloud.plus/img/currency/USDT.svg\",\"icon_network\":\"https://cdn.cryptocloud.plus/img/currency_network/USDT_TRC.svg\",\"icon_qr\":\"https://cdn.cryptocloud.plus/img/stroke/USDT_STROKE.svg\",\"order\":1},\"project\":{\"id\":362100,\"name\":\"proxy.ethacore\",\"fail\":\"https://proxy.ethacore/\",\"success\":\"https://proxy.ethacore/\",\"logo\":\"https://cdn.example.com/project-logo\"},\"test_mode\":true}}"

	var apiResponse InvoiceResponse
	err := json.Unmarshal([]byte(responseJSON), &apiResponse)

	assert.NoError(t, err, "Deserialization should not produce an error")
	assert.Equal(t, "success", apiResponse.Status, "Status should be 'success'")

	// asserts all result fields
	assert.Equal(t, "INV-H5LY5T6R", apiResponse.Result.UUID, "UUID should match")
	assert.Equal(t, "2025-01-26 13:05:50.381108", apiResponse.Result.Created, "Created time should match")
	assert.Equal(t, "1600 Pennsylvania Avenue NW in Washington, D.C., USA", apiResponse.Result.Address, "Address should match")
	assert.Equal(t, "2025-01-27 13:05:50.367824", apiResponse.Result.ExpiryDate, "ExpiryDate should match")
	assert.Equal(t, "merchant", apiResponse.Result.SideCommission, "SideCommission should match")
	assert.Equal(t, "merchant", apiResponse.Result.SideCommissionService, "SideCommissionService should match")
	assert.Equal(t, "crypto", apiResponse.Result.TypePayments, "TypePayments should match")
	assert.Equal(t, 100.0, apiResponse.Result.Amount, "Amount should match")
	assert.Equal(t, 100.0, apiResponse.Result.AmountUSD, "AmountUSD should match")
	assert.Equal(t, 100.0, apiResponse.Result.AmountInFiat, "AmountInFiat should match")
	assert.Equal(t, 1.4, apiResponse.Result.Fee, "Fee should match")
	assert.Equal(t, 1.4, apiResponse.Result.FeeUSD, "FeeUSD should match")
	assert.Equal(t, 1.9, apiResponse.Result.ServiceFee, "ServiceFee should match")
	assert.Equal(t, 1.9, apiResponse.Result.ServiceFeeUSD, "ServiceFeeUSD should match")
	assert.Equal(t, "USD", apiResponse.Result.FiatCurrency, "FiatCurrency should match")
	assert.Equal(t, "created", apiResponse.Result.Status, "Status should match")
	assert.False(t, apiResponse.Result.IsEmailRequired, "IsEmailRequired should be false")
	assert.Equal(t, "https://pay.cryptocloud.plus/H5LY5T6R", apiResponse.Result.Link, "Link should match")

	// asserts all currency fields
	currency := apiResponse.Result.Currency
	assert.Equal(t, 4, currency.ID, "Currency ID should match")
	assert.Equal(t, "USDT", currency.Code, "Currency code should match")
	assert.Equal(t, "USDT_TRC20", currency.FullCode, "Currency fullcode should match")
	assert.Equal(t, "TRC20", currency.Network.Code, "Network code should match")
	assert.Equal(t, 4, currency.Network.ID, "Network ID should match")
	assert.Equal(t, "https://cdn.cryptocloud.plus/img/network/TRC20.svg", currency.Network.Icon, "Network icon should match")
	assert.Equal(t, "Tron", currency.Network.FullName, "Network fullname should match")
	assert.Equal(t, "Tether", currency.Name, "Currency name should match")
	assert.False(t, currency.IsEmailRequired, "Currency IsEmailRequired should be false")
	assert.True(t, currency.StableCoin, "Currency Stablecoin should be true")
	assert.Equal(t, "https://cdn.cryptocloud.plus/img/currency/USDT.svg", currency.IconBase, "Currency IconBase should match")
	assert.Equal(t, "https://cdn.cryptocloud.plus/img/currency_network/USDT_TRC.svg", currency.IconNetwork, "Currency IconNetwork should match")
	assert.Equal(t, "https://cdn.cryptocloud.plus/img/stroke/USDT_STROKE.svg", currency.IconQR, "Currency IconQR should match")
	assert.Equal(t, 1, currency.Order, "Currency order should match")

	// asserts all project fields
	project := apiResponse.Result.Project
	assert.Equal(t, 362100, project.ID, "Project ID should match")
	assert.Equal(t, "proxy.ethacore", project.Name, "Project name should match")
	assert.Equal(t, "https://proxy.ethacore/", project.FailURL, "Project fail URL should match")
	assert.Equal(t, "https://proxy.ethacore/", project.SuccessURL, "Project success URL should match")
	assert.Equal(t, "https://cdn.example.com/project-logo", project.Logo, "Project logo should match")

	// Проверка test_mode
	assert.True(t, apiResponse.Result.TestMode, "TestMode should be true")
}
