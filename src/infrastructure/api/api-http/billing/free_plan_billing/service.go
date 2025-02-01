package free_plan_billing

import (
	"fmt"
	"goproxy/application/contracts"
	"goproxy/domain/dataobjects"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/api/api-http/billing"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/dto"
)

type Service struct {
	orderRepository   contracts.OrderRepository
	messageBusService billing.MessageBusProducer
}

func NewService(orderRepository contracts.OrderRepository, messageBusProducer contracts.MessageBusService) *Service {
	return &Service{
		orderRepository:   orderRepository,
		messageBusService: billing.NewMessageBusProducer(messageBusProducer),
	}
}

func (s *Service) handle(dto dto.IssueInvoiceCommandDto) error {
	planId := dto.PlanId
	emailVO, emailVOErr := valueobjects.ParseEmailFromString(dto.Email)
	if emailVOErr != nil {
		return emailVOErr
	}

	//check eligibility
	eligible := s.IsUserEligibleForFreePlan(planId, emailVO)
	if !eligible {
		return fmt.Errorf("not eligible for free plan")
	}

	// create new order
	_, newPlanOrder := s.orderRepository.
		Create(dataobjects.NewOrder(-1, emailVO, planId, valueobjects.NewOrderStatus("NEW")))
	if newPlanOrder != nil {
		return fmt.Errorf("could not create order: %s", newPlanOrder)
	}

	produceEventErr := s.messageBusService.ProducePlanAssignedEvent(planId, emailVO.String())
	if produceEventErr != nil {
		return fmt.Errorf("could not produce event: %s", produceEventErr)
	}

	return nil
}

func (s *Service) IsUserEligibleForFreePlan(planId int, emailVO valueobjects.Email) bool {
	freeOrders, _ := s.orderRepository.GetByPlanIdAndEmail(planId, emailVO)
	if freeOrders == nil {
		return true
	}

	return false
}
