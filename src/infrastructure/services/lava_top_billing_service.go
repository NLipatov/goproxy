package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/infrastructure/dto"
	"io"
	"log"
	"net/http"
	"os"
)

type LavaTopBillingService struct {
	getInvoiceUrl  string
	postInvoiceUrl string
	apiKey         string
}

func NewLavaTopBillingService() LavaTopBillingService {
	getInvoiceUrl := os.Getenv("GET_INVOICE_API_URL")
	if getInvoiceUrl == "" {
		log.Fatalf("GET_INVOICE_API_URL environment variable not set")
	}

	postInvoiceUrl := os.Getenv("POST_INVOICE_API_URL")
	if postInvoiceUrl == "" {
		log.Fatalf("POST_INVOICE_API_URL environment variable not set")
	}

	apiKey := os.Getenv("LAVATOP_API_KEY")
	if apiKey == "" {
		log.Fatalf("LAVATOP_API_KEY environment variable not set")
	}

	return LavaTopBillingService{
		getInvoiceUrl:  getInvoiceUrl,
		postInvoiceUrl: postInvoiceUrl,
		apiKey:         apiKey,
	}
}

func (l *LavaTopBillingService) GetInvoiceStatus(offerId string) (string, error) {
	url := fmt.Sprintf("%s?id=%s", l.getInvoiceUrl, offerId)
	var successResponse dto.InvoicePaymentParamsResponse

	err := l.doRequest(http.MethodGet, url, nil, &successResponse)
	if err != nil {
		return "", err
	}

	status, err := lavatopvalueobjects.ParseInvoiceStatus(successResponse.Status)
	if err != nil {
		return "", fmt.Errorf("failed to parse invoice status: %w", err)
	}

	return status.String(), nil
}

func (l *LavaTopBillingService) PublishInvoice(invoice lavatopaggregates.Invoice) (lavatopaggregates.Invoice, error) {
	dtoInvoice, err := dto.ToInvoiceDTO(invoice)
	if err != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to convert invoice to DTO: %w", err)
	}

	var successResponse dto.InvoicePaymentParamsResponse
	err = l.doRequest(http.MethodPost, l.postInvoiceUrl, dtoInvoice, &successResponse)
	if err != nil {
		return lavatopaggregates.Invoice{}, err
	}

	updatedStatus, err := lavatopvalueobjects.ParseInvoiceStatus(successResponse.Status)
	if err != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to parse invoice status: %w", err)
	}

	updatedInvoice, err := lavatopaggregates.NewInvoice(
		invoice.Id(),
		successResponse.ID,
		updatedStatus,
		invoice.Email(),
		invoice.OfferId(),
		invoice.Periodicity(),
		invoice.Currency(),
		invoice.PaymentMethod(),
		invoice.BuyerLanguage(),
	)
	if err != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to create updated invoice: %w", err)
	}

	return updatedInvoice, nil
}

func (l *LavaTopBillingService) doRequest(method, url string, payload interface{}, response interface{}) error {
	var bodyReader io.Reader
	if payload != nil {
		data, err := json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("failed to marshal payload: %w", err)
		}
		bodyReader = bytes.NewBuffer(data)
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("X-Api-Key", l.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	if response != nil {
		if err := json.NewDecoder(resp.Body).Decode(response); err != nil {
			return fmt.Errorf("failed to unmarshal response: %w", err)
		}
	}

	return nil
}
