package handlers

import (
	"encoding/json"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/crypto_cloud_billing_dto"
	"goproxy/infrastructure/api/api-http/crypto_cloud_billing/services"
	"net/http"
)

type CreateInvoiceHandler struct {
	billingService services.BillingService
}

func NewCreateInvoiceHandler(billingService services.BillingService) CreateInvoiceHandler {
	return CreateInvoiceHandler{
		billingService: billingService,
	}
}

func (ih *CreateInvoiceHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var dto crypto_cloud_billing_dto.IssueInvoiceCommandDto
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&dto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	ih.billingService.IssueInvoice(w, dto)
}
