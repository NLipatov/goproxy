package commands

import "goproxy/domain/valueobjects"

type PostUser struct {
	Username string
	Email    string
	Password valueobjects.Password
}
