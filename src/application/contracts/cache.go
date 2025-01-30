package contracts

import "time"

type Cache[T any] interface {
	Get(key string) (T, error)
	Set(key string, value T) error
	Delete(key string) error
}

type CacheWithTTL[T any] interface {
	Cache[T]
	Expire(key string, ttl time.Duration) error
}
