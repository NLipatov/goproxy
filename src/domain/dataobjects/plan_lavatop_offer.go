package dataobjects

type PlanLavatopOffer struct {
	id      int
	planId  int
	offerId string
}

func NewPlanLavatopOffer(id, planId int, offerId string) PlanLavatopOffer {
	return PlanLavatopOffer{
		id:      id,
		planId:  planId,
		offerId: offerId,
	}
}

func (po *PlanLavatopOffer) Id() int {
	return po.id
}

func (po *PlanLavatopOffer) PlanId() int {
	return po.planId
}

func (po *PlanLavatopOffer) OfferId() string {
	return po.offerId
}
