package dto

type PriceDto struct {
	Currency    string  `json:"currency"`
	Amount      float64 `json:"amount"`
	Periodicity string  `json:"periodicity"`
}

type OfferResponse struct {
	ID     string     `json:"id"`
	Name   string     `json:"name"`
	Prices []PriceDto `json:"prices"`
}
type ProductResponse struct {
	ID     string          `json:"id"`
	Title  string          `json:"title"`
	Offers []OfferResponse `json:"offers"`
}
type GetOffersResponse struct {
	Items []ProductResponse `json:"items"`
}
