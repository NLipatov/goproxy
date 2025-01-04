package aggregates

import (
	"goproxy/domain/valueobjects"
	"time"
)

type Plan struct {
	id           int
	name         string
	limitBytes   valueobjects.PlanBytesLimit
	durationDays int
	createdAt    time.Time
}

func NewPlan(id int, name string, limitBytes int64, durationDays int) (Plan, error) {
	limit, limitErr := valueobjects.PlanBytesLimitFromInt64(limitBytes)
	if limitErr != nil {
		return Plan{}, limitErr
	}

	return Plan{
		id:           id,
		name:         name,
		limitBytes:   limit,
		durationDays: durationDays,
		createdAt:    time.Now().UTC(),
	}, nil
}

func (p *Plan) Id() int {
	return p.id
}

func (p *Plan) Name() string {
	return p.name
}

func (p *Plan) LimitBytes() int64 {
	return p.limitBytes.Value()
}

func (p *Plan) DurationDays() int {
	return p.durationDays
}

func (p *Plan) CreatedAt() time.Time {
	return p.createdAt
}
