package lavatopvalueobjects

import (
	"errors"
	"strings"
)

type BuyerLanguage int

const (
	EN BuyerLanguage = iota
	RU
	ES
)

func (l BuyerLanguage) String() string {
	switch l {
	case EN:
		return "EN"
	case RU:
		return "RU"
	case ES:
		return "ES"
	}

	return ""
}

func ParseBuyerLanguage(input string) (BuyerLanguage, error) {
	switch strings.ToUpper(input) {
	case "EN":
		return EN, nil
	case "RU":
		return RU, nil
	case "ES":
		return ES, nil
	default:
		return 0, errors.New("invalid buyer language")
	}
}
