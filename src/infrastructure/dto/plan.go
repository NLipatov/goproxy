package dto

type Plan struct {
	Name     string   `json:"name"`
	Limits   Limits   `json:"limits"`
	Features []string `json:"features"`
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
