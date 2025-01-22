package dataobjects

type PlanLavatopOffer struct {
	planId  int
	offerId string
}

func NewPlanLavatopOffer(planId int, offerId string) *PlanLavatopOffer {
	return &PlanLavatopOffer{planId: planId, offerId: offerId}
}

func (po *PlanLavatopOffer) PlanId() int {
	return po.planId
}

func (po *PlanLavatopOffer) OfferId() string {
	return po.offerId
}
