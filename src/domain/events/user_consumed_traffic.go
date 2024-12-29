package events

import "time"

type UserConsumedTrafficEvent struct {
	UserId    int
	Timestamp time.Time
	InMb      int
	OutMb     int
}

func NewUserConsumedTrafficEvent(userId, inMb, outMb int) UserConsumedTrafficEvent {
	return UserConsumedTrafficEvent{
		UserId:    userId,
		Timestamp: time.Now().UTC(),
		InMb:      inMb,
		OutMb:     outMb,
	}
}
