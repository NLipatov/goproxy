package DTO

import (
	"goproxy/Application/Commands"
	"goproxy/Domain/ValueObjects"
)

type DeleteUserCommandDTO struct {
	Id       string
	Password string
}

func (dto *DeleteUserCommandDTO) ToDeleteUserCommandDTO() (Commands.DeleteUserCommand, error) {
	password, err := ValueObjects.NewPasswordFromString(dto.Password)
	if err != nil {
		return Commands.DeleteUserCommand{}, err
	}

	return Commands.DeleteUserCommand{
		Id:       dto.Id,
		Password: password,
	}, nil
}
