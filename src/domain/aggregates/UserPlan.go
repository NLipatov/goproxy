package aggregates

import "time"

type UserPlan struct {
	id        int
	userId    int
	planId    int
	validTo   time.Time
	createdAt time.Time
}

func NewUserPlan(id, userId, planId int, validTo, createdAt time.Time) (UserPlan, error) {
	return UserPlan{
		id:        id,
		userId:    userId,
		planId:    planId,
		validTo:   validTo,
		createdAt: createdAt,
	}, nil
}

func (up *UserPlan) Id() int {
	return up.id
}

func (up *UserPlan) UserId() int {
	return up.userId
}

func (up *UserPlan) PlanId() int {
	return up.planId
}

func (up *UserPlan) ValidTo() time.Time {
	return up.validTo
}

func (up *UserPlan) ProlongDays(days int) {
	up.validTo = up.validTo.AddDate(0, 0, days)
}
