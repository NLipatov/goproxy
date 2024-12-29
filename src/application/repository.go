package application

import "goproxy/domain/aggregates"

type Repository[T any] interface {
	GetById(id int) (T, error)
	Create(entity T) (int, error)
	Update(entity T) error
	Delete(entity T) error
}

type UserRepository interface {
	Repository[aggregates.User]
	GetByUsername(username string) (aggregates.User, error)
}
