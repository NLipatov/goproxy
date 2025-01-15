package dto

import (
	"goproxy/application/commands"
	"goproxy/domain/valueobjects"
)

type PostUserCommand struct {
	Username string
	Email    string
	Password string
}

func (dto *PostUserCommand) ToCreateUserCommand() (commands.PostUser, error) {
	password, err := valueobjects.NewPasswordFromString(dto.Password)
	if err != nil {
		return commands.PostUser{}, err
	}

	return commands.PostUser{
		Username: dto.Username,
		Email:    dto.Email,
		Password: password,
	}, nil
}
