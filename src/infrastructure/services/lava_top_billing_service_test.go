package services

import (
	"encoding/json"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/dto"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestPublishInvoice(t *testing.T) {
	mockResponse := dto.InvoicePaymentParamsResponse{
		ID:     "e624e74b-a109-4775-b8e2-be27ce89a0b8",
		Status: "in-progress",
		AmountTotal: dto.AmountTotalDto{
			Currency: "RUB",
			Amount:   50,
		},
		PaymentURL: "https://app.lava.top/products/c48a74d5-92f7-4671-b560-3d2635fc3f80/6c0cf730-3432-4755-941b-ca23b419d6df",
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("X-Api-Key") == "" {
			t.Fatal("Missing X-Api-Key header")
		}

		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	email, _ := valueobjects.ParseEmailFromString("example@example.com")
	offerId, _ := valueobjects.ParseGuidFromString("6c0cf730-3432-4755-941b-ca23b419d6df")
	status, _ := lavatopvalueobjects.ParseInvoiceStatus("new")
	invoice, _ := lavatopaggregates.NewInvoice(
		1,
		"",
		status,
		email,
		offerId,
		lavatopvalueobjects.ONE_TIME,
		lavatopvalueobjects.RUB,
		lavatopvalueobjects.BANK131,
		lavatopvalueobjects.EN,
	)

	service := LavaTopBillingService{
		postInvoiceUrl: mockServer.URL,
		apiKey:         "test-api-key",
	}

	updatedInvoice, err := service.PublishInvoice(invoice)
	if err != nil {
		t.Fatalf("PublishInvoice returned error: %v", err)
	}

	if updatedInvoice.Id() != invoice.Id() {
		t.Errorf("Expected ID %d, got %d", invoice.Id(), updatedInvoice.Id())
	}

	if updatedInvoice.OfferId() != invoice.OfferId() {
		t.Errorf("Expected OfferId %s, got %s", invoice.OfferId().String(), updatedInvoice.OfferId().String())
	}

	if updatedInvoice.Status().String() != mockResponse.Status {
		t.Errorf("Expected Status %s, got %s", mockResponse.Status, updatedInvoice.Status().String())
	}

	if updatedInvoice.Currency() != invoice.Currency() {
		t.Errorf("Expected Currency %s, got %s", invoice.Currency().String(), updatedInvoice.Currency().String())
	}

	if updatedInvoice.ExtId() == "" {
		t.Error("Expected ExtId to have a value, got empty string")
	}
}

func TestPublishInvoice_404Response(t *testing.T) {
	mockErrorResponse := dto.ErrorResponse{
		Error: "Input fields are invalid",
		Details: map[string]string{
			"email":    "should be well-formed email",
			"password": "should contain at least 8 symbols",
		},
		Timestamp: "2025-01-10T20:08:42.378Z",
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("Expected POST request, got %s", r.Method)
		}

		if r.Header.Get("X-Api-Key") == "" {
			t.Fatal("Missing X-Api-Key header")
		}

		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(mockErrorResponse)
	}))
	defer mockServer.Close()

	email, _ := valueobjects.ParseEmailFromString("invalid_email")
	offerId, _ := valueobjects.ParseGuidFromString("6c0cf730-3432-4755-941b-ca23b419d6df")
	status, _ := lavatopvalueobjects.ParseInvoiceStatus("new")
	invoice, _ := lavatopaggregates.NewInvoice(
		1,
		"",
		status,
		email,
		offerId,
		lavatopvalueobjects.ONE_TIME,
		lavatopvalueobjects.RUB,
		lavatopvalueobjects.BANK131,
		lavatopvalueobjects.EN,
	)

	service := LavaTopBillingService{
		postInvoiceUrl: mockServer.URL,
		apiKey:         "test-api-key",
	}

	_, err := service.PublishInvoice(invoice)
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	expectedErrorMessage := "unexpected status code: 404"
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Errorf("Unexpected error message: got %q, expected %q", err.Error(), expectedErrorMessage)
	}
}

func TestGetInvoiceStatus(t *testing.T) {
	mockResponse := dto.InvoicePaymentParamsResponse{
		ID:     "f1f47a26-4795-420e-8dc1-2260dd065fbb",
		Status: "in-progress",
		AmountTotal: dto.AmountTotalDto{
			Currency: "RUB",
			Amount:   50,
		},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("Expected GET request, got %s", r.Method)
		}

		if r.Header.Get("X-Api-Key") == "" {
			t.Fatal("Missing X-Api-Key header")
		}

		query := r.URL.Query()
		if query.Get("id") != mockResponse.ID {
			t.Fatalf("Expected ID %s, got %s", mockResponse.ID, query.Get("id"))
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	service := LavaTopBillingService{
		getInvoiceUrl: mockServer.URL,
		apiKey:        "test-api-key",
	}

	status, err := service.GetInvoiceStatus("f1f47a26-4795-420e-8dc1-2260dd065fbb")
	if err != nil {
		t.Fatalf("GetInvoiceStatus returned error: %v", err)
	}

	expectedStatus := "in-progress"
	if status != expectedStatus {
		t.Errorf("Expected status %s, got %s", expectedStatus, status)
	}
}

func TestGetInvoiceStatus_404Response(t *testing.T) {
	mockErrorResponse := map[string]interface{}{
		"error":     "{\"errors\":[\"Contract with id 'f1f47a26-4795-420e-8dc1-2260dd065f8b' does not exists\"]}",
		"details":   map[string]interface{}{},
		"timestamp": "2025-01-11T16:29:57.326842+02:00",
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("Expected GET request, got %s", r.Method)
		}

		if r.Header.Get("X-Api-Key") == "" {
			t.Fatal("Missing X-Api-Key header")
		}

		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(mockErrorResponse)
	}))
	defer mockServer.Close()

	service := LavaTopBillingService{
		getInvoiceUrl: mockServer.URL,
		apiKey:        "test-api-key",
	}

	_, err := service.GetInvoiceStatus("f1f47a26-4795-420e-8dc1-2260dd065f8b")
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}

	expectedErrorMessage := "unexpected status code: 404"
	if !strings.Contains(err.Error(), expectedErrorMessage) {
		t.Errorf("Unexpected error message: got %q, expected %q", err.Error(), expectedErrorMessage)
	}

	expectedErrorDetails := "Contract with id 'f1f47a26-4795-420e-8dc1-2260dd065f8b' does not exists"
	if !strings.Contains(err.Error(), expectedErrorDetails) {
		t.Errorf("Error message should include details: got %q, expected to contain %q", err.Error(), expectedErrorDetails)
	}
}
