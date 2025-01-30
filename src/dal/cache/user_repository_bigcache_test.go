package cache

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"goproxy/domain/aggregates"
	"math"
	"math/rand/v2"
	"reflect"
	"testing"
	"time"
)

func TestNewBigCacheUserRepositoryCache(t *testing.T) {
	cache, cacheErr := NewBigCacheUserRepositoryCache(time.Second, time.Millisecond*300, 2, 512)
	if cacheErr != nil {
		t.Fatalf("failed to instantiate cache: %s", cacheErr)
	}

	defer func(cache BigCacheUserRepositoryCache) {
		_ = cache.Dispose()
	}(cache)

	t.Run("DeepEqual positive", func(t *testing.T) {
		user, userErr := aggregates.NewUser(-1, "username", "email@example.com", "password_hash")
		if userErr != nil {
			t.Fatal(userErr)
		}

		sameUser, sameUserErr := aggregates.NewUser(-1, "username", "email@example.com", "password_hash")
		if sameUserErr != nil {
			t.Fatal(sameUserErr)
		}

		assert.True(t, reflect.DeepEqual(user, sameUser))
	})

	t.Run("DeepEqual negative", func(t *testing.T) {
		user, userErr := aggregates.NewUser(-1, "username", "email@example.com", "password_hash")
		if userErr != nil {
			t.Fatal(userErr)
		}

		diffIdUser, diffIdUserErr := aggregates.NewUser(1, "username", "email@example.com", "password_hash")
		if diffIdUserErr != nil {
			t.Fatal(diffIdUserErr)
		}

		diffUsernameUser, diffUsernameUserErr := aggregates.NewUser(-1, "another_username", "email@example.com", "password_hash")
		if diffUsernameUserErr != nil {
			t.Fatal(diffUsernameUserErr)
		}

		diffEmailUser, diffEmailUserErr := aggregates.NewUser(-1, "username", "another_email@example.com", "password_hash")
		if diffEmailUserErr != nil {
			t.Fatal(diffEmailUser)
		}

		diffHashUser, diffHashUserErr := aggregates.NewUser(-1, "username", "email@example.com", "another_password_hash")
		if diffHashUserErr != nil {
			t.Fatal(diffHashUserErr)
		}

		assert.False(t, reflect.DeepEqual(user, diffIdUser))
		assert.False(t, reflect.DeepEqual(user, diffUsernameUser))
		assert.False(t, reflect.DeepEqual(user, diffEmailUser))
		assert.False(t, reflect.DeepEqual(user, diffHashUser))
	})

	t.Run("Set and Set", func(t *testing.T) {
		username := fmt.Sprintf("user_%d", rand.IntN(math.MaxInt32))
		email := fmt.Sprintf("%s@example.com", username)
		passwordHash := fmt.Sprintf("%s_password_hash", username)
		user, userErr := aggregates.NewUser(1, username, email, passwordHash)
		if userErr != nil {
			t.Fatalf("could not instantiate user: %s", userErr)
		}

		setErr := cache.Set(username, user)
		if setErr != nil {
			t.Fatal(setErr)
		}

		cacheUser, getErr := cache.Get(username)
		if getErr != nil {
			t.Fatal(getErr)
		}

		assert.True(t, reflect.DeepEqual(user, cacheUser))
	})

	t.Run("Delete", func(t *testing.T) {
		username := fmt.Sprintf("user_%d", rand.IntN(math.MaxInt32))
		email := fmt.Sprintf("%s@example.com", username)
		passwordHash := fmt.Sprintf("%s_password_hash", username)
		user, userErr := aggregates.NewUser(1, username, email, passwordHash)
		if userErr != nil {
			t.Fatalf("could not instantiate user: %s", userErr)
		}

		setErr := cache.Set(username, user)
		if setErr != nil {
			t.Fatal(setErr)
		}

		cacheUser, getErr := cache.Get(username)
		if getErr != nil {
			t.Fatal(getErr)
		}

		deleteErr := cache.Delete(username)

		_, deletedUserErr := cache.Get(username)

		assert.True(t, reflect.DeepEqual(user, cacheUser))
		assert.Nil(t, deleteErr)
		assert.NotNil(t, deletedUserErr)
	})
}
