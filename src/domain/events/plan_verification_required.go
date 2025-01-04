package events

type PlanVerificationRequired struct {
	UserId int
}

func NewPlanVerificationRequired(userId int) PlanVerificationRequired {
	return PlanVerificationRequired{
		UserId: userId,
	}

}
