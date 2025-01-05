package events

type UserConsumedTrafficWithoutPlan struct {
	UserId int
}

func NewUserConsumedTrafficWithoutPlan(userId int) UserConsumedTrafficWithoutPlan {
	return UserConsumedTrafficWithoutPlan{
		UserId: userId,
	}
}
