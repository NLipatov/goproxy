package repositories

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/vmihailenco/msgpack/v5"
	"goproxy/domain/aggregates"
	"math"
	"time"
)

type BigCacheUserRepositoryCache struct {
	cache *bigcache.BigCache
}

type StoredUser struct {
	Id           int    `msgpack:"Id"`
	Username     string `msgpack:"Username"`
	Email        string `msgpack:"Email"`
	PasswordHash string `msgpack:"PasswordHash"`
}

func ToStoredUser(user aggregates.User) (StoredUser, error) {
	return StoredUser{
		Id:           user.Id(),
		Username:     user.Username(),
		Email:        user.Email(),
		PasswordHash: user.PasswordHash(),
	}, nil
}

func FromStoredUser(user StoredUser) (aggregates.User, error) {
	return aggregates.NewUser(user.Id, user.Username, user.Email, user.PasswordHash)
}

func NewBigCacheUserRepositoryCache(ttl time.Duration, cleanInterval time.Duration, shards int, maxEntrySizeBytes int) (BigCacheUserRepositoryCache, error) {
	// bigcache restriction: value must be a power of two
	shards = nextPowerOfTwo(shards)

	cache, err := bigcache.New(context.Background(), bigcache.Config{
		LifeWindow:   ttl,
		CleanWindow:  cleanInterval,
		Shards:       shards,
		MaxEntrySize: maxEntrySizeBytes,
		Verbose:      false,
	})
	if err != nil {
		return BigCacheUserRepositoryCache{}, err
	}

	return BigCacheUserRepositoryCache{cache: cache}, nil
}

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}

func (b BigCacheUserRepositoryCache) Get(key string) (aggregates.User, error) {
	userBytes, err := b.cache.Get(key)
	if err != nil {
		return aggregates.User{}, err
	}

	var data StoredUser
	err = msgpack.Unmarshal(userBytes, &data)
	user, err := FromStoredUser(data)
	return user, err
}

func (b BigCacheUserRepositoryCache) Set(key string, user aggregates.User) error {
	data, err := ToStoredUser(user)
	userBytes, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}
	return b.cache.Set(key, userBytes)
}

func (b BigCacheUserRepositoryCache) Delete(key string) error {
	return b.cache.Delete(key)
}

func (b BigCacheUserRepositoryCache) Dispose() error {
	return b.cache.Close()
}
