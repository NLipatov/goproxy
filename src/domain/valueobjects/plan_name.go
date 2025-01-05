package valueobjects

import "fmt"

type PlanName struct {
	name string
}

func ParsePlanNameFromString(name string) (PlanName, error) {
	if len(name) > 100 {
		return PlanName{}, fmt.Errorf("plan name must no exceed 100 characters length")
	}

	return PlanName{
		name: name,
	}, nil
}

func (p *PlanName) Name() string {
	return p.name
}
