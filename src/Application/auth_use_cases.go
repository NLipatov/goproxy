package Application

import (
	"fmt"
	"goproxy/Domain/ValueObjects"
)

type AuthUseCases struct {
	authService    AuthService
	userRepository UserRepository
}

func NewAuthUseCases(authService AuthService, userRepository UserRepository) AuthUseCases {
	return AuthUseCases{
		authService:    authService,
		userRepository: userRepository,
	}
}

func (a *AuthUseCases) Authorize(credentials ValueObjects.Credentials) (bool, error) {
	bCredentials, ok := credentials.(*ValueObjects.BasicCredentials)
	if ok {
		user, err := a.userRepository.GetByUsername(bCredentials.Username)
		if err != nil {
			return false, fmt.Errorf("user not found")
		}

		return a.authService.AuthorizeBasic(user, *bCredentials)
	}

	return false, fmt.Errorf("invalid credentials")
}
