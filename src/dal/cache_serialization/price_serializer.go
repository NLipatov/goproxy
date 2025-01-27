package cache_serialization

import "goproxy/domain/dataobjects"

type PriceDto struct {
	Id       int
	PlanId   int
	Cents    int64
	Currency string
}

type PriceCacheSerializer struct {
}

func NewPriceCacheSerializer() CacheSerializer[dataobjects.PlanPrice, PriceDto] {
	return &PriceCacheSerializer{}
}

func (p *PriceCacheSerializer) ToT(dto PriceDto) dataobjects.PlanPrice {
	return dataobjects.NewPlanPrice(dto.Id, dto.PlanId, dto.Cents, dto.Currency)
}

func (p *PriceCacheSerializer) ToD(price dataobjects.PlanPrice) PriceDto {
	return PriceDto{
		Id:       price.Id(),
		PlanId:   price.PlanId(),
		Cents:    price.Cents(),
		Currency: price.Currency(),
	}
}

func (p *PriceCacheSerializer) ToTArray(dto []PriceDto) []dataobjects.PlanPrice {
	result := make([]dataobjects.PlanPrice, len(dto))
	for i, d := range dto {
		result[i] = p.ToT(d)
	}
	return result
}

func (p *PriceCacheSerializer) ToDArray(dto []dataobjects.PlanPrice) []PriceDto {
	result := make([]PriceDto, len(dto))
	for i, d := range dto {
		result[i] = p.ToD(d)
	}
	return result
}
