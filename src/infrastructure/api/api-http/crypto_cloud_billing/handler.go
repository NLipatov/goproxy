package crypto_cloud_billing

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/crypto_cloud_billing_dto"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_currencies"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_dto"
	"log"
	"net/http"
)

type Handler struct {
	planPriceRepository application.PlanPriceRepository
	paymentService      crypto_cloud.PaymentProvider
}

func NewHandler(paymentService crypto_cloud.PaymentProvider, planPriceRepository application.PlanPriceRepository) Handler {
	return Handler{
		paymentService:      paymentService,
		planPriceRepository: planPriceRepository,
	}
}

func (h *Handler) IssueInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("PostBack handling: 405 as method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var dto crypto_cloud_billing_dto.IssueInvoiceCommandDto
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&dto); err != nil {
		log.Printf("PostBack handling: 400 as body object scheme is invalid: %s", err)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	currency, amount, planPriceErr := h.getPlanPrice(dto.PlanId, dto.Currency)
	if planPriceErr != nil {
		log.Printf("PostBack handling: 500 could not get plan prices: %s", planPriceErr)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("could not load plan prices"))
		return
	}

	if amount == 0 {
		w.WriteHeader(http.StatusOK)
		return
	}

	cmd := crypto_cloud_commands.IssueInvoiceCommand{
		Currency: currency,
		Amount:   amount,
		Email:    dto.Email,
		OrderId:  dto.OrderId,
	}

	result, resultErr := h.paymentService.IssueInvoice(cmd)
	if resultErr != nil {
		log.Printf("PostBack handling: 400 as body object is invalid: %s", resultErr)
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	log.Printf("PostBack handling: 200 OK")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
	return
}

func (h *Handler) HandlePostBack(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("PostBack handling: 405 as method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var dto crypto_cloud_dto.PostbackRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8_000))
	if err := decoder.Decode(&dto); err != nil {
		log.Printf("PostBack handling: 400 as body object scheme is invalid: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	cmd := crypto_cloud_commands.PostBackCommand{
		OrderID: dto.OrderID,
		Token:   dto.Token,
	}

	handlePostBackErr := h.paymentService.HandlePostBack(cmd)
	if handlePostBackErr != nil {
		log.Printf("PostBack handling: 400 as body object data is invalid: %s", handlePostBackErr.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	log.Printf("PostBack handling: 200 OK")
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write([]byte("{'message': 'Postback received'}"))
	return
}

func (h *Handler) getPlanPrice(planId int, preferredCurrencyCode string) (crypto_cloud_currencies.CryptoCloudCurrency, float64, error) {
	currency := crypto_cloud_currencies.NewCryptoCloudCurrency(preferredCurrencyCode)

	prices, pricesErr := h.planPriceRepository.GetAllWithPlanId(planId)
	if pricesErr != nil {
		return 0, 0, fmt.Errorf("failed to load prices for plan %d: %s", planId, pricesErr)
	}

	if len(prices) == 0 {
		return 0, 0, fmt.Errorf("failed to load prices for plan %d: no prices found", planId)
	}

	for _, p := range prices {
		if crypto_cloud_currencies.NewCryptoCloudCurrency(p.Currency()) == currency {
			return currency, float64(p.Cents()) / 100.0, nil
		}
	}

	return crypto_cloud_currencies.NewCryptoCloudCurrency(prices[0].Currency()), float64(prices[0].Cents()) / 100.0, pricesErr
}
