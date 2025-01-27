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
		Id:     "e624e74b-a109-4775-b8e2-be27ce89a0b8",
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

	userId := 2
	email, _ := valueobjects.ParseEmailFromString("example@example.com")
	offer := lavatopvalueobjects.NewOffer("6c0cf730-3432-4755-941b-ca23b419d6df", "1 Month Plan", make([]lavatopvalueobjects.Price, 0))
	status, _ := lavatopvalueobjects.ParseInvoiceStatus("new")
	invoice, _ := lavatopaggregates.NewInvoice(
		1,
		userId,
		"",
		status,
		email,
		offer,
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

	if updatedInvoice.UserId() != userId {
		t.Errorf("Expected userId %d, got %d", userId, invoice.UserId())
	}

	if updatedInvoice.Id() != invoice.Id() {
		t.Errorf("Expected Id %d, got %d", invoice.Id(), updatedInvoice.Id())
	}

	if updatedInvoice.Offer().ExtId() != invoice.Offer().ExtId() {
		t.Errorf("Expected Offer %s, got %s", invoice.Offer().ExtId(), updatedInvoice.Offer().ExtId())
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

	userId := 2
	email, _ := valueobjects.ParseEmailFromString("invalid_email")
	offer := lavatopvalueobjects.NewOffer("6c0cf730-3432-4755-941b-ca23b419d6df", "1 Month Plan", make([]lavatopvalueobjects.Price, 0))
	status, _ := lavatopvalueobjects.ParseInvoiceStatus("new")
	invoice, _ := lavatopaggregates.NewInvoice(
		1,
		userId,
		"",
		status,
		email,
		offer,
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
		Id:     "f1f47a26-4795-420e-8dc1-2260dd065fbb",
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
		if query.Get("id") != mockResponse.Id {
			t.Fatalf("Expected Id %s, got %s", mockResponse.Id, query.Get("id"))
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

func TestGetOffers(t *testing.T) {
	mockResponse := dto.GetOffersResponse{
		Items: []dto.ProductResponse{
			{
				ID:    "c48a74d5-92f7-4671-b560-3d2635fc3f80",
				Title: "1 Month Plan",
				Offers: []dto.OfferResponse{
					{
						Id:   "6c0cf730-3432-4755-941b-ca23b419d6df",
						Name: "1 Month Plan",
						Prices: []dto.PriceDto{
							{
								Currency:    "EUR",
								Amount:      0.48,
								Periodicity: "ONE_TIME",
							},
							{
								Currency:    "RUB",
								Amount:      50,
								Periodicity: "ONE_TIME",
							},
							{
								Currency:    "USD",
								Amount:      0.49,
								Periodicity: "ONE_TIME",
							},
						},
					},
				},
			},
		},
	}

	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("Expected GET request, got %s", r.Method)
		}

		if r.Header.Get("X-Api-Key") == "" {
			t.Fatal("Missing X-Api-Key header")
		}

		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(mockResponse)
	}))
	defer mockServer.Close()

	service := LavaTopBillingService{
		getOffersUrl: mockServer.URL,
		apiKey:       "test-api-key",
	}

	eurPrice := lavatopvalueobjects.NewPrice(48, lavatopvalueobjects.EUR, lavatopvalueobjects.ONE_TIME)
	rubPrice := lavatopvalueobjects.NewPrice(5000, lavatopvalueobjects.RUB, lavatopvalueobjects.ONE_TIME)
	usdPrice := lavatopvalueobjects.NewPrice(49, lavatopvalueobjects.USD, lavatopvalueobjects.ONE_TIME)
	prices := []lavatopvalueobjects.Price{eurPrice, rubPrice, usdPrice}

	expectedOffer := lavatopvalueobjects.NewOffer(
		"6c0cf730-3432-4755-941b-ca23b419d6df",
		"1 Month Plan",
		prices,
	)

	offers, err := service.GetOffers()
	if err != nil {
		t.Fatalf("GetOffers returned error: %v", err)
	}

	expectedOfferCount := 1
	if len(offers) != expectedOfferCount {
		t.Errorf("Expected %d offers, got %d", expectedOfferCount, len(offers))
	}

	actualOffer := offers[0]

	if actualOffer.ExtId() != expectedOffer.ExtId() {
		t.Errorf("Expected offer Id %s, got %s", expectedOffer.ExtId(), actualOffer.ExtId())
	}

	if actualOffer.Name() != expectedOffer.Name() {
		t.Errorf("Expected offer name %s, got %s", expectedOffer.Name(), actualOffer.Name())
	}

	if actualOffer.Prices()[0].Cents() != expectedOffer.Prices()[0].Cents() {
		t.Errorf("Expected price %d, got %d", expectedOffer.Prices()[0].Cents(), actualOffer.Prices()[0].Cents())
	}

	if actualOffer.Prices()[0].Currency() != expectedOffer.Prices()[0].Currency() {
		t.Errorf("Expected currency %s, got %s", expectedOffer.Prices()[0].Currency(), actualOffer.Prices()[0].Currency())
	}

	if actualOffer.Prices()[0].Periodicity() != expectedOffer.Prices()[0].Periodicity() {
		t.Errorf("Expected periodicity %s, got %s", expectedOffer.Prices()[0].Periodicity(), actualOffer.Prices()[0].Periodicity())
	}

	expectedPrices := make(map[string]lavatopvalueobjects.Price)
	for _, price := range expectedOffer.Prices() {
		expectedPrices[price.Currency().String()] = price
	}

	for _, price := range offers[0].Prices() {
		expectedPrice, exists := expectedPrices[price.Currency().String()]
		if !exists {
			t.Errorf("Expected price for currency %s, but it was not found", price.Currency().String())
			continue
		}

		if expectedPrice.Cents() != price.Cents() {
			t.Errorf("Expected %d cents for currency %s, got %d", price.Cents(), price.Currency().String(), expectedPrice.Cents())
		}

		if expectedPrice.Periodicity() != price.Periodicity() {
			t.Errorf("Expected periodicity %s for currency %s, got %s", price.Periodicity(), price.Currency(), expectedPrice.Periodicity())
		}
	}
}
