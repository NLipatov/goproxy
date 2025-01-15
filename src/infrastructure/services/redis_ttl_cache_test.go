package services

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"os"
	"strings"
	"testing"
	"time"
)

type TestStruct struct {
	Field1 string `json:"field1"`
	Field2 int    `json:"field2"`
}

func setupRedisContainer() (testcontainers.Container, string, error) {
	ctx := context.Background()

	req := testcontainers.ContainerRequest{
		Image:        "valkey/valkey:8.0-bookworm",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForListeningPort("6379/tcp"),
	}
	redisContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, "", err
	}

	mappedPort, err := redisContainer.MappedPort(ctx, "6379")
	if err != nil {
		return nil, "", err
	}

	host, err := redisContainer.Host(ctx)
	if err != nil {
		return nil, "", err
	}

	address := host + ":" + mappedPort.Port()
	return redisContainer, address, nil
}

func TestRedisCache(t *testing.T) {
	// Set up Redis container
	redisContainer, address, err := setupRedisContainer()
	assert.NoError(t, err)
	defer func(redisContainer testcontainers.Container, ctx context.Context) {
		_ = redisContainer.Terminate(ctx)
	}(redisContainer, context.Background())

	host, port := strings.Split(address, ":")[0], strings.Split(address, ":")[1]
	// Set environment variables for RedisCache
	_ = os.Setenv("TC_CACHE_HOST", host)
	defer func() {
		_ = os.Unsetenv("TC_CACHE_HOST")
	}()

	_ = os.Setenv("TC_CACHE_PORT", port)
	defer func() {
		_ = os.Unsetenv("TC_CACHE_PORT")
	}()

	_ = os.Setenv("TC_CACHE_USER", "")
	defer func() {
		_ = os.Unsetenv("TC_CACHE_USER")
	}()

	_ = os.Setenv("TC_CACHE_PASSWORD", "")
	defer func() {
		_ = os.Unsetenv("TC_CACHE_PASSWORD")
	}()

	_ = os.Setenv("TC_CACHE_USER", "default")
	defer func() {
		_ = os.Unsetenv("TC_CACHE_USER")
	}()

	// Initialize RedisCache
	cache, err := NewRedisCache[TestStruct]()
	if err != nil {
		t.Fatal(err)
	}

	// Test Set and Get
	t.Run("Set and Get", func(t *testing.T) {
		key := "testKey"
		value := TestStruct{
			Field1: "value1",
			Field2: 123,
		}

		// Set value in Redis
		err := cache.Set(key, value)
		assert.NoError(t, err, "Failed to set value in Redis")

		// Get value from Redis
		fetchedValue, err := cache.Get(key)
		assert.NoError(t, err, "Failed to get value from Redis")
		assert.Equal(t, value, fetchedValue, "Fetched value does not match the original")
	})

	// Test Expire
	t.Run("Expire", func(t *testing.T) {
		key := "testExpireKey"
		value := TestStruct{
			Field1: "value2",
			Field2: 456,
		}

		// Set value in Redis
		err := cache.Set(key, value)
		assert.NoError(t, err, "Failed to set value in Redis")

		// Set expiration for the key
		err = cache.Expire(key, 1*time.Second)
		assert.NoError(t, err, "Failed to set expiration for Redis key")

		// Wait for expiration
		time.Sleep(2 * time.Second)

		// Attempt to get the expired key
		_, err = cache.Get(key)
		assert.Error(t, err, "Expected error for expired key")
	})

	// Test Delete
	t.Run("Delete", func(t *testing.T) {
		key := "testDeleteKey"
		value := TestStruct{
			Field1: "value3",
			Field2: 789,
		}

		// Set value in Redis
		err := cache.Set(key, value)
		assert.NoError(t, err, "Failed to set value in Redis")

		// Delete the key
		err = cache.Delete(key)
		assert.NoError(t, err, "Failed to delete Redis key")

		time.Sleep(time.Millisecond * 1300)

		// Attempt to get the deleted key
		_, err = cache.Get(key)
		assert.Error(t, err, "Expected error for deleted key")
	})
}
