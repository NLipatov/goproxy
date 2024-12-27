package Commands

import "goproxy/Domain/ValueObjects"

type PostUserCommand struct {
	Username string
	Password ValueObjects.Password
}
