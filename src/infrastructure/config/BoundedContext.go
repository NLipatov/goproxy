package config

type BoundedContext string

const (
	UNSET BoundedContext = "UNSET"
	PROXY BoundedContext = "PROXY"
	PLAN  BoundedContext = "PLAN"
)
