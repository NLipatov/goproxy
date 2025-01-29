package crypto_cloud_handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	crypto_cloud_dto2 "goproxy/infrastructure/api/api-http/crypto_cloud_billing/crypto_cloud/crypto_cloud_dto"
	"io"
	"net/http"
)

func HandleCreateInvoice(httpClient *http.Client, url string, apiKey string, requestData crypto_cloud_dto2.InvoiceRequest) (interface{}, error) {
	requestBody, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Token "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
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

	var response crypto_cloud_dto2.InvoiceResponse
	if unmarshalErr := json.Unmarshal(body, &response); unmarshalErr != nil {
		return nil, unmarshalErr
	}

	if response.Status != "success" {
		return nil, errors.New("failed to create invoice: API error")
	}

	return response.Result, nil
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
