package services

import (
	"errors"
	"sync"
	"time"
)

type cacheEntry[T any] struct {
	value      T
	expiration time.Time
}

type MapCacheWithTTL[T any] struct {
	data  map[string]cacheEntry[T]
	mutex sync.RWMutex
}

func NewMapCacheWithTTL[T any]() *MapCacheWithTTL[T] {
	return &MapCacheWithTTL[T]{
		data: make(map[string]cacheEntry[T]),
	}
}

func (c *MapCacheWithTTL[T]) Get(key string) (T, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	entry, exists := c.data[key]
	if !exists {
		var zero T
		return zero, errors.New("key not found")
	}

	if !entry.expiration.IsZero() && time.Now().After(entry.expiration) {
		c.mutex.RUnlock()
		c.mutex.Lock()
		delete(c.data, key)
		c.mutex.Unlock()
		c.mutex.RLock()
		var zero T
		return zero, errors.New("key not found (expired)")
	}

	return entry.value, nil
}

func (c *MapCacheWithTTL[T]) Set(key string, value T) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data[key] = cacheEntry[T]{value: value}
	return nil
}

func (c *MapCacheWithTTL[T]) Expire(key string, ttl time.Duration) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	entry, exists := c.data[key]
	if !exists {
		return errors.New("key not found")
	}

	entry.expiration = time.Now().Add(ttl)
	c.data[key] = entry
	return nil
}
