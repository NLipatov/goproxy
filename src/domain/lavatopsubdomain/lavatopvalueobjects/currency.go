package lavatopvalueobjects

import (
	"errors"
	"strings"
)

type Currency int

const (
	RUB Currency = iota
	USD
	EUR
)

func (c Currency) String() string {
	switch c {
	case RUB:
		return "RUB"
	case USD:
		return "USD"
	case EUR:
		return "EUR"
	}

	return ""
}

func ParseCurrency(input string) (Currency, error) {
	switch strings.ToUpper(input) {
	case "RUB":
		return RUB, nil
	case "USD":
		return USD, nil
	case "EUR":
		return EUR, nil
	default:
		return 0, errors.New("invalid currency")
	}
}
