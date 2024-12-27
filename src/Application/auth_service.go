package Application

import (
	"goproxy/Domain/Aggregates"
	"goproxy/Domain/ValueObjects"
)

type AuthService interface {
	AuthorizeBasic(user Aggregates.User, credentials ValueObjects.BasicCredentials) (bool, error)
}
