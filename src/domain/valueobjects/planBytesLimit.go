package valueobjects

import "fmt"

type PlanBytesLimit struct {
	value int64
}

func PlanBytesLimitFromInt64(val int64) (PlanBytesLimit, error) {
	if val < 0 {
		return PlanBytesLimit{}, fmt.Errorf("value can not be less than 0")
	}

	return PlanBytesLimit{
		value: val,
	}, nil
}

func (p *PlanBytesLimit) Value() int64 {
	return p.value
}
