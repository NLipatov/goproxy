package crypto_cloud_billing

import (
	"encoding/json"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/crypto_cloud_billing_dto"
	"goproxy/infrastructure/payments/crypto_cloud/crypto_cloud_dto"
	"log"
	"net/http"
)

type Handler struct {
	paymentService crypto_cloud.PaymentProvider
}

func NewHandler(paymentService crypto_cloud.PaymentProvider) Handler {
	return Handler{
		paymentService: paymentService,
	}
}

func (h Handler) IssueInvoice(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		log.Printf("PostBack handling: 405 as method not allowed")
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var dto crypto_cloud_billing_dto.IssueInvoiceCommandDto
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&dto); err != nil {
		log.Printf("PostBack handling: 400 as body object scheme is invalid: %s", err.Error())
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	cmd := crypto_cloud_commands.IssueInvoiceCommand{
		AmountUSD: dto.AmountUSD,
		Email:     dto.Email,
		OrderId:   dto.OrderId,
	}

	result, resultErr := h.paymentService.IssueInvoice(cmd)
	if resultErr != nil {
		log.Printf("PostBack handling: 400 as body object is invalid: %s", resultErr.Error())
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

func (h Handler) HandlePostBack(w http.ResponseWriter, r *http.Request) {
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
