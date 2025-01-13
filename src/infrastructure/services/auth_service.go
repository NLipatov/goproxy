package services

import (
	"fmt"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
	"time"
)

type validateResult struct {
	result bool
	err    error
}

type AuthService struct {
	cryptoService application.CryptoService
	validateCache application.CacheWithTTL[validateResult]
}

func NewAuthService(cryptoService application.CryptoService) *AuthService {
	return &AuthService{
		cryptoService: cryptoService,
		validateCache: NewMapCacheWithTTL[validateResult](),
	}
}

func (authService *AuthService) AuthorizeBasic(user aggregates.User, credentials valueobjects.BasicCredentials) (bool, error) {
	cacheKey := fmt.Sprintf("%s:%x", credentials.Username, user.PasswordSalt())
	if cached, cachedErr := authService.validateCache.Get(cacheKey); cachedErr == nil {
		return cached.result, cached.err
	}

	isPasswordValid := authService.cryptoService.ValidateHash(user.PasswordHash(), user.PasswordSalt(), credentials.Password)
	if !isPasswordValid {
		return false, fmt.Errorf("invalid credentials")
	}

	_ = authService.validateCache.Set(cacheKey, validateResult{true, nil})
	_ = authService.validateCache.Expire(cacheKey, time.Second*10)

	return true, nil
}
