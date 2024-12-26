package services

import (
	"0trace/Domain/ValueObjects"
)

type AuthService struct {
}

func NewAuthService() *AuthService {
	return &AuthService{}
}

func (authService *AuthService) Authorize(credentials ValueObjects.Credentials) (bool, error) {
	//ToDo: implement auth check
	if credentials.Type() == ValueObjects.Basic {
		bCredentials, ok := credentials.(*ValueObjects.BasicCredentials)
		if !ok {
			return false, nil
		}
		if bCredentials.Username == "admin" && bCredentials.Password == "admin" {
			return true, nil
		}
	}
	return false, nil
}

func (authService *AuthService) Register(username, password string) error {
	//TODO implement me
	panic("implement me")
}
