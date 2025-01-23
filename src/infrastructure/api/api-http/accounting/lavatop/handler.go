package lavatop

import (
	"encoding/json"
	"github.com/golang-jwt/jwt/v4"
	"goproxy/application"
	"goproxy/domain/lavatopsubdomain/lavatopaggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
	"goproxy/infrastructure/api/api-http/google_auth"
	"goproxy/infrastructure/dto"
	"log"
	"net/http"
)

type Handler struct {
	billingService      application.BillingService[lavatopaggregates.Invoice, lavatopvalueobjects.Offer]
	userUseCases        application.UserUseCases
	plansRepository     application.PlanRepository
	planOfferRepository application.PlanOfferRepository
}

func NewHandler(billingService application.BillingService[lavatopaggregates.Invoice, lavatopvalueobjects.Offer],
	planRepository application.PlanRepository, planOfferRepository application.PlanOfferRepository) *Handler {
	return &Handler{
		billingService:      billingService,
		plansRepository:     planRepository,
		planOfferRepository: planOfferRepository,
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

func (h Handler) PostInvoices(writer http.ResponseWriter, request *http.Request) {
	panic("not implemented")
}

func (h Handler) GetPlans(w http.ResponseWriter, _ *http.Request) {
	response := dto.ApiResponse[[]dto.Plan]{
		Payload:      nil,
		ErrorCode:    0,
		ErrorMessage: "",
	}

	plans, plansErr := h.plansRepository.GetAllWithFeatures()
	if plansErr != nil {
		response.ErrorCode = http.StatusInternalServerError
		response.ErrorMessage = "could not load plans"
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(response)
		return
	}

	planFeatures := make(map[int][]string)
	for _, plan := range plans {
		features := make([]string, len(plan.Features()))
		for fi, feature := range plan.Features() {
			features[fi] = feature.Feature()
		}
		planFeatures[plan.Id()] = features
	}
	planResponses := make([]dto.Plan, len(plans))
	for i, plan := range plans {
		planResponses[i] = dto.Plan{
			Name:     plan.Name(),
			Limits:   dto.Limits{},
			Features: planFeatures[plan.Id()],
		}
	}

	response.Payload = &planResponses

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}
