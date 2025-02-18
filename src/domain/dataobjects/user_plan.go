package dataobjects

import "time"

type UserPlan struct {
	Name         string
	Bandwidth    int64
	CreatedAt    time.Time
	DurationDays int
}
