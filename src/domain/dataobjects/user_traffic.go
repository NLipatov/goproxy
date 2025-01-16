package dataobjects

import "time"

type UserTraffic struct {
	InBytes        int64
	OutBytes       int64
	PlanLimitBytes int64
	ActualizedAt   time.Time
}
