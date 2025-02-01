package services

import (
	"fmt"
	"goproxy/application/contracts"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/domain/dataobjects"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/api/api-http/billing"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/dto"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_api_dto"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_currencies"
)

type IssueInvoiceResult struct {
	IsPaymentRequired bool
	Invoice           crypto_cloud_api_dto.InvoiceResponse
	PaymentLinq       string
}

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

func (h *BillingService) IssueInvoice(dto dto.IssueInvoiceCommandDto) (IssueInvoiceResult, error) {
	currency, amount, planPriceErr := h.getPlanPrice(dto.PlanId, dto.Currency)
	if planPriceErr != nil {
		return IssueInvoiceResult{}, planPriceErr
	}

	emailValueObject, emailValidationErr := valueobjects.ParseEmailFromString(dto.Email)
	if emailValidationErr != nil {
		return IssueInvoiceResult{}, emailValidationErr
	}

	if amount == 0 {
		handlingErr := h.handleFreePlan(dto.PlanId, emailValueObject)
		if handlingErr != nil {
			return IssueInvoiceResult{}, handlingErr
		}

		return IssueInvoiceResult{
			IsPaymentRequired: false,
			Invoice:           crypto_cloud_api_dto.InvoiceResponse{},
			PaymentLinq:       "",
		}, handlingErr
	} else {
		response, responseErr := h.handlePaidPlan(dto.PlanId, emailValueObject, amount, currency)
		if responseErr != nil {
			return IssueInvoiceResult{
				IsPaymentRequired: true,
				Invoice:           crypto_cloud_api_dto.InvoiceResponse{},
				PaymentLinq:       "",
			}, responseErr
		}
		return IssueInvoiceResult{
			IsPaymentRequired: true,
			Invoice:           response,
			PaymentLinq:       response.Result.Link,
		}, nil
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

func (h *BillingService) handlePaidPlan(planId int, emailVO valueobjects.Email, amount float64, currency crypto_cloud_currencies.CryptoCloudCurrency) (crypto_cloud_api_dto.InvoiceResponse, error) {
	orderId, orderErr := h.orderRepository.
		Create(dataobjects.
			NewOrder(-1, emailVO, planId, valueobjects.NewOrderStatus("NEW")))
	if orderErr != nil {
		return crypto_cloud_api_dto.InvoiceResponse{}, orderErr
	}

	cmd := crypto_cloud_commands.IssueInvoiceCommand{
		Currency: currency,
		Amount:   amount,
		Email:    emailVO.String(),
		OrderId:  orderId,
	}

	result, resultErr := h.paymentService.IssueInvoice(cmd)
	if resultErr != nil {
		return crypto_cloud_api_dto.InvoiceResponse{}, resultErr
	}

	return result.(crypto_cloud_api_dto.InvoiceResponse), nil
}

func (h *BillingService) handleFreePlan(planId int, emailVO valueobjects.Email) error {
	//check eligibility
	eligible := h.IsUserEligibleForFreePlan(planId, emailVO)
	if !eligible {
		return fmt.Errorf("not eligible for free plan")
	}

	// create new order
	_, newPlanOrder := h.orderRepository.
		Create(dataobjects.NewOrder(-1, emailVO, planId, valueobjects.NewOrderStatus("NEW")))
	if newPlanOrder != nil {
		return fmt.Errorf("could not create order: %s", newPlanOrder)
	}

	produceEventErr := h.messageBusService.ProducePlanAssignedEvent(planId, emailVO.String())
	if produceEventErr != nil {
		return fmt.Errorf("could not produce event: %s", produceEventErr)
	}

	return nil
}

func (h *BillingService) IsUserEligibleForFreePlan(planId int, emailVO valueobjects.Email) bool {
	freeOrders, _ := h.orderRepository.GetByPlanIdAndEmail(planId, emailVO)
	if freeOrders == nil {
		return true
	}

	return false
}
