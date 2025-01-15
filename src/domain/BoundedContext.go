package domain

type BoundedContexts string

const (
	UNSET BoundedContexts = "UNSET"
	PROXY BoundedContexts = "PROXY"
	AUTH  BoundedContexts = "AUTH"
	PLAN  BoundedContexts = "PLAN"
)
