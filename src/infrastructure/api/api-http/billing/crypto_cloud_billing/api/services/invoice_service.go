package services

import (
	"fmt"
	"goproxy/application/contracts"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/domain/dataobjects"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/api/api-http/billing"
	commands2 "goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/commands"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_api_dto"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_currencies"
)

type BillingService struct {
	orderRepository     contracts.OrderRepository
	planPriceRepository contracts.PlanPriceRepository
	paymentService      crypto_cloud.PaymentProvider
	messageBusService   billing.MessageBusProducer
}

func NewBillingService(orderRepository contracts.OrderRepository,
	planPriceRepository contracts.PlanPriceRepository,
	paymentService crypto_cloud.PaymentProvider,
	messageBusService billing.MessageBusProducer) BillingService {
	return BillingService{
		orderRepository:     orderRepository,
		planPriceRepository: planPriceRepository,
		paymentService:      paymentService,
		messageBusService:   messageBusService,
	}
}

func (h *BillingService) CreateInvoice(cmd commands2.CreateInvoiceCommand) (crypto_cloud_api_dto.InvoiceResponse, error) {
	currency, amount, priceErr := h.getPlanPrice(cmd.PlanId(), cmd.Currency())
	if priceErr != nil {
		return crypto_cloud_api_dto.InvoiceResponse{}, priceErr
	}

	orderId, orderErr := h.createOrder(cmd.Email(), cmd.PlanId())
	if orderErr != nil {
		return crypto_cloud_api_dto.InvoiceResponse{}, orderErr
	}

	result, resultErr := h.paymentService.IssueInvoice(crypto_cloud_commands.IssueInvoiceCommand{
		Currency: currency,
		Amount:   amount,
		Email:    cmd.Email().String(),
		OrderId:  orderId,
	})
	if resultErr != nil {
		return crypto_cloud_api_dto.InvoiceResponse{}, resultErr
	}

	return result.(crypto_cloud_api_dto.InvoiceResponse), nil
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

func (h *BillingService) createOrder(email valueobjects.Email, planId int) (int, error) {
	orderId, orderErr := h.orderRepository.
		Create(dataobjects.
			NewOrder(-1, email, planId, valueobjects.NewOrderStatus("NEW")))
	if orderErr != nil {
		return 0, orderErr
	}

	return orderId, nil
}
