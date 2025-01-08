package services

import (
	"github.com/stretchr/testify/mock"
	"time"
)

type MockCacheWithTTL[T any] struct {
	mock.Mock
}

func (m *MockCacheWithTTL[T]) Get(key string) (T, error) {
	args := m.Called(key)
	return args.Get(0).(T), args.Error(1)
}

func (m *MockCacheWithTTL[T]) Set(key string, value T) error {
	args := m.Called(key, value)
	return args.Error(0)
}

func (m *MockCacheWithTTL[T]) Expire(key string, ttl time.Duration) error {
	args := m.Called(key, ttl)
	return args.Error(0)
}
