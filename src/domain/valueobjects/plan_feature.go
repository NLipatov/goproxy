package valueobjects

type PlanFeature struct {
	planId      int
	feature     string
	description string
}

func NewPlanFeature(planId int, feature, description string) PlanFeature {
	return PlanFeature{
		planId:      planId,
		feature:     feature,
		description: description,
	}
}

func (pf *PlanFeature) Feature() string {
	return pf.feature
}

func (pf *PlanFeature) Description() string {
	return pf.description
}

func (pf *PlanFeature) PlanId() int {
	return pf.planId
}
