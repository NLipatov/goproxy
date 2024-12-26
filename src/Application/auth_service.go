package Application

import "0trace/Domain/ValueObjects"

type AuthService interface {
	Authorize(credentials ValueObjects.Credentials) (bool, error)
}
