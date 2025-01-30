package events

import "time"

type PlanAssigned struct {
	UserEmail string
	PlanId    int
	Timestamp time.Time
}

func NewPlanAssigned(userEmail string, planId int, timestamp time.Time) PlanAssigned {
	return PlanAssigned{
		UserEmail: userEmail,
		PlanId:    planId,
		Timestamp: timestamp,
	}
}
