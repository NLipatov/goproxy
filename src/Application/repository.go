package Application

import "goproxy/Domain/Aggregates"

type Repository[T any] interface {
	GetById(id int) (T, error)
	Create(entity T) (int, error)
	Update(entity T) error
	Delete(entity T) error
}

type UserRepository interface {
	Repository[Aggregates.User]
	GetByUsername(username string) (Aggregates.User, error)
}
