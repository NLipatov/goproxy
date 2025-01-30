package cache

import (
	"context"
	"github.com/allegro/bigcache/v3"
	"github.com/vmihailenco/msgpack/v5"
	"goproxy/dal/cache_serialization"
	"goproxy/domain/aggregates"
	"math"
	"time"
)

type BigCacheUserRepositoryCache struct {
	cache      *bigcache.BigCache
	serializer cache_serialization.CacheSerializer[aggregates.User, cache_serialization.UserDto]
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

	return BigCacheUserRepositoryCache{
		cache:      cache,
		serializer: cache_serialization.NewUserSerializer(),
	}, nil
}

func nextPowerOfTwo(n int) int {
	if n <= 1 {
		return 1
	}
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}

func (b BigCacheUserRepositoryCache) Get(key string) (aggregates.User, error) {
	dtoBytes, err := b.cache.Get(key)
	if err != nil {
		return aggregates.User{}, err
	}

	var dto cache_serialization.UserDto
	err = msgpack.Unmarshal(dtoBytes, &dto)
	user := b.serializer.ToT(dto)
	return user, err
}

func (b BigCacheUserRepositoryCache) Set(key string, user aggregates.User) error {
	dto := b.serializer.ToD(user)
	dtoBytes, err := msgpack.Marshal(dto)
	if err != nil {
		return err
	}
	return b.cache.Set(key, dtoBytes)
}

func (b BigCacheUserRepositoryCache) Delete(key string) error {
	return b.cache.Delete(key)
}

func (b BigCacheUserRepositoryCache) Dispose() error {
	return b.cache.Close()
}
