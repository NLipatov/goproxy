package crypto_cloud_billing

import (
	"encoding/json"
	"goproxy/application"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/crypto_cloud/crypto_cloud_dto"
	"log"
	"net/http"
)

type Handler struct {
	orderRepository     application.OrderRepository
	planPriceRepository application.PlanPriceRepository
	paymentService      crypto_cloud.PaymentProvider
}

func NewHandler(paymentService crypto_cloud.PaymentProvider, planPriceRepository application.PlanPriceRepository,
	orderRepository application.OrderRepository) Handler {
	return Handler{
		paymentService:      paymentService,
		planPriceRepository: planPriceRepository,
		orderRepository:     orderRepository,
	}
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
