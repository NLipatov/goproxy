package free_plan_billing

import (
	"encoding/json"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/api/api-http/billing/free_plan_billing/free_plan_billing_dtos"
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

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		h.respondError(w, http.StatusMethodNotAllowed, "")
		return
	}

	userRequest := http_objects.NewAuthenticatedUserBillingRequest(r)
	email, emailErr := userRequest.UserEmail()
	if emailErr != nil {
		h.respondError(w, http.StatusUnauthorized, "")
		return
	}

	emailVO, emailVOErr := valueobjects.ParseEmailFromString(email)
	if emailVOErr != nil {
		h.respondError(w, http.StatusUnauthorized, "invalid email")
		return
	}

	var requestDto free_plan_billing_dtos.Request
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64))
	if err := decoder.Decode(&requestDto); err != nil {
		h.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	handleErr := h.service.handle(emailVO, requestDto.PlanId)
	if handleErr != nil {
		if handleErr.Error() == "not eligible for free plan" {
			h.respondError(w, http.StatusBadRequest, "free plan can only be activated once")
			return
		}

		h.respondError(w, http.StatusInternalServerError, "")
		return
	}

	h.respond(w, free_plan_billing_dtos.Response{PlanAssigned: true})
}

func (h *Handler) respondError(w http.ResponseWriter, statusCode int, msg string) {
	w.WriteHeader(statusCode)
	jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[free_plan_billing_dtos.Response]](w)
	_ = jsonResponse.Respond(dto2.ApiResponse[free_plan_billing_dtos.Response]{
		Payload: &free_plan_billing_dtos.Response{
			PlanAssigned: false,
		},
		ErrorCode:    0,
		ErrorMessage: msg,
	})
}

func (h *Handler) respond(w http.ResponseWriter, response free_plan_billing_dtos.Response) {
	w.WriteHeader(http.StatusOK)
	jsonResponse := http_objects.NewJSONResponse[dto2.ApiResponse[free_plan_billing_dtos.Response]](w)
	_ = jsonResponse.Respond(dto2.ApiResponse[free_plan_billing_dtos.Response]{
		Payload:      &response,
		ErrorCode:    0,
		ErrorMessage: "",
	})
}
