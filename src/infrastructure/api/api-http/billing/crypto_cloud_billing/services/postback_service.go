package services

import (
	"encoding/json"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_api_dto"
	"net/http"
)

type PostbackService struct {
	orderRepository     application.OrderRepository
	planPriceRepository application.PlanPriceRepository
	paymentService      crypto_cloud.PaymentProvider
}

func NewPostbackService(orderRepository application.OrderRepository,
	planPriceRepository application.PlanPriceRepository,
	paymentService crypto_cloud.PaymentProvider) PostbackService {
	return PostbackService{
		orderRepository:     orderRepository,
		planPriceRepository: planPriceRepository,
		paymentService:      paymentService,
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

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Postback received"})
	return

}
