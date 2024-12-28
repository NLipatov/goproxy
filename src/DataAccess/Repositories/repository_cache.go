package Repositories

import "goproxy/Domain/Aggregates"

type RepositoryCache[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T) error
	Dispose()
}

type UserRepositoryCache[User Aggregates.User] interface {
	Get(key string) (Aggregates.User, error)
	Set(key string, value Aggregates.User) error
	Dispose()
}
