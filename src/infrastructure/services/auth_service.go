package services

import (
	"fmt"
	"golang.org/x/sync/singleflight"
	"goproxy/application/contracts"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
	"os"
	"strconv"
	"time"
)

const defaultValidateCacheTTL = time.Hour * 8

type ValidateResult struct {
	result bool
	err    error
}

type AuthService struct {
	cryptoService     contracts.CryptoService
	validateCache     contracts.CacheWithTTL[ValidateResult]
	validateCacheTTL  time.Duration
	singleFlightGroup singleflight.Group
}

func NewAuthService(cryptoService contracts.CryptoService, cache contracts.CacheWithTTL[ValidateResult]) *AuthService {
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

	// Critical section: ValidateHash involves computationally expensive Argon2 logic.
	// Using singleflight ensures that password hash validation for a single user
	// is performed exactly once at a time, reducing redundant processing.
	result, resultErr, _ := a.singleFlightGroup.Do(cacheKey, func() (interface{}, error) {
		cached, cachedErr := a.validateCache.Get(cacheKey)
		if cachedErr == nil {
			return ValidateResult{cached.result, cached.err}, nil
		}

		isPasswordValid := a.cryptoService.ValidateHash(user.PasswordHash(), credentials.Password)
		if !isPasswordValid {
			return ValidateResult{false, fmt.Errorf("invalid credentials")}, nil
		}

		_ = a.validateCache.Set(cacheKey, ValidateResult{true, nil})
		_ = a.validateCache.Expire(cacheKey, a.validateCacheTTL)

		return ValidateResult{true, nil}, nil
	})

	if resultErr != nil {
		return false, resultErr
	}

	validateResult := result.(ValidateResult)
	return validateResult.result, validateResult.err
}
