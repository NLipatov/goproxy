package free_plan_billing

import (
	"encoding/json"
	"goproxy/infrastructure/api/api-http/billing/crypto_cloud_billing/api/dto"
	"goproxy/infrastructure/api/api-http/http_objects"
	dto2 "goproxy/infrastructure/dto"
	"net/http"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) Handler {
	return Handler{
		service: service,
	}
}

func (c *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	userRequest := http_objects.NewAuthenticatedUserBillingRequest(r)
	email, emailErr := userRequest.UserEmail()
	if emailErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[dto.CreateInvoiceResponse]](w)
		_ = jsonResponse.Respond(dto2.ApiResponse[dto.CreateInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    401,
			ErrorMessage: "invalid token",
		})
		return
	}

	var requestDto dto.IssueInvoiceCommandDto
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&requestDto); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[InvoiceResponse]](w)
		_ = jsonResponse.Respond(dto2.ApiResponse[InvoiceResponse]{
			Payload: &InvoiceResponse{
				PlanAssigned: false,
			},
			ErrorCode:    400,
			ErrorMessage: "invalid body",
		})
		return
	}

	requestDto.Email = email

	handleErr := c.service.handle(requestDto)
	if handleErr != nil {
		if handleErr.Error() == "not eligible for free plan" {
			w.WriteHeader(http.StatusBadRequest)
			jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[InvoiceResponse]](w)
			_ = jsonResponse.Respond(dto2.ApiResponse[InvoiceResponse]{
				Payload: &InvoiceResponse{
					PlanAssigned: false,
				},
				ErrorCode:    400,
				ErrorMessage: "you are not allowed to assign this plan more than once",
			})
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[InvoiceResponse]](w)
		_ = jsonResponse.Respond(dto2.ApiResponse[InvoiceResponse]{
			Payload: &InvoiceResponse{
				PlanAssigned: false,
			},
			ErrorCode:    500,
			ErrorMessage: "server failed to process request",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[InvoiceResponse]](w)
	_ = jsonResponse.Respond(dto2.ApiResponse[InvoiceResponse]{
		Payload: &InvoiceResponse{
			PlanAssigned: true,
		},
		ErrorCode:    0,
		ErrorMessage: "",
	})
}
