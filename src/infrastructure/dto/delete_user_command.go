package dto

import (
	"goproxy/application/commands"
	"goproxy/domain/valueobjects"
)

type DeleteUserCommand struct {
	Id       string
	Password string
}

func (dto *DeleteUserCommand) ToDeleteUserCommandDTO() (commands.DeleteUser, error) {
	password, err := valueobjects.NewPasswordFromString(dto.Password)
	if err != nil {
		return commands.DeleteUser{}, err
	}

	return commands.DeleteUser{
		Id:       dto.Id,
		Password: password,
	}, nil
}
