package lavatopvalueobjects

import (
	"errors"
	"strings"
)

type Periodicity int

const (
	ONE_TIME Periodicity = iota
	MONTHLY
	PERIOD_90_DAYS
	PERIOD_180_DAYS
	PERIOD_YEAR
)

func (p Periodicity) String() string {
	switch p {
	case ONE_TIME:
		return "ONE_TIME"
	case MONTHLY:
		return "MONTHLY"
	case PERIOD_90_DAYS:
		return "PERIOD_90_DAYS"
	case PERIOD_180_DAYS:
		return "PERIOD_180_DAYS"
	case PERIOD_YEAR:
		return "PERIOD_YEAR"
	}

	return ""
}

func ParsePeriodicity(input string) (Periodicity, error) {
	switch strings.ToUpper(input) {
	case "ONE_TIME":
		return ONE_TIME, nil
	case "MONTHLY":
		return MONTHLY, nil
	case "PERIOD_90_DAYS":
		return PERIOD_90_DAYS, nil
	case "PERIOD_180_DAYS":
		return PERIOD_180_DAYS, nil
	case "PERIOD_YEAR":
		return PERIOD_YEAR, nil
	default:
		return 0, errors.New("invalid period")
	}
}
