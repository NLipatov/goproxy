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
	postInvoiceUrl string
	apiKey         string
}

func NewLavaTopBillingService() LavaTopBillingService {
	postInvoiceUrl := os.Getenv("POST_INVOICE_API_URL")
	if postInvoiceUrl == "" {
		log.Fatalf("POST_INVOICE_API_URL environment variable not set")
	}

	apiKey := os.Getenv("LAVATOP_API_KEY")
	if apiKey == "" {
		log.Fatalf("LAVATOP_API_KEY environment variable not set")
	}

	return LavaTopBillingService{
		postInvoiceUrl: postInvoiceUrl,
		apiKey:         apiKey,
	}
}

func (l *LavaTopBillingService) PublishInvoice(invoice lavatopaggregates.Invoice) (lavatopaggregates.Invoice, error) {
	dtoInvoice, err := dto.ToInvoiceDTO(invoice)
	if err != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to convert invoice to DTO: %w", err)
	}

	data, err := json.Marshal(dtoInvoice)
	if err != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to marshal invoice DTO: %w", err)
	}

	req, err := http.NewRequest("POST", l.postInvoiceUrl, bytes.NewBuffer(data))
	if err != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("accept", "application/json")
	req.Header.Set("X-Api-Key", l.apiKey)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusCreated {
		return lavatopaggregates.Invoice{}, fmt.Errorf("unexpected status code: %d, body: %s", resp.StatusCode, string(body))
	}

	var successResponse dto.InvoicePaymentParamsResponse
	deserializationError := json.Unmarshal(body, &successResponse)
	if deserializationError != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to unmarshal success response: %w", err)
	}

	extId := successResponse.ID
	updatedStatus, updatedStatusErr := lavatopvalueobjects.ParseInvoiceStatus(successResponse.Status)
	if updatedStatusErr != nil {
		return lavatopaggregates.Invoice{}, fmt.Errorf("failed to parse invoice status: %w", updatedStatusErr)
	}

	updatedInvoice, err := lavatopaggregates.NewInvoice(
		invoice.Id(),
		extId,
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
