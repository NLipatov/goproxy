package events

import "time"

type UserConsumedTrafficEvent struct {
	UserId    int
	Timestamp time.Time
	InBytes   int64
	OutBytes  int64
}

func NewUserConsumedTrafficEvent(userId int, in, out int64) UserConsumedTrafficEvent {
	return UserConsumedTrafficEvent{
		UserId:    userId,
		Timestamp: time.Now().UTC(),
		InBytes:   in,
		OutBytes:  out,
	}
}
