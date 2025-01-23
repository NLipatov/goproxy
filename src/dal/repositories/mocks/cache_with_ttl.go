package mocks

import (
	"errors"
	"goproxy/application"
	"sync"
	"time"
)

type mockCacheItem[T any] struct {
	value      T
	expiration time.Time
}

type MockCacheWithTTL[T any] struct {
	data  map[string]mockCacheItem[T]
	mutex sync.RWMutex
}

func NewMockCacheWithTTL[T any]() application.CacheWithTTL[T] {
	return &MockCacheWithTTL[T]{
		data: make(map[string]mockCacheItem[T]),
	}
}

func (m *MockCacheWithTTL[T]) Get(key string) (T, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	item, exists := m.data[key]
	if !exists || (item.expiration.Before(time.Now()) && !item.expiration.IsZero()) {
		var zeroValue T
		return zeroValue, errors.New("cache miss")
	}
	return item.value, nil
}

func (m *MockCacheWithTTL[T]) Set(key string, value T) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.data[key] = mockCacheItem[T]{
		value:      value,
		expiration: time.Now().Add(1 * time.Hour),
	}
	return nil
}

func (m *MockCacheWithTTL[T]) Delete(key string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.data[key]; !exists {
		return errors.New("key not found")
	}
	delete(m.data, key)
	return nil
}

func (m *MockCacheWithTTL[T]) Expire(key string, ttl time.Duration) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	item, exists := m.data[key]
	if !exists {
		return errors.New("key not found")
	}

	item.expiration = time.Now().Add(ttl)
	m.data[key] = item
	return nil
}
