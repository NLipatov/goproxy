package lavatop

import (
	"encoding/json"
	"fmt"
	"github.com/golang-jwt/jwt/v4"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/infrastructure/api/api-http/google_auth"
	"goproxy/infrastructure/dto"
	"io"
	"log"
	"net/http"
	"os"
)

type Handler struct {
	billingService      application.BillingService[lavatopaggregates.Invoice, lavatopvalueobjects.Offer]
	userUseCases        application.UserUseCases
	plansRepository     application.PlanRepository
	planOfferRepository application.PlanOfferRepository
	lavaTopUseCases     application.LavaTopUseCases
	plansResponse       PlansResponse
}

func NewHandler(billingService application.BillingService[lavatopaggregates.Invoice, lavatopvalueobjects.Offer],
	planRepository application.PlanRepository, planOfferRepository application.PlanOfferRepository,
	lavaTopUseCases application.LavaTopUseCases, userUseCases application.UserUseCases) *Handler {

	plansResponse := NewPlansResponse(planRepository, lavaTopUseCases, planOfferRepository)

	return &Handler{
		billingService:      billingService,
		plansRepository:     planRepository,
		planOfferRepository: planOfferRepository,
		lavaTopUseCases:     lavaTopUseCases,
		plansResponse:       plansResponse,
		userUseCases:        userUseCases,
	}
}

func (h Handler) GetOffers(w http.ResponseWriter, _ *http.Request) {
	offers := getMockedOffers()
	offerResponses := make([]dto.OfferResponse, len(offers))
	for i, o := range offers {
		offerResponses[i] = dto.ToOfferResponse(o)
	}

	response := dto.ApiResponse[[]dto.OfferResponse]{
		Payload:      &offerResponses,
		ErrorCode:    0,
		ErrorMessage: "",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func getMockedOffers() []lavatopvalueobjects.Offer {
	plusPrices := make([]lavatopvalueobjects.Price, 3)
	plusPrices[0] = lavatopvalueobjects.NewPrice(50, lavatopvalueobjects.EUR, lavatopvalueobjects.ONE_TIME)
	plusPrices[1] = lavatopvalueobjects.NewPrice(5000, lavatopvalueobjects.RUB, lavatopvalueobjects.ONE_TIME)
	plusPrices[2] = lavatopvalueobjects.NewPrice(50, lavatopvalueobjects.USD, lavatopvalueobjects.ONE_TIME)

	proPrices := make([]lavatopvalueobjects.Price, 3)
	proPrices[0] = lavatopvalueobjects.NewPrice(100, lavatopvalueobjects.EUR, lavatopvalueobjects.ONE_TIME)
	proPrices[1] = lavatopvalueobjects.NewPrice(10000, lavatopvalueobjects.RUB, lavatopvalueobjects.ONE_TIME)
	proPrices[2] = lavatopvalueobjects.NewPrice(100, lavatopvalueobjects.USD, lavatopvalueobjects.ONE_TIME)

	proMaxPrices := make([]lavatopvalueobjects.Price, 3)
	proMaxPrices[0] = lavatopvalueobjects.NewPrice(100, lavatopvalueobjects.EUR, lavatopvalueobjects.ONE_TIME)
	proMaxPrices[1] = lavatopvalueobjects.NewPrice(10000, lavatopvalueobjects.RUB, lavatopvalueobjects.ONE_TIME)
	proMaxPrices[2] = lavatopvalueobjects.NewPrice(100, lavatopvalueobjects.USD, lavatopvalueobjects.ONE_TIME)

	offers := make([]lavatopvalueobjects.Offer, 3)
	offers[0] = lavatopvalueobjects.NewOffer("3fbd90cb-d357-45aa-a10e-b0b0b7eee808", "Plus", plusPrices)
	offers[1] = lavatopvalueobjects.NewOffer("846b6a2f-2f5e-486b-b170-64a62c457c3d", "Pro", proPrices)
	offers[2] = lavatopvalueobjects.NewOffer("960a9763-221b-4fb7-818f-867c57f6fcb5", "Pro Max", proMaxPrices)

	return offers
}

func (h Handler) GetInvoices(w http.ResponseWriter, r *http.Request) {
	idToken, err := google_auth.GetIdTokenFromCookie(r)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	verifiedToken, err := google_auth.VerifyIDToken(idToken)
	if err != nil {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok {
		http.Error(w, "Failed to parse token claims", http.StatusInternalServerError)
		return
	}

	email := claims["email"].(string)
	if email == "" {
		log.Printf("failed to reset proxy password: email claim empty")
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}

	user, userErr := h.userUseCases.GetByEmail(email)
	if userErr != nil {
		log.Printf("failed to reset proxy user - failed to fetch user: %s", userErr)
		http.Error(w, "Failed to reset password", http.StatusInternalServerError)
		return
	}
	log.Printf("%v", user)
	panic("not implemented")
}

func (h Handler) PostInvoice(w http.ResponseWriter, r *http.Request) {
	user, userErr := h.getUser(r)
	if userErr != nil {
		w.WriteHeader(http.StatusUnauthorized)
		_ = json.NewEncoder(w).Encode(dto.ApiResponse[dto.PostInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    http.StatusUnauthorized,
			ErrorMessage: "not authorized",
		})
		return
	}

	var cmd dto.AccountingIssueInvoiceCommand
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 512))
	if err := decoder.Decode(&cmd); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ApiResponse[dto.PostInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid body",
		})
		return
	}

	currency, currencyErr := lavatopvalueobjects.ParseCurrency(cmd.Currency)
	if currencyErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ApiResponse[dto.PostInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid currency",
		})
		return
	}

	paymentMethod, paymentMethodErr := lavatopvalueobjects.ParsePaymentMethod(cmd.PaymentMethod)
	if paymentMethodErr != nil {
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(dto.ApiResponse[dto.PostInvoiceResponse]{
			Payload:      nil,
			ErrorCode:    http.StatusBadRequest,
			ErrorMessage: "invalid payment method",
		})
		return
	}

	newIssueInvoiceResponse := NewIssueInvoiceResponse(h.lavaTopUseCases, user, currency, paymentMethod, cmd.OfferId)

	response, responseErr := newIssueInvoiceResponse.Build()
	if responseErr != nil {
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h Handler) GetPlans(w http.ResponseWriter, _ *http.Request) {
	response, responseErr := h.plansResponse.Build()
	if responseErr != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(dto.ApiResponse[[]dto.Plan]{
			Payload:      nil,
			ErrorCode:    http.StatusInternalServerError,
			ErrorMessage: "could not load plans",
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

func (h Handler) getUser(r *http.Request) (aggregates.User, error) {
	idToken, err := google_auth.GetIdTokenFromCookie(r)
	if err != nil {
		return aggregates.User{}, err
	}

	verifiedToken, err := google_auth.VerifyIDToken(idToken)
	if err != nil {
		return aggregates.User{}, fmt.Errorf("failed to verify token: %w", err)
	}

	claims, ok := verifiedToken.Claims.(jwt.MapClaims)
	if !ok {
		return aggregates.User{}, fmt.Errorf("failed to parse token claims")
	}

	email := claims["email"].(string)
	if email == "" {
		return aggregates.User{}, fmt.Errorf("email claim empty")
	}

	usersApiHost := os.Getenv("USERS_API_HOST")
	if usersApiHost == "" {
		return aggregates.User{}, fmt.Errorf("users api host empty")
	}

	resp, err := http.Get(fmt.Sprintf("%s/users/get?email=%s", usersApiHost, email))
	if err != nil {
		return aggregates.User{}, fmt.Errorf("failed to fetch user id: %v", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	body, bodyErr := io.ReadAll(resp.Body)
	if bodyErr != nil {
		return aggregates.User{}, fmt.Errorf("failed to read response body: %v", err)
	}

	var userResult dto.GetUserResult
	deserializationErr := json.Unmarshal(body, &userResult)
	if deserializationErr != nil {
		return aggregates.User{}, fmt.Errorf("failed to deserialize user result: %v", deserializationErr)
	}

	user, userErr := aggregates.NewUser(userResult.Id, userResult.Username, email, email)
	if userErr != nil {
		return aggregates.User{}, fmt.Errorf("failed to load user: %v", userErr)
	}

	return user, nil
}
