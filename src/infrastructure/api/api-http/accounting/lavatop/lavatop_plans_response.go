package lavatop

import (
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
	"goproxy/infrastructure/dto"
)

type PlansResponse struct {
	plansRepository     application.PlanRepository
	lavaTopUseCases     application.LavaTopUseCases
	planOfferRepository application.PlanOfferRepository
}

func NewPlansResponse(plansRepository application.PlanRepository, lavaTopUseCases application.LavaTopUseCases,
	planOfferRepository application.PlanOfferRepository) PlansResponse {
	return PlansResponse{
		plansRepository:     plansRepository,
		lavaTopUseCases:     lavaTopUseCases,
		planOfferRepository: planOfferRepository,
	}
}

func (h *PlansResponse) Build() (dto.ApiResponse[[]dto.Plan], error) {
	plans, err := h.plansRepository.GetAllWithFeatures()
	if err != nil {
		return dto.ApiResponse[[]dto.Plan]{}, err
	}

	planPrices := h.extractPlanPrices(plans)

	response := dto.ApiResponse[[]dto.Plan]{
		Payload:      nil,
		ErrorCode:    0,
		ErrorMessage: "",
	}

	response.Payload = h.buildPlanResponses(plans, planPrices)

	return response, nil
}

func (h *PlansResponse) extractPlanPrices(plans []aggregates.Plan) map[int][]dto.Price {
	planPrices := make(map[int][]dto.Price)
	lavatopOffers, lavatopOffersErr := h.lavaTopUseCases.GetOffers()
	if lavatopOffersErr != nil {
		return planPrices
	}

	for _, plan := range plans {
		planOfferIds, planOfferIdsErr := h.planOfferRepository.GetOffers(plan.Id())
		if planOfferIdsErr != nil {
			continue
		}

		for _, offer := range lavatopOffers {
			for _, planOffer := range planOfferIds {
				if offer.ExtId() == planOffer.OfferId() {
					for _, price := range offer.Prices() {
						planPrices[plan.Id()] = append(planPrices[plan.Id()], dto.Price{
							Currency: price.Currency().String(),
							Cents:    price.Cents(),
						})
					}
				}
			}
		}
	}

	return planPrices
}

func (h *PlansResponse) buildPlanResponses(plans []aggregates.Plan, planPrices map[int][]dto.Price) *[]dto.Plan {
	planResponses := make([]dto.Plan, len(plans))

	for i, plan := range plans {
		features := h.buildFeatures(plan.Features())
		prices := h.getPlanPrices(plan.Id(), planPrices)

		planResponses[i] = dto.Plan{
			Name: plan.Name(),
			Limits: dto.Limits{
				Bandwidth: dto.BandwidthLimit{
					IsLimited: plan.LimitBytes() != 0,
					Used:      0,
					Total:     plan.LimitBytes(),
				},
				Connections: dto.ConnectionLimit{
					IsLimited:                true,
					MaxConcurrentConnections: 25,
				},
				Speed: dto.SpeedLimit{
					IsLimited:         false,
					MaxBytesPerSecond: 125_000_000, // 1 Gbps
				},
			},
			Features:     features,
			DurationDays: plan.DurationDays(),
			Prices:       prices,
		}
	}

	return &planResponses
}

func (h *PlansResponse) extractPlanFeatures(plans []aggregates.Plan) map[int][]string {
	planFeatures := make(map[int][]string)

	for _, plan := range plans {
		features := make([]string, len(plan.Features()))
		for i, feature := range plan.Features() {
			features[i] = feature.Feature()
		}
		planFeatures[plan.Id()] = features
	}

	return planFeatures
}

func (h *PlansResponse) buildFeatures(features []valueobjects.PlanFeature) []dto.Feature {
	result := make([]dto.Feature, len(features))
	for i, feature := range features {
		result[i] = dto.Feature{
			Feature:            feature.Feature(),
			FeatureDescription: feature.Description(),
		}
	}
	return result
}

func (h *PlansResponse) getPlanPrices(planID int, planPrices map[int][]dto.Price) []dto.Price {
	prices := planPrices[planID]
	if prices == nil {
		return []dto.Price{
			{Cents: 0, Currency: "RUB"},
			{Cents: 0, Currency: "USD"},
			{Cents: 0, Currency: "EUR"},
		}
	}
	return prices
}
