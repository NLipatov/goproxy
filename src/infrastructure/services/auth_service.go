package services

import (
	"fmt"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
	"os"
	"strconv"
	"time"
)

const defaultValidateCacheTTL = time.Second * 10

type validateResult struct {
	result bool
	err    error
}

type AuthService struct {
	cryptoService    application.CryptoService
	validateCache    application.CacheWithTTL[validateResult]
	validateCacheTTL time.Duration
}

func NewAuthService(cryptoService application.CryptoService) *AuthService {
	validateCacheTTL := defaultValidateCacheTTL
	validateTtlEnv := os.Getenv("AUTH_SERVICE_VALIDATE_TTL_MS")
	if validateTtlEnv != "" {
		ttlMillis, err := strconv.Atoi(validateTtlEnv)
		if err == nil {
			validateCacheTTL = time.Duration(ttlMillis) * time.Millisecond
		}
	}

	return &AuthService{
		cryptoService:    cryptoService,
		validateCache:    NewMapCacheWithTTL[validateResult](),
		validateCacheTTL: validateCacheTTL,
	}
}

func (authService *AuthService) AuthorizeBasic(user aggregates.User, credentials valueobjects.BasicCredentials) (bool, error) {
	cacheKey := fmt.Sprintf("%s:%x", credentials.Username, user.PasswordHash())
	cached, cachedErr := authService.validateCache.Get(cacheKey)
	if cachedErr == nil {
		return cached.result, cached.err
	}

	isPasswordValid := authService.cryptoService.ValidateHash(user.PasswordHash(), credentials.Password)
	if !isPasswordValid {
		return false, fmt.Errorf("invalid credentials")
	}

	_ = authService.validateCache.Set(cacheKey, validateResult{true, nil})
	_ = authService.validateCache.Expire(cacheKey, authService.validateCacheTTL)

	return true, nil
}
