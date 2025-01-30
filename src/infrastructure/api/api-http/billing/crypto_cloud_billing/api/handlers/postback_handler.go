package handlers

import (
	"encoding/json"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/services"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/crypto_cloud_api/crypto_cloud_api_dto"
	"net/http"
)

type PostbackHandler struct {
	postbackService services.PostbackService
}

func NewPostbackHandler(postbackService services.PostbackService) PostbackHandler {
	return PostbackHandler{
		postbackService: postbackService,
	}
}

func (h *PostbackHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var dto crypto_cloud_api_dto.PostbackRequest
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8_000))
	if err := decoder.Decode(&dto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	h.postbackService.Handle(w, dto)
}
