package events

import "time"

type UserExceededTrafficLimitEvent struct {
	UserId    int
	Timestamp time.Time
}

func NewUserExceededTrafficLimitEvent(userId int) UserExceededTrafficLimitEvent {
	return UserExceededTrafficLimitEvent{
		UserId:    userId,
		Timestamp: time.Now().UTC(),
	}
}
