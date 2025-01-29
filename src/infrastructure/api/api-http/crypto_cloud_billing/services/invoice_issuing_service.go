package services

import (
	"encoding/json"
	"fmt"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/domain/dataobjects"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/crypto_cloud/crypto_cloud_currencies"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/crypto_cloud_billing_dto"
	"net/http"
)

type BillingService struct {
	orderRepository     application.OrderRepository
	planPriceRepository application.PlanPriceRepository
	paymentService      crypto_cloud.PaymentProvider
}

func NewBillingService(orderRepository application.OrderRepository,
	planPriceRepository application.PlanPriceRepository,
	paymentService crypto_cloud.PaymentProvider) BillingService {
	return BillingService{
		orderRepository:     orderRepository,
		planPriceRepository: planPriceRepository,
		paymentService:      paymentService,
	}
}

func (h *BillingService) IssueInvoice(w http.ResponseWriter, dto crypto_cloud_billing_dto.IssueInvoiceCommandDto) {
	currency, amount, planPriceErr := h.getPlanPrice(dto.PlanId, dto.Currency)
	if planPriceErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("could not load plan prices"))
		return
	}

	emailValueObject, emailValidationErr := valueobjects.ParseEmailFromString(dto.Email)
	if emailValidationErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid email"))
		return
	}

	if amount == 0 {
		h.handleFreePlan(w, dto.PlanId, emailValueObject)
	} else {
		h.handlePaidPlan(w, dto.PlanId, emailValueObject, amount, currency)
	}
}

func (h *BillingService) getPlanPrice(planId int, preferredCurrencyCode string) (crypto_cloud_currencies.CryptoCloudCurrency, float64, error) {
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

func (h *BillingService) handlePaidPlan(w http.ResponseWriter, planId int, emailVO valueobjects.Email, amount float64, currency crypto_cloud_currencies.CryptoCloudCurrency) {
	orderId, orderErr := h.orderRepository.
		Create(dataobjects.
			NewOrder(-1, emailVO, planId, valueobjects.NewOrderStatus("NEW")))
	if orderErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("could not create order"))
		return
	}

	cmd := crypto_cloud_commands.IssueInvoiceCommand{
		Currency: currency,
		Amount:   amount,
		Email:    emailVO.String(),
		OrderId:  orderId,
	}

	result, resultErr := h.paymentService.IssueInvoice(cmd)
	if resultErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
	return
}

func (h *BillingService) handleFreePlan(w http.ResponseWriter, planId int, emailVO valueobjects.Email) {
	//check eligibility
	eligible := h.IsUserEligibleForFreePlan(planId, emailVO)
	if !eligible {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("not eligible for free plan"))
		return
	}

	// create new order
	_, newPlanOrder := h.orderRepository.
		Create(dataobjects.NewOrder(-1, emailVO, planId, valueobjects.NewOrderStatus("NEW")))
	if newPlanOrder != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("could not create order"))
		return
	}

	w.WriteHeader(http.StatusCreated)
	return
}

func (h *BillingService) IsUserEligibleForFreePlan(planId int, emailVO valueobjects.Email) bool {
	freeOrders, _ := h.orderRepository.GetByPlanIdAndEmail(planId, emailVO)
	if freeOrders == nil {
		return true
	}

	return false
}
