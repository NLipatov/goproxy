package lavatopvalueobjects

import (
	"errors"
	"strings"
)

type PaymentMethod int

const (
	BANK131 PaymentMethod = iota
	UNLIMINT
	PAYPAL
	STRIPE
)

func (c PaymentMethod) String() string {
	switch c {
	case BANK131:
		return "BANK131"
	case UNLIMINT:
		return "UNLIMINT"
	case PAYPAL:
		return "PAYPAL"
	case STRIPE:
		return "STRIPE"
	}

	return ""
}

func ParsePaymentMethod(input string) (PaymentMethod, error) {
	switch strings.ToUpper(input) {
	case "BANK131":
		return BANK131, nil
	case "UNLIMINT":
		return UNLIMINT, nil
	case "PAYPAL":
		return PAYPAL, nil
	case "STRIPE":
		return STRIPE, nil
	default:
		return 0, errors.New("invalid payment method")
	}
}
