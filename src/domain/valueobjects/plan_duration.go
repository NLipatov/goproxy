package valueobjects

import (
	"fmt"
	"time"
)

const minimalPlanDuration = time.Hour * 24

type PlanDuration struct {
	durationDays int
}

func ParsePlanDurationFromDays(days int) (PlanDuration, error) {
	duration := time.Hour * 24 * time.Duration(days)
	if minimalPlanDuration > duration {
		return PlanDuration{}, fmt.Errorf("plan duration can not be less than %v", minimalPlanDuration)
	}

	return PlanDuration{
		durationDays: days,
	}, nil
}

func (p *PlanDuration) DurationDays() int {
	return p.durationDays
}
