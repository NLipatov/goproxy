package services

import (
	"errors"
	"time"
)

type MockTTLCache[T any] struct {
	storage map[string]T
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

func (m *MockTTLCache[T]) Expire(key string, _ time.Duration) error {
	delete(m.storage, key)
	return nil
}

func (m *MockTTLCache[T]) Delete(key string) error {
	delete(m.storage, key)
	return nil
}
