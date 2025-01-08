package contracts

type Invoice interface {
	GetID() string
	GetStatus() string
}
