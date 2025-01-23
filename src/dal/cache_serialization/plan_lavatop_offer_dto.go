package cache_serialization

import (
	"goproxy/domain/dataobjects"
)

type PlanLavatopOfferDto struct {
	Id      int
	PlanId  int
	OfferId string
}

func NewPlanLavatopOfferSerializer() CacheSerializer[dataobjects.PlanLavatopOffer, PlanLavatopOfferDto] {
	return &PlanLavatopOfferDto{}
}

func (p PlanLavatopOfferDto) ToT(dto PlanLavatopOfferDto) dataobjects.PlanLavatopOffer {
	return dataobjects.NewPlanLavatopOffer(dto.Id, dto.PlanId, dto.OfferId)
}

func (p PlanLavatopOfferDto) ToD(offer dataobjects.PlanLavatopOffer) PlanLavatopOfferDto {
	return PlanLavatopOfferDto{
		Id:      offer.Id(),
		PlanId:  offer.PlanId(),
		OfferId: offer.OfferId(),
	}
}

func (p PlanLavatopOfferDto) ToTArray(dto []PlanLavatopOfferDto) []dataobjects.PlanLavatopOffer {
	result := make([]dataobjects.PlanLavatopOffer, len(dto))
	for i, d := range dto {
		result[i] = d.ToT(d)
	}
	return result
}

func (p PlanLavatopOfferDto) ToDArray(plans []dataobjects.PlanLavatopOffer) []PlanLavatopOfferDto {
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
