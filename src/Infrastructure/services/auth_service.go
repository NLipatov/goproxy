package services

import (
	"fmt"
	"goproxy/Application"
	"goproxy/Domain/Aggregates"
	"goproxy/Domain/ValueObjects"
)

type AuthService struct {
	cryptoService Application.CryptoService
}

func NewAuthService(cryptoService Application.CryptoService) *AuthService {
	return &AuthService{
		cryptoService: cryptoService,
	}
}

func (authService *AuthService) AuthorizeBasic(user Aggregates.User, credentials ValueObjects.BasicCredentials) (bool, error) {
	isPasswordValid := authService.cryptoService.ValidateHash(user.PasswordHash(), user.PasswordSalt(), credentials.Password)
	if !isPasswordValid {
		return false, fmt.Errorf("invalid credentials")
	}

	return true, nil
}
