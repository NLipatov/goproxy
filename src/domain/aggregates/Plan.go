package aggregates

import (
	"goproxy/domain/valueobjects"
	"time"
)

type Plan struct {
	id         int
	name       valueobjects.PlanName
	limitBytes valueobjects.PlanBytesLimit
	duration   valueobjects.PlanDuration
	features   []valueobjects.PlanFeature
	createdAt  time.Time
}

func NewPlan(id int, name string, limitBytes int64, durationDays int, features []valueobjects.PlanFeature) (Plan, error) {
	limit, limitErr := valueobjects.PlanBytesLimitFromInt64(limitBytes)
	if limitErr != nil {
		return Plan{}, limitErr
	}

	planName, nameErr := valueobjects.ParsePlanNameFromString(name)
	if nameErr != nil {
		return Plan{}, nameErr
	}

	planDuration, durationErr := valueobjects.ParsePlanDurationFromDays(durationDays)
	if durationErr != nil {
		return Plan{}, durationErr
	}

	return Plan{
		id:         id,
		name:       planName,
		limitBytes: limit,
		duration:   planDuration,
		features:   features,
		createdAt:  time.Now().UTC(),
	}, nil
}

func (p *Plan) Id() int {
	return p.id
}

func (p *Plan) Name() string {
	return p.name.Name()
}

func (p *Plan) LimitBytes() int64 {
	return p.limitBytes.Value()
}

func (p *Plan) DurationDays() int {
	return p.duration.DurationDays()
}

func (p *Plan) Features() []valueobjects.PlanFeature {
	return p.features
}

func (p *Plan) CreatedAt() time.Time {
	return p.createdAt
}
