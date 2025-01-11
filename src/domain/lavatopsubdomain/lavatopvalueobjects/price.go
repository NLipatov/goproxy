package lavatopvalueobjects

type Price struct {
	cents       int64
	periodicity Periodicity
	currency    Currency
}

func NewPrice(cents int64, currency Currency, periodicity Periodicity) Price {
	return Price{
		cents:       cents,
		periodicity: periodicity,
		currency:    currency,
	}
}

func (p Price) Cents() int64 {
	return p.cents
}

func (p Price) Periodicity() Periodicity {
	return p.periodicity
}

func (p Price) Currency() Currency {
	return p.currency
}
