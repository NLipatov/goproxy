package services

import (
	"fmt"
	"goproxy/application"
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
)

type AuthService struct {
	cryptoService application.CryptoService
}

func NewAuthService(cryptoService application.CryptoService) *AuthService {
	return &AuthService{
		cryptoService: cryptoService,
	}
}

func (authService *AuthService) AuthorizeBasic(user aggregates.User, credentials valueobjects.BasicCredentials) (bool, error) {
	isPasswordValid := authService.cryptoService.ValidateHash(user.PasswordHash(), user.PasswordSalt(), credentials.Password)
	if !isPasswordValid {
		return false, fmt.Errorf("invalid credentials")
	}

	return true, nil
}
