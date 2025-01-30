package use_cases

import (
	"fmt"
	"goproxy/application/contracts"
	"goproxy/domain/valueobjects"
)

type AuthUseCases struct {
	authService            contracts.AuthService
	userRepository         contracts.UserRepository
	userRestrictionService contracts.UserRestrictionService
}

func NewAuthUseCases(authService contracts.AuthService, userRepository contracts.UserRepository, userRestrictionService contracts.UserRestrictionService) AuthUseCases {
	return AuthUseCases{
		authService:            authService,
		userRepository:         userRepository,
		userRestrictionService: userRestrictionService,
	}
}

func (a *AuthUseCases) AuthorizeBasic(credentials valueobjects.Credentials) (bool, int, error) {
	bCredentials, ok := credentials.(*valueobjects.BasicCredentials)
	if ok {
		user, err := a.userRepository.GetByUsername(bCredentials.Username)
		if err != nil {
			return false, 0, fmt.Errorf("user not found")
		}

		if a.userRestrictionService.IsRestricted(user) {
			return false, 0, fmt.Errorf("user is restricted")
		}

		credentialsValid, err := a.authService.AuthorizeBasic(user, *bCredentials)
		if err != nil {
			return false, 0, err
		}

		return credentialsValid, user.Id(), nil
	}

	return false, 0, fmt.Errorf("invalid credentials")
}
