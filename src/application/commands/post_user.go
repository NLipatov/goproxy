package commands

import "goproxy/domain/valueobjects"

type PostUser struct {
	Username string
	Password valueobjects.Password
}
