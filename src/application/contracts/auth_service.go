package contracts

import (
	"goproxy/domain/aggregates"
	"goproxy/domain/valueobjects"
)

type AuthService interface {
	AuthorizeBasic(user aggregates.User, credentials valueobjects.BasicCredentials) (bool, error)
}
