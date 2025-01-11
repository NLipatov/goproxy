package lavatopvalueobjects

type Offer struct {
	extId  string
	name   string
	prices []Price
}

func NewOffer(extId string, name string, price []Price) Offer {
	return Offer{
		extId:  extId,
		name:   name,
		prices: price,
	}
}

func (o Offer) ExtId() string {
	return o.extId
}

func (o Offer) Name() string {
	return o.name
}

func (o Offer) Prices() []Price {
	return o.prices

}
