package dataobjects

type PlanPrice struct {
	id       int
	planId   int
	cents    int64
	currency string
}

func NewPlanPrice(id, planId int, cents int64, currency string) PlanPrice {
	return PlanPrice{
		id:       id,
		planId:   planId,
		cents:    cents,
		currency: currency,
	}
}

func (p *PlanPrice) Id() int {
	return p.id
}

func (p *PlanPrice) PlanId() int {
	return p.planId
}

func (p *PlanPrice) Cents() int64 {
	return p.cents
}

func (p *PlanPrice) Currency() string {
	return p.currency
}
