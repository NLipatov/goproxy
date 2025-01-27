package handlers

import (
	"encoding/json"
	"goproxy/application"
	"goproxy/infrastructure/api/api-http/plans/plans_dto"
	"log"
	"net/http"
	"strconv"
)

type GetPlanPricesHandler struct {
	planPriceRepository application.PlanPriceRepository
}

func NewGetPlanPricesHandler(planPriceRepository application.PlanPriceRepository) GetPlanPricesHandler {
	return GetPlanPricesHandler{
		planPriceRepository: planPriceRepository,
	}
}

func (g *GetPlanPricesHandler) Handle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		log.Printf("GetPlanPricesHandler.Handle: 405 as invalid method (%s)", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	planId := r.URL.Query().Get("plan_id")
	if planId == "" {
		log.Printf("GetPlanPricesHandler.Handle: 400 as no 'plan_id' provided")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	planIdInt, err := strconv.Atoi(planId)
	if err != nil {
		log.Printf("GetPlanPricesHandler.Handle: 400 as 'plan_id' is not a number (%s)", planId)
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	prices, pricesErr := g.planPriceRepository.GetAllWithPlanId(planIdInt)
	if pricesErr != nil {
		log.Printf("GetPlanPricesHandler.Handle: 500 failed to retrieve prices: %s", pricesErr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(plans_dto.ToDArray(prices))
}
