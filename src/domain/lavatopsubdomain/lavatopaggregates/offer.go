package lavatopaggregates

import "goproxy/domain/lavatopsubdomain/lavatopvalueobjects"

type Offer struct {
	extId string
	name  string
	price lavatopvalueobjects.Price
}

func NewOffer(extId string, name string, price lavatopvalueobjects.Price) Offer {
	return Offer{
		extId: extId,
		name:  name,
		price: price,
	}
}

func (o Offer) ExtId() string {
	return o.extId
}

func (o Offer) Name() string {
	return o.name
}

func (o Offer) Price() lavatopvalueobjects.Price {
	return o.price

}
