package crypto_cloud_billing

import (
	"encoding/json"
	"goproxy/application/payments/crypto_cloud"
	"goproxy/application/payments/crypto_cloud/crypto_cloud_commands"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/crypto_cloud_billing_dto"
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
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var dto crypto_cloud_billing_dto.IssueInvoiceCommandDto
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&dto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	cmd := crypto_cloud_commands.IssueInvoiceCommand{
		AmountUSD: dto.AmountUSD,
		Email:     dto.Email,
		OrderId:   dto.OrderId,
	}

	result, resultErr := h.paymentService.IssueInvoice(cmd)
	if resultErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(resultErr.Error()))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(result)
}

func (h Handler) HandlePostBack(w http.ResponseWriter, r *http.Request) {
	panic("implement me")
}
