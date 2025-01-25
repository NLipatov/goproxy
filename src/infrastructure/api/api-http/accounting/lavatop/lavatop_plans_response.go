package lavatop

import (
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/lavatopsubdomain/lavatopvalueobjects"
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

	planOffers := h.extractPlanOffers(plans)

	response := dto.ApiResponse[[]dto.Plan]{
		Payload:      nil,
		ErrorCode:    0,
		ErrorMessage: "",
	}

	response.Payload = h.buildPlanResponses(plans, planOffers)

	return response, nil
}

func (h *PlansResponse) extractPlanOffers(plans []aggregates.Plan) map[int][]dto.Offer {
	planOffers := make(map[int][]dto.Offer)
	lavatopOffers, lavatopOffersErr := h.lavaTopUseCases.GetOffers()
	if lavatopOffersErr != nil {
		return planOffers
	}

	for _, plan := range plans {
		planOfferIds, planOfferIdsErr := h.planOfferRepository.GetOffers(plan.Id())
		if planOfferIdsErr != nil {
			continue
		}

		for _, offer := range lavatopOffers {
			for _, planOffer := range planOfferIds {
				if offer.ExtId() == planOffer.OfferId() {

					prices := make([]dto.Price, len(offer.Prices()))
					for i, price := range offer.Prices() {
						var paymentMethods []string
						switch price.Currency() {
						case lavatopvalueobjects.RUB:
							paymentMethods = make([]string, 1)
							paymentMethods[0] = "BANK131"

						case lavatopvalueobjects.USD:
							paymentMethods = make([]string, 3)
							paymentMethods[0] = "PAYPAL"
							paymentMethods[1] = "UNLIMINT"
							paymentMethods[2] = "STRIPE"

						case lavatopvalueobjects.EUR:
							paymentMethods = make([]string, 3)
							paymentMethods[0] = "PAYPAL"
							paymentMethods[1] = "UNLIMINT"
							paymentMethods[2] = "STRIPE"

						default:
							continue
						}

						prices[i] = dto.Price{
							Cents:          price.Cents(),
							Currency:       price.Currency().String(),
							PaymentMethods: paymentMethods,
						}
					}

					planOffers[plan.Id()] = append(planOffers[plan.Id()], dto.Offer{
						Description: "",
						OfferId:     offer.ExtId(),
						Prices:      prices,
					})
				}
			}
		}
	}

	return planOffers
}

func (h *PlansResponse) buildPlanResponses(plans []aggregates.Plan, planOffers map[int][]dto.Offer) *[]dto.Plan {
	planResponses := make([]dto.Plan, len(plans))

	for i, plan := range plans {
		features := h.buildFeatures(plan.Features())

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
			Offers:       planOffers[plan.Id()],
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
