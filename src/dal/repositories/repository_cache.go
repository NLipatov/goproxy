package repositories

import "goproxy/domain/aggregates"

type RepositoryCache[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T) error
	Dispose()
}

type UserRepositoryCache[User aggregates.User] interface {
	Get(key string) (aggregates.User, error)
	Set(key string, value aggregates.User) error
	Dispose()
}
