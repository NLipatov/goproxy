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

type ValidateResult struct {
	result bool
	err    error
}

type AuthService struct {
	cryptoService    application.CryptoService
	validateCache    application.CacheWithTTL[ValidateResult]
	validateCacheTTL time.Duration
}

func NewAuthService(cryptoService application.CryptoService, cache application.CacheWithTTL[ValidateResult]) *AuthService {
	validateCacheTTL := defaultValidateCacheTTL
	validateTtlEnv := os.Getenv("AUTH_SERVICE_VALIDATE_TTL_MS")
	if validateTtlEnv != "" {
		ttlMillis, err := strconv.Atoi(validateTtlEnv)
		if err == nil {
			validateCacheTTL = time.Duration(ttlMillis) * time.Millisecond
		}
	}

	service := AuthService{
		cryptoService:    cryptoService,
		validateCache:    cache,
		validateCacheTTL: validateCacheTTL,
	}

	return &service
}

func (a *AuthService) AuthorizeBasic(user aggregates.User, credentials valueobjects.BasicCredentials) (bool, error) {
	cacheKey := fmt.Sprintf("%s:%x", credentials.Username, user.PasswordHash())
	cached, cachedErr := a.validateCache.Get(cacheKey)
	if cachedErr == nil {
		return cached.result, cached.err
	}

	isPasswordValid := a.cryptoService.ValidateHash(user.PasswordHash(), credentials.Password)
	if !isPasswordValid {
		return false, fmt.Errorf("invalid credentials")
	}

	_ = a.validateCache.Set(cacheKey, ValidateResult{true, nil})
	_ = a.validateCache.Expire(cacheKey, a.validateCacheTTL)

	return true, nil
}
