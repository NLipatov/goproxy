package dto

type Plan struct {
	Name         string    `json:"name"`
	Limits       Limits    `json:"limits"`
	Features     []Feature `json:"features"`
	DurationDays int       `json:"duration_days"`
	Prices       []Price   `json:"prices"`
}

type Feature struct {
	Feature            string `json:"name"`
	FeatureDescription string `json:"description"`
}

type Price struct {
	Cents    int64  `json:"cents"`
	Currency string `json:"currency"`
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
