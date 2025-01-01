package events

import "time"

type UserConsumedTrafficEvent struct {
	UserId    int
	Timestamp time.Time
	InBytes   int
	OutBytes  int
}

func NewUserConsumedTrafficEvent(userId, in, out int) UserConsumedTrafficEvent {
	return UserConsumedTrafficEvent{
		UserId:    userId,
		Timestamp: time.Now().UTC(),
		InBytes:   in,
		OutBytes:  out,
	}
}
