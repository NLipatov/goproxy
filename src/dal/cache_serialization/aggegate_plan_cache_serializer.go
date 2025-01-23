package cache_serialization

import (
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
	"time"
)

type PlanFeatureDto struct {
	PlanId             int    `json:"plan_id"`
	FeatureName        string `json:"name"`
	FeatureDescription string `json:"description"`
}

type PlanDto struct {
	Id         int              `json:"id"`
	Name       string           `json:"name"`
	BytesLimit int64            `json:"bytes_limit"`
	Duration   int              `json:"duration"`
	Features   []PlanFeatureDto `json:"features"`
	CreatedAt  time.Time        `json:"created_at"`
}

type AggegatePlanCacheSerializer struct{}

func NewAggegatePlanCacheSerializer() CacheSerializer[aggregates.Plan, PlanDto] {
	return &AggegatePlanCacheSerializer{}
}

func (a *AggegatePlanCacheSerializer) ToT(dto PlanDto) aggregates.Plan {
	features := make([]valueobjects.PlanFeature, len(dto.Features))
	for i, v := range dto.Features {
		features[i] = valueobjects.NewPlanFeature(v.PlanId, v.FeatureName, v.FeatureDescription)
	}

	plan, _ := aggregates.NewPlan(dto.Id, dto.Name, dto.BytesLimit, dto.Duration, features)
	return plan
}

func (a *AggegatePlanCacheSerializer) ToTArray(dtos []PlanDto) []aggregates.Plan {
	arr := make([]aggregates.Plan, len(dtos))
	for i, dto := range dtos {
		arr[i] = a.ToT(dto)
	}

	return arr
}

func (a *AggegatePlanCacheSerializer) ToD(plan aggregates.Plan) PlanDto {
	features := make([]PlanFeatureDto, len(plan.Features()))
	for i, v := range plan.Features() {
		features[i] = PlanFeatureDto{
			PlanId:             v.PlanId(),
			FeatureName:        v.Feature(),
			FeatureDescription: v.Description(),
		}
	}
	return PlanDto{
		Id:         plan.Id(),
		Name:       plan.Name(),
		BytesLimit: plan.LimitBytes(),
		Duration:   plan.DurationDays(),
		Features:   features,
		CreatedAt:  plan.CreatedAt(),
	}
}

func (a *AggegatePlanCacheSerializer) ToDArray(plans []aggregates.Plan) []PlanDto {
	arr := make([]PlanDto, len(plans))
	for i, plan := range plans {
		arr[i] = a.ToD(plan)
	}

	return arr
}
