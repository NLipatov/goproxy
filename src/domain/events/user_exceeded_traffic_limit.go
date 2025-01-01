package events

import "time"

type UserExceededTrafficLimitEvent struct {
	UserId    int
	Timestamp time.Time
}
