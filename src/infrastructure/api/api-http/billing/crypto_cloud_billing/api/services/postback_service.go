package services

import (
	"encoding/json"
	"goproxy/application/contracts"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/billing"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_api_dto"
	"log"
	"net/http"
	"strconv"
	"strings"
)

type PostbackService struct {
	orderRepository     contracts.OrderRepository
	planPriceRepository contracts.PlanPriceRepository
	paymentService      crypto_cloud.PaymentProvider
	messageBusService   billing.MessageBusProducer
}

func NewPostbackService(orderRepository contracts.OrderRepository,
	planPriceRepository contracts.PlanPriceRepository,
	paymentService crypto_cloud.PaymentProvider,
	messageBusService billing.MessageBusProducer) PostbackService {
	return PostbackService{
		orderRepository:     orderRepository,
		planPriceRepository: planPriceRepository,
		paymentService:      paymentService,
		messageBusService:   messageBusService,
	}
}

func (p *PostbackService) Handle(w http.ResponseWriter, dto crypto_cloud_api_dto.PostbackRequest) {
	cmd := crypto_cloud_commands.PostBackCommand{
		OrderID: dto.OrderID,
		Token:   dto.Token,
	}

	handlePostBackErr := p.paymentService.HandlePostBack(cmd)
	if handlePostBackErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	produceEventErr := p.producePlanAssignedEvent(dto)
	if produceEventErr != nil {
		log.Printf("could not produce plan assigned event: %v", produceEventErr)
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Postback received"})
	return
}

func (p *PostbackService) producePlanAssignedEvent(cmd crypto_cloud_api_dto.PostbackRequest) error {
	planId, email, cmdParseErr := p.getPlanIdAndEmail(cmd)
	if cmdParseErr != nil {
		return cmdParseErr
	}

	produceErr := p.messageBusService.ProducePlanAssignedEvent(planId, email)
	if produceErr != nil {
		return produceErr
	}

	return nil
}

func (p *PostbackService) getPlanIdAndEmail(cmd crypto_cloud_api_dto.PostbackRequest) (int, string, error) {
	orderIdPart := strings.Split(cmd.OrderID, "_")[1]
	orderId, err := strconv.Atoi(orderIdPart)
	if err != nil {
		return 0, "", err
	}

	order, orderErr := p.orderRepository.GetById(orderId)
	if orderErr != nil {
		return 0, "", orderErr
	}

	return order.PlanId(), order.Email(), nil
}
