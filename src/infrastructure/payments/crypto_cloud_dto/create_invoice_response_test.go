package crypto_cloud_dto

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_CreateInvoice_Deserialization(t *testing.T) {
	responseJSON := "{\"status\":\"success\",\"result\":{\"uuid\":\"INV-H5LY5T6R\",\"created\":\"2025-01-26 13:05:50.381108\",\"address\":\"\",\"expiry_date\":\"2025-01-27 13:05:50.367824\",\"side_commission\":\"merchant\",\"side_commission_service\":\"merchant\",\"type_payments\":\"crypto\",\"amount\":100.0,\"amount_usd\":100.0,\"amount_in_fiat\":100.0,\"fee\":1.4,\"fee_usd\":1.4,\"service_fee\":1.9,\"service_fee_usd\":1.9,\"fiat_currency\":\"USD\",\"status\":\"created\",\"is_email_required\":false,\"link\":\"https://pay.cryptocloud.plus/H5LY5T6R\",\"invoice_id\":null,\"currency\":{\"id\":4,\"code\":\"USDT\",\"fullcode\":\"USDT_TRC20\",\"network\":{\"code\":\"TRC20\",\"id\":4,\"icon\":\"https://cdn.cryptocloud.plus/img/network/TRC20.svg\",\"fullname\":\"Tron\"},\"name\":\"Tether\",\"is_email_required\":false,\"stablecoin\":true,\"icon_base\":\"https://cdn.cryptocloud.plus/img/currency/USDT.svg\",\"icon_network\":\"https://cdn.cryptocloud.plus/img/currency_network/USDT_TRC.svg\",\"icon_qr\":\"https://cdn.cryptocloud.plus/img/stroke/USDT_STROKE.svg\",\"order\":1},\"project\":{\"id\":362100,\"name\":\"proxy.ethacore\",\"fail\":\"https://proxy.ethacore/\",\"success\":\"https://proxy.ethacore/\",\"logo\":\"\"},\"test_mode\":true}}"

	var apiResponse InvoiceResponse
	err := json.Unmarshal([]byte(responseJSON), &apiResponse)

	assert.NoError(t, err, "Deserialization should not produce an error")
	assert.Equal(t, "success", apiResponse.Status, "Status should be 'success'")
	assert.Equal(t, "INV-H5LY5T6R", apiResponse.Result.UUID, "UUID should match")
	assert.Equal(t, 100.0, apiResponse.Result.Amount, "Amount should match")
	assert.Equal(t, "proxy.ethacore", apiResponse.Result.Project.Name, "Project name should match")
	assert.Truef(t, apiResponse.Result.TestMode, "TestMode should be true")
}
