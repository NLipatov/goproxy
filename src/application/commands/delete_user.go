package commands

import "goproxy/domain/valueobjects"

type DeleteUser struct {
	Id       string
	Password valueobjects.Password
}
