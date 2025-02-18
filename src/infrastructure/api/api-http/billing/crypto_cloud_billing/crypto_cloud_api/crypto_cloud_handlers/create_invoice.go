package crypto_cloud_handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_api_dto"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_configuration"
	"io"
	"net/http"
)

type CreateInvoiceHandler struct {
	config     crypto_cloud_configuration.Configuration
	httpClient *http.Client
}

func NewCreateInvoiceHandler() CreateInvoiceHandler {
	return CreateInvoiceHandler{
		config:     crypto_cloud_configuration.NewConfiguration(),
		httpClient: &http.Client{},
	}
}

func (c *CreateInvoiceHandler) HandleCreateInvoice(command crypto_cloud_commands.IssueInvoiceCommand) (interface{}, error) {
	currencyCode, currencyCodeErr := command.Currency.String()
	if currencyCodeErr != nil {
		return nil, currencyCodeErr
	}

	request := crypto_cloud_api_dto.InvoiceRequest{
		Amount:   command.Amount,
		Currency: currencyCode,
		ShopID:   c.config.ShopId(),
		Email:    command.Email,
		OrderID:  fmt.Sprintf("N_%d", command.OrderId),
	}

	url := c.config.BaseUrl() + c.config.CreateInvoiceUrl()

	requestBody, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+c.config.ApiKey())
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errorResponse struct {
			Status string                 `json:"status"`
			Result map[string]interface{} `json:"result"`
		}
		if err := json.Unmarshal(body, &errorResponse); err != nil {
			return nil, errors.New("failed to parse error response")
		}
		return nil, errors.New("API error: " + formatErrorResult(errorResponse.Result))
	}

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("unexpected HTTP status: " + resp.Status)
	}

	var response crypto_cloud_api_dto.InvoiceResponse
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	if response.Status != "success" {
		return nil, errors.New("failed to create invoice: API error")
	}

	return response, nil
}

func formatErrorResult(result map[string]interface{}) string {
	errMessages := ""
	for key, value := range result {
		valueStr, ok := value.(string)
		if !ok {
			valueStr = "unknown error type"
		}
		errMessages += key + ": " + valueStr + "; "
	}
	return errMessages
}
