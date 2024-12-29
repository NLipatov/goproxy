package repositories

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/vmihailenco/msgpack/v5"
	"goproxy/Domain/Aggregates"
	"math"
	"time"
)

type BigCacheUserRepositoryCache struct {
	cache *bigcache.BigCache
}

type StoredUser struct {
	Id           int    `msgpack:"Id"`
	Username     string `msgpack:"Username"`
	PasswordHash []byte `msgpack:"PasswordHash"`
	PasswordSalt []byte `msgpack:"PasswordSalt"`
}

func ToStoredUser(user Aggregates.User) (StoredUser, error) {
	return StoredUser{
		Id:           user.Id(),
		Username:     user.Username(),
		PasswordHash: user.PasswordHash(),
		PasswordSalt: user.PasswordSalt(),
	}, nil
}

func FromStoredUser(user StoredUser) (Aggregates.User, error) {
	return Aggregates.NewUser(user.Id, user.Username, user.PasswordHash, user.PasswordSalt)
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

func (b BigCacheUserRepositoryCache) Get(key string) (Aggregates.User, error) {
	userBytes, err := b.cache.Get(key)
	if err != nil {
		return Aggregates.User{}, err
	}

	var data StoredUser
	err = msgpack.Unmarshal(userBytes, &data)
	user, err := FromStoredUser(data)
	return user, err
}

func (b BigCacheUserRepositoryCache) Set(key string, user Aggregates.User) error {
	data, err := ToStoredUser(user)
	userBytes, err := msgpack.Marshal(data)
	if err != nil {
		return err
	}
	return b.cache.Set(key, userBytes)
}

func (b BigCacheUserRepositoryCache) Dispose() error {
	return b.cache.Close()
}
