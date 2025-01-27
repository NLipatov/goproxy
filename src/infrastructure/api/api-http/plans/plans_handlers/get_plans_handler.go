package plans_handlers

import (
	"encoding/json"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/infrastructure/dto"
	"log"
	"net/http"
)

type GetPlansHandler struct {
	planRepository application.PlanRepository
}

func NewHandler(planRepository application.PlanRepository) GetPlansHandler {
	return GetPlansHandler{
		planRepository: planRepository,
	}
}

func (h *GetPlansHandler) Handle(w http.ResponseWriter, r *http.Request) {
	plans, plansErr := h.planRepository.GetAllWithFeatures()
	if plansErr != nil {
		log.Printf("Handle: 500 as error %v", plansErr)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("could not get plans"))
		return
	}

	response := h.MapPlanArrayToPlanDtoArray(plans)

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
	return
}

func (h *GetPlansHandler) MapPlanArrayToPlanDtoArray(plans []aggregates.Plan) []dto.PlanDto {
	result := make([]dto.PlanDto, len(plans))
	for i, plan := range plans {
		result[i] = h.MapPlanToPlanDto(plan)
	}

	return result
}

func (h *GetPlansHandler) MapPlanToPlanDto(plan aggregates.Plan) dto.PlanDto {
	features := make([]string, len(plan.Features()))
	for j, feature := range plan.Features() {
		features[j] = feature.Feature()
	}

	return dto.PlanDto{
		Name:         plan.Name(),
		Features:     features,
		DurationDays: plan.DurationDays(),
	}
}
