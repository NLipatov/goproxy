package config

type BoundedContext string

const (
	UNSET   BoundedContext = "UNSET"
	TRAFFIC BoundedContext = "TRAFFIC"
	PLANS   BoundedContext = "PLANS"
)
