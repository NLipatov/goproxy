package cache_serialization

import (
	"goproxy/domain/dataobjects"
)

type PlanLavatopOfferDto struct {
	Id      int
	PlanId  int
	OfferId string
}

type DataObjectsPlanLavatopOfferCacheSerializer struct {
}

func NewPlanLavatopOfferSerializer() CacheSerializer[dataobjects.PlanLavatopOffer, PlanLavatopOfferDto] {
	return &DataObjectsPlanLavatopOfferCacheSerializer{}
}

func (p *DataObjectsPlanLavatopOfferCacheSerializer) ToT(dto PlanLavatopOfferDto) dataobjects.PlanLavatopOffer {
	return dataobjects.NewPlanLavatopOffer(dto.Id, dto.PlanId, dto.OfferId)
}

func (p *DataObjectsPlanLavatopOfferCacheSerializer) ToD(offer dataobjects.PlanLavatopOffer) PlanLavatopOfferDto {
	return PlanLavatopOfferDto{
		Id:      offer.Id(),
		PlanId:  offer.PlanId(),
		OfferId: offer.OfferId(),
	}
}

func (p *DataObjectsPlanLavatopOfferCacheSerializer) ToTArray(dto []PlanLavatopOfferDto) []dataobjects.PlanLavatopOffer {
	result := make([]dataobjects.PlanLavatopOffer, len(dto))
	for i, d := range dto {
		result[i] = p.ToT(d)
	}
	return result
}

func (p *DataObjectsPlanLavatopOfferCacheSerializer) ToDArray(plans []dataobjects.PlanLavatopOffer) []PlanLavatopOfferDto {
	result := make([]PlanLavatopOfferDto, len(plans))
	for i, plan := range plans {
		result[i] = PlanLavatopOfferDto{
			Id:      plan.Id(),
			PlanId:  plan.PlanId(),
			OfferId: plan.OfferId(),
		}
	}
	return result
}
