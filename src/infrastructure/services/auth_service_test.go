package services

import (
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
)

const sampleValidArgon2idHash = "$argon2id$v=19$m=65536,t=3,p=2$c29tZXNhbHQ$RdescudvJCsgt3ub+b+dWRWJTmaaJObG"

type mockCryptoService struct {
	ValidateHashFunc func(fullHash, password string) bool
}

func (m *mockCryptoService) GenerateRandomString(length int) (string, error) {
	randomBytes := make([]byte, length)
	return string(randomBytes), nil
}

func (m *mockCryptoService) ValidateHash(fullHash, password string) bool {
	return m.ValidateHashFunc(fullHash, password)
}

func (m *mockCryptoService) HashValue(string) (string, error) {
	return sampleValidArgon2idHash, nil
}

type mockCache struct {
	data     map[string]validateResult
	ttl      map[string]time.Time
	setCalls int
	getCalls int
}

type noSetMockCache struct {
	data     map[string]validateResult
	ttl      map[string]time.Time
	setCalls int
	getCalls int
}

func newMockCache() *mockCache {
	return &mockCache{
		data: make(map[string]validateResult),
		ttl:  make(map[string]time.Time),
	}
}

func (mc *mockCache) Get(key string) (validateResult, error) {
	mc.getCalls++
	if result, exists := mc.data[key]; exists {
		if time.Now().Before(mc.ttl[key]) {
			return result, nil
		}
		delete(mc.data, key)
	}
	return validateResult{}, errors.New("not found")
}

func (mc *mockCache) Set(key string, value validateResult) error {
	mc.setCalls++
	mc.data[key] = value
	return nil
}

func (mc *mockCache) Expire(key string, ttl time.Duration) error {
	mc.ttl[key] = time.Now().Add(ttl)
	return nil
}

func (mc *noSetMockCache) Get(key string) (validateResult, error) {
	mc.getCalls++
	if result, exists := mc.data[key]; exists {
		if time.Now().Before(mc.ttl[key]) {
			return result, nil
		}
		delete(mc.data, key)
	}
	return validateResult{}, errors.New("not found")
}

func (mc *noSetMockCache) Set(string, validateResult) error {
	return fmt.Errorf("no set mock cache failed to set to cache")
}

func (mc *noSetMockCache) Expire(key string, ttl time.Duration) error {
	mc.ttl[key] = time.Now().Add(ttl)
	return nil
}
func TestAuthorizeBasic_Success(t *testing.T) {
	var ValidateHashFuncCalls int
	cryptoService := &mockCryptoService{
		ValidateHashFunc: func(fullHash, password string) bool {
			ValidateHashFuncCalls++
			return true
		},
	}
	cache := newMockCache()

	username := fmt.Sprintf("test_user_%d", time.Now().UTC().UnixNano())
	user, userErr := aggregates.NewUser(1, username, fmt.Sprintf("%s@example.com", username), sampleValidArgon2idHash)
	if userErr != nil {
		t.Fatal(userErr)
	}

	authService := AuthService{
		cryptoService: cryptoService,
		validateCache: cache,
	}

	credentials := valueobjects.BasicCredentials{
		Username: "test_user",
		Password: "password",
	}

	result, err := authService.AuthorizeBasic(user, credentials)

	assert.NoError(t, err)
	assert.True(t, result)
	assert.Equal(t, 1, ValidateHashFuncCalls)
}

func TestAuthorizeBasic_InvalidCredentials(t *testing.T) {
	var ValidateHashFuncCalls int
	cryptoService := &mockCryptoService{
		ValidateHashFunc: func(fullHash, password string) bool {
			ValidateHashFuncCalls++
			return false
		},
	}
	cache := newMockCache()

	username := fmt.Sprintf("test_user_%d", time.Now().UTC().UnixNano())
	user, userErr := aggregates.NewUser(1, username, fmt.Sprintf("%s@example.com", username), sampleValidArgon2idHash)
	if userErr != nil {
		t.Fatal(userErr)
	}

	authService := AuthService{
		cryptoService: cryptoService,
		validateCache: cache,
	}

	credentials := valueobjects.BasicCredentials{
		Username: "test_user",
		Password: "wrong_password",
	}

	result, err := authService.AuthorizeBasic(user, credentials)

	assert.Error(t, err)
	assert.False(t, result)
	assert.Equal(t, 1, ValidateHashFuncCalls)
}

func TestAuthorizeBasic_CacheUsage(t *testing.T) {
	var ValidateHashFuncCalls int
	cryptoService := &mockCryptoService{
		ValidateHashFunc: func(fullHash, password string) bool {
			ValidateHashFuncCalls++
			return true
		},
	}
	cache := newMockCache()

	username := fmt.Sprintf("test_user_%d", time.Now().UTC().UnixNano())
	user, userErr := aggregates.NewUser(1, username, fmt.Sprintf("%s@example.com", username), sampleValidArgon2idHash)
	if userErr != nil {
		t.Fatal(userErr)
	}

	authService := AuthService{
		cryptoService:    cryptoService,
		validateCache:    cache,
		validateCacheTTL: time.Second,
	}

	credentials := valueobjects.BasicCredentials{
		Username: "test_user",
		Password: "password",
	}

	// First call - no cache. ValidateHash, get and set should be called.
	result, err := authService.AuthorizeBasic(user, credentials)
	assert.NoError(t, err)
	assert.True(t, result)

	// Second call - should use cache. Should not call ValidateHash function, get should be called second time.
	result, err = authService.AuthorizeBasic(user, credentials)
	assert.NoError(t, err)
	assert.True(t, result)

	// Validate that cache was accessed twice
	assert.Equal(t, 1, cache.setCalls)
	assert.Equal(t, 2, cache.getCalls)
	assert.Equal(t, 1, ValidateHashFuncCalls)
}

func TestAuthorizeBasic_CacheExpiry(t *testing.T) {
	var ValidateHashFuncCalls int
	cryptoService := &mockCryptoService{
		ValidateHashFunc: func(fullHash, password string) bool {
			ValidateHashFuncCalls++
			return true
		},
	}
	cache := newMockCache()

	username := fmt.Sprintf("test_user_%d", time.Now().UTC().UnixNano())
	user, userErr := aggregates.NewUser(1, username, fmt.Sprintf("%s@example.com", username), sampleValidArgon2idHash)
	if userErr != nil {
		t.Fatal(userErr)
	}

	validateTtlMs := 50

	authService := AuthService{
		cryptoService:    cryptoService,
		validateCache:    cache,
		validateCacheTTL: time.Duration(validateTtlMs) * time.Millisecond,
	}

	credentials := valueobjects.BasicCredentials{
		Username: "test_user",
		Password: "password",
	}

	// First call - no cache
	result, err := authService.AuthorizeBasic(user, credentials)
	assert.NoError(t, err)
	assert.True(t, result)

	// Simulate cache expiry
	time.Sleep(time.Millisecond*10 + time.Millisecond*time.Duration(validateTtlMs))

	// Second call - cache expired, should call ValidateHash again
	result, err = authService.AuthorizeBasic(user, credentials)
	assert.NoError(t, err)
	assert.True(t, result)

	// Validate call count
	assert.Equal(t, 2, cache.setCalls)
	assert.Equal(t, 2, cache.getCalls)
	assert.Equal(t, 2, ValidateHashFuncCalls)
}

func TestAuthorizeBasic_MinTTL(t *testing.T) {
	var ValidateHashFuncCalls int
	cryptoService := &mockCryptoService{
		ValidateHashFunc: func(fullHash, password string) bool {
			ValidateHashFuncCalls++
			return true
		},
	}
	cache := newMockCache()

	username := fmt.Sprintf("test_user_%d", time.Now().UTC().UnixNano())
	user, userErr := aggregates.NewUser(1, username, fmt.Sprintf("%s@example.com", username), sampleValidArgon2idHash)
	if userErr != nil {
		t.Fatal(userErr)
	}

	validateTtlMs := 1
	_ = os.Setenv("AUTH_SERVICE_VALIDATE_TTL_MS", fmt.Sprintf("%d", validateTtlMs))
	defer func() {
		_ = os.Unsetenv("AUTH_SERVICE_VALIDATE_TTL_MS")
	}()

	authService := AuthService{
		cryptoService: cryptoService,
		validateCache: cache,
	}

	credentials := valueobjects.BasicCredentials{
		Username: "test_user",
		Password: "password",
	}

	// First call - no cache. ValidateHashFunc call, cache get and set get called.
	result, err := authService.AuthorizeBasic(user, credentials)
	assert.NoError(t, err)
	assert.True(t, result)

	// Wait for TTL to expire
	time.Sleep(time.Millisecond * time.Duration(validateTtlMs+1))

	// Second call - cache expired. ValidateHashFunc call, cache get and set get called.
	result, err = authService.AuthorizeBasic(user, credentials)
	assert.NoError(t, err)
	assert.True(t, result)

	// Validate call count
	assert.Equal(t, 2, cache.setCalls)
	assert.Equal(t, 2, cache.getCalls)
	assert.Equal(t, 2, ValidateHashFuncCalls)
}

func TestAuthorizeBasic_CacheError(t *testing.T) {
	cryptoService := &mockCryptoService{
		ValidateHashFunc: func(fullHash, password string) bool {
			return true
		},
	}

	cache := &noSetMockCache{
		data: make(map[string]validateResult),
		ttl:  make(map[string]time.Time),
	}

	username := fmt.Sprintf("test_user_%d", time.Now().UTC().UnixNano())
	user, _ := aggregates.NewUser(1, username, fmt.Sprintf("%s@example.com", username), sampleValidArgon2idHash)

	authService := AuthService{
		cryptoService:    cryptoService,
		validateCache:    cache,
		validateCacheTTL: time.Second,
	}

	credentials := valueobjects.BasicCredentials{
		Username: "test_user",
		Password: "password",
	}

	// Test authorization despite cache error
	result, err := authService.AuthorizeBasic(user, credentials)
	assert.NoError(t, err)
	assert.True(t, result)
}
