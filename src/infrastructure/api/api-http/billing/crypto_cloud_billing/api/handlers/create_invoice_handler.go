package handlers

import (
	"encoding/json"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/dto"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/services"
	"goproxy/infrastructure/api/api-http/google_auth"
	"goproxy/infrastructure/api/api-http/http_objects"
	dto2 "goproxy/infrastructure/dto"
	"net/http"
)

type CreateInvoiceHandler struct {
	billingService        services.BillingService
	authenticationService google_auth.GoogleAuthService
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

	userRequest := http_objects.NewAuthenticatedUserBillingRequest(r)
	email, emailErr := userRequest.UserEmail()
	if emailErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[dto.CreateInvoiceResponse]](w)
		respondErr := jsonResponse.Respond(dto2.ApiResponse[dto.CreateInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    401,
			ErrorMessage: "invalid token",
		})

		if respondErr != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}

		return
	}

	var requestDto dto.IssueInvoiceCommandDto
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&requestDto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	requestDto.Email = email

	issueInvoiceResult, issueInvoiceResultErr := ih.billingService.IssueInvoice(requestDto)
	if issueInvoiceResultErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[dto.CreateInvoiceResponse]](w)
		_ = jsonResponse.Respond(dto2.ApiResponse[dto.CreateInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "cannot process request at the moment",
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[dto.CreateInvoiceResponse]](w)
	_ = jsonResponse.Respond(dto2.ApiResponse[dto.CreateInvoiceResponse]{
		Payload: &dto.CreateInvoiceResponse{
			PaymentLink: issueInvoiceResult.PaymentLinq,
		},
		ErrorCode:    0,
		ErrorMessage: "",
	})
}
