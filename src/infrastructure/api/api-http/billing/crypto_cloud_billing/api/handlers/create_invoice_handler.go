package handlers

import (
	"encoding/json"
	"fmt"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/commands"
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
		jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[dto.Response]](w)
		_ = jsonResponse.Respond(dto2.ApiResponse[dto.Response]{
			Payload:      nil,
			ErrorCode:    401,
			ErrorMessage: "invalid token",
		})

		return
	}

	emailVO, emailVOErr := valueobjects.ParseEmailFromString(email)
	if emailVOErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[dto.Response]](w)
		_ = jsonResponse.Respond(dto2.ApiResponse[dto.Response]{
			Payload:      nil,
			ErrorCode:    400,
			ErrorMessage: fmt.Sprintf("email '%s' is considered invalid", email),
		})

		return
	}

	var request dto.Request
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&request); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid request body"))
		return
	}

	cmd := commands.NewCreateInvoiceCommand(emailVO, request.PlanId, request.Currency)

	issueInvoiceResult, issueInvoiceResultErr := ih.billingService.CreateInvoice(cmd)
	if issueInvoiceResultErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[dto.Response]](w)
		_ = jsonResponse.Respond(dto2.ApiResponse[dto.Response]{
			Payload:      nil,
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "cannot process request at the moment",
		})

		return
	}

	w.WriteHeader(http.StatusOK)
	jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[dto.Response]](w)
	_ = jsonResponse.Respond(dto2.ApiResponse[dto.Response]{
		Payload: &dto.Response{
			PaymentLink: issueInvoiceResult.Result.Link,
		},
		ErrorCode:    0,
		ErrorMessage: "",
	})
}
