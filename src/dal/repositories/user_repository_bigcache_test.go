package repositories

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"goproxy/domain/aggregates"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	ttl := 1 * time.Minute
	cleanInterval := 30 * time.Second
	shards := 10
	maxEntrySize := 512

	cache, err := NewBigCacheUserRepositoryCache(ttl, cleanInterval, shards, maxEntrySize)
	assert.NoError(t, err)
	defer func() { _ = cache.Dispose() }()

	username := fmt.Sprintf("test_user_%d", time.Now().UTC().UnixNano())
	email := fmt.Sprintf("%s@example.com", username)
	user, err := aggregates.NewUser(1, username, email, make([]byte, 32), make([]byte, 32))
	assert.NoError(t, err)

	err = cache.Set(username, user)
	assert.NoError(t, err)

	fetchedUser, err := cache.Get(username)
	assert.NoError(t, err)
	assert.Equal(t, user, fetchedUser)
}

func TestTTLExpiration(t *testing.T) {
	ttl := 100 * time.Millisecond
	cleanInterval := 100 * time.Millisecond
	shards := 1024
	maxEntrySize := 500

	cache, err := NewBigCacheUserRepositoryCache(ttl, cleanInterval, shards, maxEntrySize)
	assert.NoError(t, err)
	defer func(cache BigCacheUserRepositoryCache) {
		_ = cache.Dispose()
	}(cache)

	key := "expired_user"
	user, err := aggregates.NewUser(2, "mark_doe", "example@example.com", make([]byte, 32), make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}

	err = cache.Set(key, user)
	assert.NoError(t, err, "failed to set user in cache")

	fetchedUser, err := cache.Get(key)
	assert.NoError(t, err, "failed to fetch user before TTL expiration")
	assert.Equal(t, user, fetchedUser, "fetched user does not match the original before TTL expiration")

	time.Sleep(1 * time.Second)

	_, err = cache.Get(key)
	assert.Error(t, err, "expected error due to TTL expiration")
}

func TestUpdateAndFetch(t *testing.T) {
	ttl := 1 * time.Minute
	cleanInterval := 30 * time.Second
	shards := 1024
	maxEntrySize := 500

	cache, err := NewBigCacheUserRepositoryCache(ttl, cleanInterval, shards, maxEntrySize)
	assert.NoError(t, err)
	defer func(cache BigCacheUserRepositoryCache) {
		_ = cache.Dispose()
	}(cache)

	user, err := aggregates.NewUser(3, "alex_doe", "example@example.com", make([]byte, 32), make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}

	key := "user125"
	err = cache.Set(key, user)
	assert.NoError(t, err)

	updatedUser, err := aggregates.NewUser(3, "seth_doe", "example@example.com", make([]byte, 32), make([]byte, 32))
	if err != nil {
		t.Fatal(err)
	}

	err = cache.Set(key, updatedUser)
	assert.NoError(t, err, "failed to update user in cache")

	fetchedUpdatedUser, err := cache.Get(key)
	assert.NoError(t, err, "failed to fetch updated user from cache")
	assert.Equal(t, updatedUser, fetchedUpdatedUser, "fetched updated user does not match the original")
}

func TestDispose(t *testing.T) {
	ttl := 1 * time.Minute
	cleanInterval := 30 * time.Second
	shards := 1024
	maxEntrySize := 500

	cache, err := NewBigCacheUserRepositoryCache(ttl, cleanInterval, shards, maxEntrySize)
	assert.NoError(t, err)

	err = cache.Dispose()
	assert.NoError(t, err, "failed to dispose the cache")
}
