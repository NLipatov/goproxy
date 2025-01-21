package valueobjects

type PlanFeature struct {
	planId  int
	feature string
}

func NewPlanFeature(planId int, feature string) PlanFeature {
	return PlanFeature{
		planId:  planId,
		feature: feature,
	}
}

func (pf *PlanFeature) Feature() string {
	return pf.feature
}

func (pf *PlanFeature) PlanId() int {
	return pf.planId
}
