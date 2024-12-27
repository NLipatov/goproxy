package DTO

import (
	"goproxy/Application/Commands"
	"goproxy/Domain/ValueObjects"
)

type PostUserCommandDTO struct {
	Username string
	Password string
}

func (dto *PostUserCommandDTO) ToCreateUserCommand() (Commands.PostUserCommand, error) {
	password, err := ValueObjects.NewPasswordFromString(dto.Password)
	if err != nil {
		return Commands.PostUserCommand{}, err
	}

	return Commands.PostUserCommand{
		Username: dto.Username,
		Password: password,
	}, nil
}
