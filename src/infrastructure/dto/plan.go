package dto

type PlanDto struct {
	Name         string   `json:"name"`
	Features     []string `json:"features"`
	DurationDays int      `json:"duration_days"`
}

type Plan struct {
	Name         string    `json:"name"`
	Limits       Limits    `json:"limits"`
	Features     []Feature `json:"features"`
	DurationDays int       `json:"duration_days"`
	Offers       []Offer   `json:"offers"`
}

type Offer struct {
	Description string  `json:"description"`
	OfferId     string  `json:"offer_id"`
	Prices      []Price `json:"prices"`
}

type Feature struct {
	Feature            string `json:"name"`
	FeatureDescription string `json:"description"`
}

type Price struct {
	Cents          int64    `json:"cents"`
	Currency       string   `json:"currency"`
	PaymentMethods []string `json:"payment_method"`
}

type Limits struct {
	Bandwidth   BandwidthLimit  `json:"bandwidth"`
	Connections ConnectionLimit `json:"connections"`
	Speed       SpeedLimit      `json:"speed"`
}

type BandwidthLimit struct {
	IsLimited bool  `json:"isLimited"`
	Used      int64 `json:"used"`
	Total     int64 `json:"total"`
}

type ConnectionLimit struct {
	IsLimited                bool `json:"isLimited"`
	MaxConcurrentConnections int  `json:"maxConcurrentConnections"`
}

type SpeedLimit struct {
	IsLimited         bool  `json:"isLimited"`
	MaxBytesPerSecond int64 `json:"maxBytesPerSecond"`
}
