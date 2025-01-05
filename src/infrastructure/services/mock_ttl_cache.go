package services

import (
	"errors"
	"goproxy/application"
	"time"
)

type MockTTLCache[T any] struct {
	storage map[string]T
}

func NewMockTTLCache[T any]() application.CacheWithTTL[T] {
	return &MockTTLCache[T]{storage: make(map[string]T)}
}

func (m *MockTTLCache[T]) Get(key string) (T, error) {
	value, exists := m.storage[key]
	if !exists {
		var zero T
		return zero, errors.New("not found")
	}
	return value, nil
}

func (m *MockTTLCache[T]) Set(key string, value T) error {
	m.storage[key] = value
	return nil
}

func (m *MockTTLCache[T]) Expire(key string, ttl time.Duration) error {
	delete(m.storage, key)
	return nil
}
