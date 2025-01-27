package plans_dto

import "goproxy/domain/dataobjects"

type PlanPriceDto struct {
	Cents    int64  `json:"cents"`
	Currency string `json:"currency"`
}

func ToD(price dataobjects.PlanPrice) PlanPriceDto {
	return PlanPriceDto{
		Cents:    price.Cents(),
		Currency: price.Currency(),
	}
}

func ToDArray(price []dataobjects.PlanPrice) []PlanPriceDto {
	result := make([]PlanPriceDto, len(price))
	for i, p := range price {
		result[i] = ToD(p)
	}
	return result
}
