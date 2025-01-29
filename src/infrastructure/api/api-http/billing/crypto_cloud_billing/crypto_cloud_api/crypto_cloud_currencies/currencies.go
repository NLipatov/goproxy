package crypto_cloud_currencies

import (
	"fmt"
	"strings"
)

type CryptoCloudCurrency int

const (
	USD CryptoCloudCurrency = iota
	UZS
	KGS
	KZT
	AMD
	AZN
	BYN
	AUD
	TRY
	AED
	CAD
	CNY
	HKD
	IDR
	INR
	JPY
	PHP
	SGD
	THB
	VND
	MYR
	RUB
	UAH
	EUR
	GBP
)

var currencyStrings = []string{
	"USD", "UZS", "KGS", "KZT", "AMD", "AZN", "BYN", "AUD", "TRY", "AED",
	"CAD", "CNY", "HKD", "IDR", "INR", "JPY", "PHP", "SGD", "THB", "VND",
	"MYR", "RUB", "UAH", "EUR", "GBP",
}

func NewCryptoCloudCurrency(stringCode string) CryptoCloudCurrency {
	stringCode = strings.ToUpper(stringCode)
	for i, code := range currencyStrings {
		if code == stringCode {
			return CryptoCloudCurrency(i)
		}
	}

	return USD
}

func (c CryptoCloudCurrency) String() (string, error) {
	if int(c) >= 0 && int(c) < len(currencyStrings) {
		return currencyStrings[c], nil
	}
	return "", fmt.Errorf("invalid CryptoCloudCurrency code: %d", c)
}
