package Commands

import "goproxy/Domain/ValueObjects"

type DeleteUserCommand struct {
	Id       string
	Password ValueObjects.Password
}
